package standalone

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"math/rand"
	"strings"

	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	"github.com/lucasepe/codename"
	"github.com/vshn/appcat-service-postgresql/apis/conditions"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/helmvalues"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var namegeneratorRNG *rand.Rand

func init() {
	rng, err := codename.DefaultRNG()
	if err != nil {
		panic(err)
	}
	namegeneratorRNG = rng
}

// CreateStandalonePipeline is a pipeline that creates a new instance in the target deployment namespace.
// Currently, it's optimized for first-time creation scenarios and may fail when reconciling existing instances.
type CreateStandalonePipeline struct {
	operatorNamespace string
	client            client.Client

	// TODO: Idea: Maybe store and retrieve the following fields from the context for safe access. This would require some convenience getters and setters though.

	instance            *v1alpha1.PostgresqlStandalone
	config              *v1alpha1.PostgresqlStandaloneOperatorConfig
	helmValues          helmvalues.V
	helmChart           *v1alpha1.ChartMeta
	deploymentNamespace *corev1.Namespace
	helmRelease         *helmv1beta1.Release

	connectionSecret *corev1.Secret

	k8upschedule   *k8upv1.Schedule
	s3BucketSecret *corev1.Secret
}

// NewCreateStandalonePipeline creates a new pipeline with the required dependencies.
func NewCreateStandalonePipeline(client client.Client, instance *v1alpha1.PostgresqlStandalone, operatorNamespace string) *CreateStandalonePipeline {
	return &CreateStandalonePipeline{
		instance:          instance,
		client:            client,
		operatorNamespace: operatorNamespace,
	}
}

// RunFirstPass executes the pipeline with configured business logic steps.
// This should only be executed once per pipeline as it stores intermediate results in the struct.
func (p *CreateStandalonePipeline) RunFirstPass(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("fetch operator config", p.fetchOperatorConfig),

			pipeline.NewPipeline().WithNestedSteps("compile helm values",
				pipeline.NewStepFromFunc("read template values", p.useTemplateValues),
				pipeline.NewStepFromFunc("override template values", p.overrideTemplateValues),
				pipeline.NewStepFromFunc("apply values from instance", p.applyValuesFromInstance),
			),

			pipeline.NewStepFromFunc("add finalizer", addFinalizerFn(p.instance, finalizer)),
			pipeline.NewStepFromFunc("set creating condition", setConditionFn(p.instance, &p.instance.Status.Conditions, conditions.Creating())),

			pipeline.NewPipeline().WithNestedSteps("deploy resources",
				pipeline.NewStepFromFunc("ensure deployment namespace", p.ensureDeploymentNamespace),
				pipeline.NewStepFromFunc("ensure credentials secret", p.ensureCredentialsSecret),
				pipeline.NewStepFromFunc("ensure helm release", p.ensureHelmRelease),
				pipeline.If(pipeline.Bool(p.instance.Spec.Backup.Enabled),
					pipeline.NewPipeline().WithNestedSteps("ensure backup",
						pipeline.NewStepFromFunc("ensure encryption secret", p.ensureResticRepositorySecret),
						// TODO: add step to provision S3 bucket
					),
				),
			),

			pipeline.NewStepFromFunc("enrich status with chart meta", p.enrichStatus),
		).
		RunWithContext(ctx).Err()
}

