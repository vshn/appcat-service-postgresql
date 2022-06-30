package standalone

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"math/rand"
	"strings"

	pipeline "github.com/ccremer/go-command-pipeline"
	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	"github.com/lucasepe/codename"
	"github.com/vshn/appcat-service-postgresql/apis/conditions"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
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

	connectionSecret *corev1.Secret

	k8upSchedule   *k8upv1.Schedule
	s3BucketSecret *corev1.Secret
}

// NewCreateStandalonePipeline creates a new pipeline with the required dependencies.
func NewCreateStandalonePipeline(operatorNamespace string) *CreateStandalonePipeline {
	return &CreateStandalonePipeline{
		operatorNamespace: operatorNamespace,
	}
}

// RunFirstPass executes the pipeline with configured business logic steps.
// This should only be executed once per pipeline as it stores intermediate results in the struct.
func (p *CreateStandalonePipeline) RunFirstPass(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("fetch operator config", fetchOperatorConfigF(p.operatorNamespace)),

			pipeline.NewPipeline().WithNestedSteps("compile helm values",
				pipeline.NewStepFromFunc("read template values", useTemplateValues),
				pipeline.NewStepFromFunc("override template values", overrideTemplateValues),
				pipeline.NewStepFromFunc("apply values from instance", applyValuesFromInstance),
			),

			pipeline.NewStepFromFunc("add finalizer", addFinalizerFn(getInstanceFromContext(ctx), finalizer)),
			pipeline.NewStepFromFunc("set creating condition", setConditionFn(getInstanceFromContext(ctx), &getInstanceFromContext(ctx).Status.Conditions, conditions.Creating())),

			pipeline.NewPipeline().WithNestedSteps("deploy resources",
				pipeline.NewStepFromFunc("ensure deployment namespace", p.ensureDeploymentNamespace),
				pipeline.NewStepFromFunc("enrich status with chart meta", p.enrichStatus),
				pipeline.NewStepFromFunc("ensure PVC", ensurePVC),
				pipeline.NewStepFromFunc("ensure credentials secret", p.ensureCredentialsSecret),
				pipeline.NewStepFromFunc("ensure helm release", ensureHelmRelease),
				pipeline.If(pipeline.Bool(getInstanceFromContext(ctx).Spec.Backup.Enabled),
					pipeline.NewPipeline().WithNestedSteps("ensure backup",
						pipeline.NewStepFromFunc("ensure encryption secret", p.ensureResticRepositorySecret),
						// TODO: add step to provision S3 bucket
					),
				),
			),
		).
		RunWithContext(ctx).Err()
}

// RunSecondPass runs a pipeline that verifies if all dependent resources are ready.
// It will add the conditions.TypeReady condition to the status field (and update it) if it's considered ready.
// No error is returned in case the instance is not considered ready.
func (p *CreateStandalonePipeline) RunSecondPass(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("fetch operator config", fetchOperatorConfigF(p.operatorNamespace)),
			pipeline.NewStepFromFunc("fetch helm release", fetchHelmRelease),
			pipeline.If(p.isBackupEnabledPredicate(ctx),
				pipeline.NewPipeline().WithNestedSteps("ensure backup",
					pipeline.NewStepFromFunc("fetch bucket secret", p.fetchS3BucketSecret),
					pipeline.NewStepFromFunc("ensure k8up schedule", p.ensureK8upSchedule),
				),
			),
			pipeline.If(pipeline.Not(p.isBackupEnabledPredicate(ctx)), pipeline.NewStepFromFunc("delete k8up schedule", deleteK8upSchedule)),
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

func (p *CreateStandalonePipeline) isBackupEnabledPredicate(ctx context.Context) pipeline.Predicate {
	return pipeline.BoolPtr(&getInstanceFromContext(ctx).Spec.Backup.Enabled)
}

// ensureDeploymentNamespace creates the deployment namespace where the Helm release is ultimately deployed in.
func (p *CreateStandalonePipeline) ensureDeploymentNamespace(ctx context.Context) error {
	instance := getInstanceFromContext(ctx)
	deploymentNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: generateClusterScopedNameForInstance(),
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, getClientFromContext(ctx), deploymentNamespace, func() error {
		// TODO: Add APPUiO cloud organization label that identifies ownership.
		deploymentNamespace.Labels = labels.Merge(deploymentNamespace.Labels, getCommonLabels(instance.Name))
		deploymentNamespace.Labels["app.kubernetes.io/instance-namespace"] = instance.Namespace
		return nil
	})
	setDeploymentNamespaceInContext(ctx, deploymentNamespace)
	return err
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
			Namespace: getDeploymentNamespaceFromContext(ctx).Name,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, getClientFromContext(ctx), secret, func() error {
		secret.Labels = labels.Merge(secret.Labels, getCommonLabels(getInstanceFromContext(ctx).Name))
		secret.StringData = map[string]string{
			"postgres-password":    generatePassword(),
			"password":             generatePassword(),
			"replication-password": generatePassword(),
		}
		return nil
	})
	return err
}