// RunSecondPass runs a pipeline that verifies if all dependent resources are ready.
// It will add the conditions.TypeReady condition to the status field (and update it) if it's considered ready.
// No error is returned in case the instance is not considered ready.
func (p *CreateStandalonePipeline) RunSecondPass(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("fetch operator config", p.fetchOperatorConfig),
			pipeline.NewStepFromFunc("fetch helm release", p.fetchHelmRelease),
			pipeline.If(p.isBackupEnabledPredicate(),
				pipeline.NewPipeline().WithNestedSteps("ensure backup",
					pipeline.NewStepFromFunc("fetch bucket secret", p.fetchS3BucketSecret),
					pipeline.NewStepFromFunc("ensure k8up schedule", p.ensureK8upSchedule),
				),
			),
			pipeline.If(pipeline.Not(p.isBackupEnabledPredicate()), pipeline.NewStepFromFunc("delete k8up schedule", p.ensureK8upSchedule)),
			pipeline.If(p.isHelmReleaseReady,
				pipeline.NewPipeline().WithNestedSteps("finish creation",
					pipeline.NewPipeline().WithNestedSteps("create connection secret",
						pipeline.NewStepFromFunc("fetch credentials", p.fetchCredentialSecret),
						pipeline.NewStepFromFunc("fetch service", p.fetchService),
						pipeline.NewStepFromFunc("set owner reference to connection secret", p.setOwnerReferenceInConnectionSecret),
						pipeline.NewStepFromFunc("ensure connection secret", p.ensureConnectionSecret),
					),
					pipeline.NewStepFromFunc("mark instance ready", p.markInstanceAsReady),
				),
			),
		).
		RunWithContext(ctx).Err()
}

func (p *CreateStandalonePipeline) isBackupEnabledPredicate() pipeline.Predicate {
	return pipeline.BoolPtr(&p.instance.Spec.Backup.Enabled)
}

// fetchOperatorConfig fetches a matching v1alpha1.PostgresqlStandaloneOperatorConfig from the OperatorNamespace.
// The Major version specified in v1alpha1.PostgresqlStandalone is used to filter the correct config by the v1alpha1.PostgresqlMajorVersionLabelKey label.
// If there is none or multiple found, it returns an error.
func (p *CreateStandalonePipeline) fetchOperatorConfig(ctx context.Context) error {
	list := &v1alpha1.PostgresqlStandaloneOperatorConfigList{}
	labels := client.MatchingLabels{
		v1alpha1.PostgresqlMajorVersionLabelKey: p.instance.Spec.Parameters.MajorVersion.String(),
	}
	ns := client.InNamespace(p.operatorNamespace)
	err := p.client.List(ctx, list, labels, ns)
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		return fmt.Errorf("no %s found with label '%s' in namespace '%s'", v1alpha1.PostgresqlStandaloneOperatorConfigKind, labels, ns)
	}
	if len(list.Items) > 1 {
		return fmt.Errorf("multiple versions of %s found with label '%s' in namespace '%s'", v1alpha1.PostgresqlStandaloneOperatorConfigKind, labels, ns)
	}
	p.config = &list.Items[0]
	return nil
}

// useTemplateValues copies the Helm values and Chart metadata from the v1alpha1.PostgresqlStandaloneOperatorConfig spec as the starting parameters.
//
// This step assumes that the config has been fetched first via fetchOperatorConfig.
func (p *CreateStandalonePipeline) useTemplateValues(_ context.Context) error {
	values := helmvalues.V{}
	err := helmvalues.Unmarshal(p.config.Spec.HelmReleaseTemplate.Values, &values)
	p.helmValues = values
	p.helmChart = &p.config.Spec.HelmReleaseTemplate.Chart
	return err
}

// overrideTemplateValues searches for a specific HelmRelease spec that matches the Chart version from the template spec.
// If it does, the template values are replaced or merged.
//
// This step assumes that the config has been fetched first via fetchOperatorConfig.
func (p *CreateStandalonePipeline) overrideTemplateValues(_ context.Context) error {
	for _, release := range p.config.Spec.HelmReleases {
		// TODO: maybe a better semver comparison later on?
		if release.Chart.Version == p.config.Spec.HelmReleaseTemplate.Chart.Version {
			overrides := helmvalues.V{}
			err := helmvalues.Unmarshal(release.Values, &overrides)
			if err != nil {
				return err
			}
			if release.MergeValuesFromTemplate {
				helmvalues.Merge(overrides, &p.helmValues)
			} else {
				p.helmValues = overrides
			}
			if release.Chart.Name != "" {
				p.helmChart.Name = release.Chart.Name
			}
			if release.Chart.Repository != "" {
				p.helmChart.Repository = release.Chart.Repository
			}
		}
	}
	return nil
}

// applyValuesFromInstance merges the user-defined and -exposed Helm values into the current Helm values map.
func (p *CreateStandalonePipeline) applyValuesFromInstance(_ context.Context) error {
	resources := helmvalues.V{
		"auth": helmvalues.V{
			"enablePostgresUser": p.instance.Spec.Parameters.EnableSuperUser,
			"existingSecret":     getCredentialSecretName(),
			"database":           p.instance.Name,
			"username":           p.instance.Name,
		},
		"primary": helmvalues.V{
			"resources": helmvalues.V{
				"limits": helmvalues.V{
					"memory": p.instance.Spec.Parameters.Resources.MemoryLimit.String(),
				},
			},
			"persistence": helmvalues.V{
				"size": p.instance.Spec.Parameters.Resources.StorageCapacity.String(),
			},
		},
		"fullnameOverride": getDeploymentName(),
		"networkPolicy": helmvalues.V{
			"enabled": true,
			"ingressRules": helmvalues.V{
				"primaryAccessOnlyFrom": helmvalues.V{
					"enabled": true,
					"namespaceSelector": helmvalues.V{
						"kubernetes.io/metadata.name": p.instance.Namespace,
					},
				},
			},
		},
	}
	helmvalues.Merge(resources, &p.helmValues)
	return nil
}

// ensureDeploymentNamespace creates the deployment namespace where the Helm release is ultimately deployed in.
func (p *CreateStandalonePipeline) ensureDeploymentNamespace(ctx context.Context) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: generateClusterScopedNameForInstance(),
			// TODO: Add APPUiO cloud organization label that identifies ownership.
			Labels: getCommonLabels(p.instance.Name),
		},
	}
	ns.Labels["app.kubernetes.io/instance-namespace"] = p.instance.Namespace
	p.deploymentNamespace = ns
	return Upsert(ctx, p.client, ns)
}

// ensureCredentialsSecret creates the secret that contains the PostgreSQL secret.
// Passwords are generated, so this step should only run once in the lifetime of the v1alpha1.PostgresqlStandalone instance.
//
// This step assumes that the deployment namespace already exists using ensureDeploymentNamespace.
func (p *CreateStandalonePipeline) ensureCredentialsSecret(ctx context.Context) error {
	// https://github.com/bitnami/charts/tree/master/bitnami/postgresql
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getCredentialSecretName(),
			Namespace: p.deploymentNamespace.Name,
			Labels:    getCommonLabels(p.instance.Name),
		},
		StringData: map[string]string{
			"postgres-password":    generatePassword(),
			"password":             generatePassword(),
			"replication-password": generatePassword(),
		},
	}
	return Upsert(ctx, p.client, secret)
}

// ensureHelmRelease creates the Helm release object.
// It uses the current Helm values that are prepared using useTemplateValues and applyValuesFromInstance.
//
// This step requires that provider-helm from Crossplane is running on the cluster (https://github.com/crossplane-contrib/provider-helm).
func (p *CreateStandalonePipeline) ensureHelmRelease(ctx context.Context) error {
	helmValues, err := helmvalues.Marshal(p.helmValues)
	if err != nil {
		return err
	}
	p.helmRelease = &helmv1beta1.Release{
		ObjectMeta: metav1.ObjectMeta{
			Name:   p.deploymentNamespace.Name,
			Labels: getCommonLabels(p.instance.Name),
		},
		Spec: helmv1beta1.ReleaseSpec{
			ForProvider: helmv1beta1.ReleaseParameters{
				Chart:               helmv1beta1.ChartSpec{Repository: p.helmChart.Repository, Name: p.helmChart.Name, Version: p.helmChart.Version},
				Namespace:           p.deploymentNamespace.Name,
				SkipCreateNamespace: true,
				SkipCRDs:            true,
				Wait:                true,
				ValuesSpec:          helmv1beta1.ValuesSpec{Values: helmValues},
			},
			ResourceSpec: crossplanev1.ResourceSpec{
				ProviderConfigReference: &crossplanev1.Reference{Name: p.config.Spec.HelmProviderConfigReference},
			},
		},
	}
	return Upsert(ctx, p.client, p.helmRelease)
}