// ensureK8upSchedule creates the K8up schedule object.
func (p *CreateStandalonePipeline) ensureK8upSchedule(ctx context.Context) error {
	schedule := &k8upv1.Schedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: getInstanceFromContext(ctx).Status.HelmChart.DeploymentNamespace,
		},
	}

	config := getConfigFromContext(ctx)
	_, err := controllerutil.CreateOrUpdate(ctx, getClientFromContext(ctx), schedule, func() error {
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
					Endpoint:                 string(p.s3BucketSecret.Data[config.Spec.BackupConfigSpec.S3BucketSecret.EndpointRef.Key]),
					Bucket:                   string(p.s3BucketSecret.Data[config.Spec.BackupConfigSpec.S3BucketSecret.BucketRef.Key]),
					AccessKeyIDSecretRef:     &config.Spec.BackupConfigSpec.S3BucketSecret.AccessKeyRef,
					SecretAccessKeySecretRef: &config.Spec.BackupConfigSpec.S3BucketSecret.SecretKeyRef,
				},
			},
			FailedJobsHistoryLimit:     pointer.Int(2),
			SuccessfulJobsHistoryLimit: pointer.Int(2),
		}
		return nil
	})
	return err
}

func deleteK8upSchedule(ctx context.Context) error {
	instance := getInstanceFromContext(ctx)
	if instance.Status.HelmChart == nil || instance.Status.HelmChart.DeploymentNamespace == "" {
		// deployment namespace is unknown, we assume it has not been created
		return nil
	}
	schedule := newK8upSchedule(instance)
	err := getClientFromContext(ctx).Delete(ctx, schedule)
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
			Namespace: getDeploymentNamespaceFromContext(ctx).Name,
		},
	}
	secretKey := "repository"
	_, err := controllerutil.CreateOrUpdate(ctx, getClientFromContext(ctx), secret, func() error {
		secret.Labels = labels.Merge(secret.Labels, getCommonLabels(getInstanceFromContext(ctx).Name))
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
		Name:      getConfigFromContext(ctx).Spec.BackupConfigSpec.S3BucketSecret.BucketRef.Name,
		Namespace: getInstanceFromContext(ctx).Status.HelmChart.DeploymentNamespace,
	}}
	err := getClientFromContext(ctx).Get(ctx, client.ObjectKeyFromObject(p.s3BucketSecret), p.s3BucketSecret)
	return err
}

func (p *CreateStandalonePipeline) enrichStatus(ctx context.Context) error {
	instance := getInstanceFromContext(ctx)
	instance.Status.HelmChart = &v1alpha1.ChartMetaStatus{
		ChartMeta:           *getHelmChartFromContext(ctx),
		DeploymentNamespace: getDeploymentNamespaceFromContext(ctx).Name,
	}
	instance.Status.DeploymentStrategy = v1alpha1.StrategyHelmChart
	instance.Status.SetObservedGeneration(instance)
	err := getClientFromContext(ctx).Status().Update(ctx, instance)
	return err
}

// isHelmReleaseReady returns true if the ModifiedTime is non-zero.
//
// Note: This only works for first-time deployments. In the future another mechanism might be better.
// This step requires that fetchHelmRelease has run before.
func (p *CreateStandalonePipeline) isHelmReleaseReady(ctx context.Context) bool {
	instance := getInstanceFromContext(ctx)
	if instance.Status.HelmChart != nil && !instance.Status.HelmChart.ModifiedTime.IsZero() {
		return true
	}
	helmRelease := getHelmReleaseFromContext(ctx)
	if helmRelease.Status.Synced {
		if readyCondition := FindCrossplaneCondition(helmRelease.Status.Conditions, crossplanev1.TypeReady); readyCondition != nil && readyCondition.Status == corev1.ConditionTrue {
			instance.Status.HelmChart.ModifiedTime = readyCondition.LastTransitionTime
			return true
		}
	}
	return false
}