// ensureK8upSchedule creates the K8up schedule object.
func (p *CreateStandalonePipeline) ensureK8upSchedule(ctx context.Context) error {
	schedule := &k8upv1.Schedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: p.instance.Status.HelmChart.DeploymentNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, p.client, schedule, func() error {
		schedule.Spec = k8upv1.ScheduleSpec{
			Backup: &k8upv1.BackupSchedule{
				BackupSpec:     k8upv1.BackupSpec{},
				ScheduleCommon: &k8upv1.ScheduleCommon{Schedule: "@daily-random"},
			},
			Archive: &k8upv1.ArchiveSchedule{
				ScheduleCommon: &k8upv1.ScheduleCommon{Schedule: "@weekly-random"},
			},
			Check: &k8upv1.CheckSchedule{
				ScheduleCommon: &k8upv1.ScheduleCommon{Schedule: "@weekly-random"},
			},
			Prune: &k8upv1.PruneSchedule{
				ScheduleCommon: &k8upv1.ScheduleCommon{Schedule: "@weekly-random"},
			},
			Backend: &k8upv1.Backend{
				RepoPasswordSecretRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: getResticRepositorySecretName()},
					Key:                  "repository",
				},
				S3: &k8upv1.S3Spec{
					Endpoint:                 string(p.s3BucketSecret.Data[p.config.Spec.BackupConfigSpec.S3BucketSecret.EndpointRef.Key]),
					Bucket:                   string(p.s3BucketSecret.Data[p.config.Spec.BackupConfigSpec.S3BucketSecret.BucketRef.Key]),
					AccessKeyIDSecretRef:     &p.config.Spec.BackupConfigSpec.S3BucketSecret.AccessKeyRef,
					SecretAccessKeySecretRef: &p.config.Spec.BackupConfigSpec.S3BucketSecret.SecretKeyRef,
				},
			},
			FailedJobsHistoryLimit:     pointer.Int(2),
			SuccessfulJobsHistoryLimit: pointer.Int(2),
		}
		return nil
	})
	return err
}

func (p *CreateStandalonePipeline) deleteK8upSchedule(ctx context.Context) error {
	if p.instance.Status.HelmChart == nil || p.instance.Status.HelmChart.DeploymentNamespace == "" {
		// deployment namespace is unknown, skip
		return nil
	}
	schedule := newK8upSchedule(p.instance)
	err := p.client.Delete(ctx, schedule)
	return client.IgnoreNotFound(err)
}

func (p *CreateStandalonePipeline) ensureResticRepositorySecret(ctx context.Context) error {
	// NOTE: we should not delete the Restic Repository secret.
	// There could be cases where the user temporarily disables backups and then re-enables.
	// This case shouldn't result in a new encryption password that renders the previously created backups unusable.
	// All this is under assumption that the Bucket is not immediately removed when backups are disabled.
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getResticRepositorySecretName(),
			Namespace: p.deploymentNamespace.Name,
		},
	}
	secretKey := "repository"
	_, err := controllerutil.CreateOrUpdate(ctx, getClientFromContext(ctx), secret, func() error {
		secret.Labels = getCommonLabels(getInstanceFromContext(ctx).Name)
		if val, exists := secret.Data[secretKey]; exists && len(val) > 0 {
			// password already set, let's reset every other key
			secret.Data = map[string][]byte{
				secretKey: val,
			}
			return nil
		}
		password := generatePassword()
		if val, exists := secret.StringData[secretKey]; exists && val != "" {
			// only generate a password on new object, not existing ones
			password = val
		}
		secret.StringData = map[string]string{
			secretKey: password,
		}
		return nil
	})
	return err
}

// fetchS3BucketSecret fetches a secret that contains the bucket configuration.
// It assumes that there is another provisioner that deploys S3 bucket ready for use.
func (p *CreateStandalonePipeline) fetchS3BucketSecret(ctx context.Context) error {
	p.s3BucketSecret = &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      p.config.Spec.BackupConfigSpec.S3BucketSecret.BucketRef.Name,
		Namespace: p.instance.Status.HelmChart.DeploymentNamespace,
	}}
	err := p.client.Get(ctx, client.ObjectKeyFromObject(p.s3BucketSecret), p.s3BucketSecret)
	return err
}

func (p *CreateStandalonePipeline) enrichStatus(ctx context.Context) error {
	p.instance.Status.HelmChart = &v1alpha1.ChartMetaStatus{
		ChartMeta:           *p.helmChart,
		DeploymentNamespace: p.deploymentNamespace.Name,
	}
	p.instance.Status.DeploymentStrategy = v1alpha1.StrategyHelmChart
	p.instance.Status.SetObservedGeneration(p.instance)
	err := p.client.Status().Update(ctx, p.instance)
	return err
}

// fetchHelmRelease fetches the Helm release for the given instance.
func (p *CreateStandalonePipeline) fetchHelmRelease(ctx context.Context) error {
	helmRelease := &helmv1beta1.Release{}
	err := p.client.Get(ctx, client.ObjectKey{Name: p.instance.Status.HelmChart.DeploymentNamespace}, helmRelease)
	p.helmRelease = helmRelease
	return err
}

// isHelmReleaseReady returns true if the ModifiedTime is non-zero.
//
// Note: This only works for first-time deployments. In the future another mechanism might be better.
// This step requires that fetchHelmRelease has run before.
func (p *CreateStandalonePipeline) isHelmReleaseReady(_ context.Context) bool {
	if p.instance.Status.HelmChart != nil && !p.instance.Status.HelmChart.ModifiedTime.IsZero() {
		return true
	}
	if p.helmRelease.Status.Synced {
		if readyCondition := FindCrossplaneCondition(p.helmRelease.Status.Conditions, crossplanev1.TypeReady); readyCondition != nil && readyCondition.Status == corev1.ConditionTrue {
			p.instance.Status.HelmChart.ModifiedTime = readyCondition.LastTransitionTime
			return true
		}
	}
	return false
}

// markInstanceAsReady marks an instance immediately as ready by updating the status conditions.
func (p *CreateStandalonePipeline) markInstanceAsReady(ctx context.Context) error {
	meta.SetStatusCondition(&p.instance.Status.Conditions, conditions.Builder().With(conditions.Ready()).WithGeneration(p.instance).Build())
	meta.RemoveStatusCondition(&p.instance.Status.Conditions, conditions.TypeCreating)
	return p.client.Status().Update(ctx, p.instance)
}

// ensureConnectionSecret creates the connection secret in the instance's namespace.
func (p *CreateStandalonePipeline) ensureConnectionSecret(ctx context.Context) error {
	err := Upsert(ctx, getClientFromContext(ctx), getObjectFromContext(ctx, connectionSecretKey{}, &corev1.Secret{}))
	return err
}