// markInstanceAsReady marks an instance immediately as ready by updating the status conditions.
func (p *CreateStandalonePipeline) markInstanceAsReady(ctx context.Context) error {
	instance := getInstanceFromContext(ctx)
	meta.SetStatusCondition(&instance.Status.Conditions, conditions.Builder().With(conditions.Ready()).WithGeneration(instance).Build())
	meta.RemoveStatusCondition(&instance.Status.Conditions, conditions.TypeCreating)
	return getClientFromContext(ctx).Status().Update(ctx, instance)
}

// ensureConnectionSecret creates the connection secret in the instance's namespace.
func (p *CreateStandalonePipeline) ensureConnectionSecret(ctx context.Context) error {
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: p.connectionSecret.Name, Namespace: p.connectionSecret.Namespace}}
	_, err := controllerutil.CreateOrUpdate(ctx, getClientFromContext(ctx), secret, func() error {
		secret.Labels = labels.Merge(secret.Labels, p.connectionSecret.Labels)
		secret.Data = p.connectionSecret.Data
		secret.StringData = p.connectionSecret.StringData
		return nil
	})
	if err != nil {
		return err
	}
	p.connectionSecret = secret
	return nil
}

// fetchCredentialSecret gets the credential secret and puts the credentials for postgresql into the connection secret.
func (p *CreateStandalonePipeline) fetchCredentialSecret(ctx context.Context) error {
	instance := getInstanceFromContext(ctx)
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      getCredentialSecretName(),
		Namespace: instance.Status.HelmChart.DeploymentNamespace,
	}}
	err := getClientFromContext(ctx).Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, secret)
	if err != nil {
		return err
	}
	if instance.Spec.Parameters.EnableSuperUser {
		p.addDataToConnectionSecret(ctx, "POSTGRESQL_POSTGRES_PASSWORD", secret.Data["postgres-password"])
	}
	p.addDataToConnectionSecret(ctx, "POSTGRESQL_PASSWORD", secret.Data["password"])
	p.addStringDataToConnectionSecret(ctx, "POSTGRESQL_DATABASE", instance.Name)
	p.addStringDataToConnectionSecret(ctx, "POSTGRESQL_USER", instance.Name)
	return err
}

// fetchService gets the service object and puts the DNS records into the connection secret.
func (p *CreateStandalonePipeline) fetchService(ctx context.Context) error {
	service := &corev1.Service{}
	err := getClientFromContext(ctx).Get(ctx, client.ObjectKey{Name: "postgresql", Namespace: getInstanceFromContext(ctx).Status.HelmChart.DeploymentNamespace}, service)
	if err != nil {
		return err
	}
	p.addStringDataToConnectionSecret(ctx, "POSTGRESQL_SERVICE_NAME", fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace))
	p.addStringDataToConnectionSecret(ctx, "POSTGRESQL_SERVICE_URL", fmt.Sprintf("postgresql://%s.%s.svc.cluster.local:%d", service.Name, service.Namespace, service.Spec.Ports[0].Port))
	p.addStringDataToConnectionSecret(ctx, "POSTGRESQL_SERVICE_PORT", fmt.Sprintf("%d", service.Spec.Ports[0].Port))
	return nil
}

func (p *CreateStandalonePipeline) addDataToConnectionSecret(ctx context.Context, key string, data []byte) {
	if p.connectionSecret == nil {
		p.connectionSecret = newConnectionSecret(ctx)
	}
	if p.connectionSecret.Data == nil {
		p.connectionSecret.Data = map[string][]byte{}
	}
	p.connectionSecret.Data[key] = data
}

func (p *CreateStandalonePipeline) addStringDataToConnectionSecret(ctx context.Context, key, data string) {
	if p.connectionSecret == nil {
		p.connectionSecret = newConnectionSecret(ctx)
	}
	if p.connectionSecret.StringData == nil {
		p.connectionSecret.StringData = map[string]string{}
	}
	p.connectionSecret.StringData[key] = data
}

func newConnectionSecret(ctx context.Context) *corev1.Secret {
	instance := getInstanceFromContext(ctx)
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetConnectionSecretName(),
			Namespace: instance.Namespace,
			Labels:    getCommonLabels(instance.Name),
		},
		StringData: map[string]string{},
		Data:       map[string][]byte{},
	}
}

func (p *CreateStandalonePipeline) setOwnerReferenceInConnectionSecret(ctx context.Context) error {
	return controllerutil.SetOwnerReference(getInstanceFromContext(ctx), p.connectionSecret, getClientFromContext(ctx).Scheme())
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

func getPVCName() string {
	return "postgresql-data"
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