// fetchCredentialSecret gets the credential secret and puts the credentials for postgresql into the connection secret.
func (p *CreateStandalonePipeline) fetchCredentialSecret(ctx context.Context) error {
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      getCredentialSecretName(),
		Namespace: p.instance.Status.HelmChart.DeploymentNamespace,
	}}
	err := getClientFromContext(ctx).Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, secret)
	if err != nil {
		return err
	}
	setObjectInContext(ctx, connectionSecretKey{}, secret)
	if p.instance.Spec.Parameters.EnableSuperUser {
		p.addDataToConnectionSecret("POSTGRESQL_POSTGRES_PASSWORD", secret.Data["postgres-password"])
	}
	p.addDataToConnectionSecret("POSTGRESQL_PASSWORD", secret.Data["password"])
	p.addStringDataToConnectionSecret("POSTGRESQL_DATABASE", p.instance.Name)
	p.addStringDataToConnectionSecret("POSTGRESQL_USER", p.instance.Name)
	return err
}

// fetchService gets the service object and puts the DNS records into the connection secret.
func (p *CreateStandalonePipeline) fetchService(ctx context.Context) error {
	service := &corev1.Service{}
	err := getClientFromContext(ctx).Get(ctx, client.ObjectKey{Name: "postgresql", Namespace: p.instance.Status.HelmChart.DeploymentNamespace}, service)
	if err != nil {
		return err
	}
	p.addStringDataToConnectionSecret("POSTGRESQL_SERVICE_NAME", fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace))
	p.addStringDataToConnectionSecret("POSTGRESQL_SERVICE_URL", fmt.Sprintf("postgresql://%s.%s.svc.cluster.local:%d", service.Name, service.Namespace, service.Spec.Ports[0].Port))
	p.addStringDataToConnectionSecret("POSTGRESQL_SERVICE_PORT", fmt.Sprintf("%d", service.Spec.Ports[0].Port))
	return nil
}

func (p *CreateStandalonePipeline) addDataToConnectionSecret(key string, data []byte) {
	if p.connectionSecret == nil {
		p.connectionSecret = p.newConnectionSecret()
	}
	p.connectionSecret.Data[key] = data
}

func (p *CreateStandalonePipeline) addStringDataToConnectionSecret(key, data string) {
	if p.connectionSecret == nil {
		p.connectionSecret = p.newConnectionSecret()
	}
	p.connectionSecret.StringData[key] = data
}

func (p *CreateStandalonePipeline) newConnectionSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.instance.Spec.WriteConnectionSecretToRef.Name,
			Namespace: p.instance.Namespace,
			Labels:    getCommonLabels(p.instance.Name),
		},
		StringData: map[string]string{},
		Data:       map[string][]byte{},
	}
}

func (p *CreateStandalonePipeline) setOwnerReferenceInConnectionSecret(ctx context.Context) error {
	return controllerutil.SetOwnerReference(p.instance, p.connectionSecret, getClientFromContext(ctx).Scheme())
}

func getCredentialSecretName() string {
	return fmt.Sprintf("%s-credentials", getDeploymentName())
}
func getResticRepositorySecretName() string {
	return fmt.Sprintf("%s-restic", getDeploymentName())
}
func getDeploymentName() string {
	return "postgresql"
}

func getCommonLabels(instanceName string) labels.Set {
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
	return labels.Set{
		"app.kubernetes.io/instance":   instanceName,
		"app.kubernetes.io/managed-by": v1alpha1.Group,
		"app.kubernetes.io/created-by": fmt.Sprintf("controller-%s", strings.ToLower(v1alpha1.PostgresqlStandaloneKind)),
	}
}

func generateClusterScopedNameForInstance() string {
	name := ""
	for i := 0; i < 10; i++ {
		name = fmt.Sprintf("%s%s", ServiceNamespacePrefix, codename.Generate(namegeneratorRNG, 4))
		if len(name) <= 63 && !strings.HasSuffix(name, "-") {
			return name
		}
	}
	panic("no generated name under 63 chars without '-' suffix after 10 attempts, giving up")
}

func generatePassword() string {
	return utilrand.String(40)
}

func newK8upSchedule(instance *v1alpha1.PostgresqlStandalone) *k8upv1.Schedule {
	return &k8upv1.Schedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: instance.Status.HelmChart.DeploymentNamespace,
		},
	}
}
