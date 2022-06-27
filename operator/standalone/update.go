package standalone

import (
	"context"
	"fmt"
	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/vshn/appcat-service-postgresql/apis/conditions"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/helmvalues"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// PostgresqlPodName the name of the postgresql instance pod running on the cluster
	PostgresqlPodName string = "postgresql-0"
)

// UpdateStandalonePipeline is a pipeline that updates an existing instance in the target deployment namespace.
type UpdateStandalonePipeline struct {
	operatorNamespace   string
	client              client.Client
	instance            *v1alpha1.PostgresqlStandalone
	config              *v1alpha1.PostgresqlStandaloneOperatorConfig
	helmValues          helmvalues.V
	deploymentNamespace *corev1.Namespace
	helmRelease         *helmv1beta1.Release

	connectionSecret      *corev1.Secret
	persistentVolumeClaim *corev1.PersistentVolumeClaim
}

// NewUpdateStandalonePipeline creates a new pipeline with the required dependencies.
func NewUpdateStandalonePipeline(client client.Client, instance *v1alpha1.PostgresqlStandalone, operatorNamespace string) *UpdateStandalonePipeline {
	return &UpdateStandalonePipeline{
		instance:          instance,
		client:            client,
		operatorNamespace: operatorNamespace,
	}
}

// RunInitialUpdatePipeline executes the pipeline with configured business logic steps.
// This should only be executed once per pipeline as it stores intermediate results in the struct.
func (u *UpdateStandalonePipeline) RunInitialUpdatePipeline(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("fetch operator config", u.fetchOperatorConfig),

			pipeline.NewPipeline().WithNestedSteps("ensure connection secret",
				pipeline.NewStepFromFunc("fetch connection secret", u.fetchConnectionSecret),
				pipeline.If(u.isRootUserNotConsistentWithSecretData,
					pipeline.NewStepFromFunc("ensure connection secret", u.ensureConnectionSecret),
				),
			),

			pipeline.NewPipeline().WithNestedSteps("ensure persistent volume claim",
				pipeline.NewStepFromFunc("fetch persistent volume claim", u.fetchPVC),
				pipeline.If(u.isStorageUpdated,
					pipeline.NewPipeline().WithNestedSteps("update helm release",
						pipeline.NewStepFromFunc("ensure persistent volume claim", u.ensurePersistentVolumeClaim),
						pipeline.NewStepFromFunc("restart postgresql pod", u.restartPostgresqlPod),
					),
				),
			),

			pipeline.NewPipeline().WithNestedSteps("update helm release",
				pipeline.NewStepFromFunc("fetch helm release", u.fetchHelmRelease),
				pipeline.NewStepFromFunc("use release values", u.useReleaseValues),
				pipeline.NewStepFromFunc("apply changes to helm release", u.applyValuesFromInstance),
				pipeline.NewStepFromFunc("ensure helm release", u.ensureHelmRelease),
			),

			pipeline.NewStepFromFunc("update status condition", u.markInstanceAsProgressing),
		).
		RunWithContext(ctx).Err()
}

// WaitUntilAllResourceReady runs a pipeline that verifies if all dependent resources are ready.
// It will add the conditions.TypeReady condition to the status field (and update it) if it's considered ready.
// No error is returned in case the instance is not considered ready.
func (u *UpdateStandalonePipeline) WaitUntilAllResourceReady(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("fetch helm release", u.fetchHelmRelease),
			pipeline.If(pipeline.And(u.isHelmReleaseReady, u.isPodRunning),
				pipeline.NewPipeline().WithNestedSteps("finish updating",
					pipeline.NewStepFromFunc("mark instance ready", u.markInstanceAsReady),
				),
			),
		).
		RunWithContext(ctx).Err()
}

// fetchHelmRelease fetches the Helm release for the given instance.
func (u *UpdateStandalonePipeline) fetchHelmRelease(ctx context.Context) error {
	helmRelease := &helmv1beta1.Release{}
	err := u.client.Get(ctx, client.ObjectKey{Name: u.instance.Status.HelmChart.DeploymentNamespace}, helmRelease)
	u.helmRelease = helmRelease
	return err
}

// useReleaseValues unmarshalls and saves release values to be used in subsequent steps
//
// This step assumes that the release has been fetched first via fetchHelmRelease.
func (u *UpdateStandalonePipeline) useReleaseValues(_ context.Context) error {
	values := helmvalues.V{}
	err := helmvalues.Unmarshal(u.helmRelease.Spec.ForProvider.Values, &values)
	u.helmValues = values
	return err
}

// applyValuesFromInstance merges the user-defined and -exposed Helm values into the current Helm values map.
func (u *UpdateStandalonePipeline) applyValuesFromInstance(_ context.Context) error {
	resources := helmvalues.V{
		"auth": helmvalues.V{
			"enablePostgresUser": u.instance.Spec.Parameters.EnableSuperUser,
			"existingSecret":     getCredentialSecretName(),
			"database":           u.instance.Name,
			"username":           u.instance.Name,
		},
		"primary": helmvalues.V{
			"resources": helmvalues.V{
				"limits": helmvalues.V{
					"memory": u.instance.Spec.Parameters.Resources.MemoryLimit.String(),
				},
			},
			"persistence": helmvalues.V{
				"size": u.instance.Spec.Parameters.Resources.StorageCapacity.String(),
			},
		},
		"fullnameOverride": getDeploymentName(),
		"networkPolicy": helmvalues.V{
			"enabled": true,
			"ingressRules": helmvalues.V{
				"primaryAccessOnlyFrom": helmvalues.V{
					"enabled": true,
					"namespaceSelector": helmvalues.V{
						"kubernetes.io/metadata.name": u.instance.Namespace,
					},
				},
			},
		},
	}
	helmvalues.Merge(resources, &u.helmValues)
	return nil
}

// ensureHelmRelease updates the Helm release object
func (u *UpdateStandalonePipeline) ensureHelmRelease(ctx context.Context) error {
	chart := &u.config.Spec.HelmReleaseTemplate.Chart
	helmValues, err := helmvalues.Marshal(u.helmValues)
	if err != nil {
		return err
	}
	u.helmRelease = &helmv1beta1.Release{
		ObjectMeta: metav1.ObjectMeta{
			Name:   u.instance.Status.HelmChart.DeploymentNamespace,
			Labels: getCommonLabels(u.instance.Name),
		},
	}

	updateRelease := func() error {
		u.helmRelease.Spec = helmv1beta1.ReleaseSpec{
			ForProvider: helmv1beta1.ReleaseParameters{
				Chart:               helmv1beta1.ChartSpec{Repository: chart.Repository, Name: chart.Name, Version: chart.Version},
				Namespace:           u.instance.Status.HelmChart.DeploymentNamespace,
				SkipCreateNamespace: true,
				SkipCRDs:            true,
				Wait:                true,
				ValuesSpec:          helmv1beta1.ValuesSpec{Values: helmValues},
			},
			ResourceSpec: crossplanev1.ResourceSpec{
				ProviderConfigReference: &crossplanev1.Reference{Name: u.config.Spec.HelmProviderConfigReference},
			},
		}
		return nil
	}
	_, err = controllerutil.CreateOrUpdate(ctx, u.client, u.helmRelease, updateRelease)
	return err
}

func (u *UpdateStandalonePipeline) fetchOperatorConfig(ctx context.Context) error {
	list := &v1alpha1.PostgresqlStandaloneOperatorConfigList{}
	labels := client.MatchingLabels{
		v1alpha1.PostgresqlMajorVersionLabelKey: u.instance.Spec.Parameters.MajorVersion.String(),
	}
	ns := client.InNamespace(u.operatorNamespace)
	err := u.client.List(ctx, list, labels, ns)
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		return fmt.Errorf("no %s found with label '%s' in namespace '%s'", v1alpha1.PostgresqlStandaloneOperatorConfigKind, labels, ns)
	}
	if len(list.Items) > 1 {
		return fmt.Errorf("multiple versions of %s found with label '%s' in namespace '%s'", v1alpha1.PostgresqlStandaloneOperatorConfigKind, labels, ns)
	}
	u.config = &list.Items[0]
	return nil
}

func (u *UpdateStandalonePipeline) ensureConnectionSecret(ctx context.Context) error {
	var updateSecretFunc func() error
	if u.instance.Spec.Parameters.EnableSuperUser {
		credentialsSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
			Name:      getCredentialSecretName(),
			Namespace: u.instance.Status.HelmChart.DeploymentNamespace,
		}}
		err := getClientFromContext(ctx).Get(ctx, types.NamespacedName{Name: credentialsSecret.Name, Namespace: credentialsSecret.Namespace}, credentialsSecret)
		if err != nil {
			return err
		}
		updateSecretFunc = func() error {
			u.connectionSecret.Data["POSTGRESQL_POSTGRES_PASSWORD"] = credentialsSecret.Data["postgres-password"]
			return nil
		}
	} else {
		updateSecretFunc = func() error {
			delete(u.connectionSecret.Data, "POSTGRESQL_POSTGRES_PASSWORD")
			return nil
		}
	}
	_, err := controllerutil.CreateOrUpdate(ctx, u.client, u.connectionSecret, updateSecretFunc)
	return err
}

func (u *UpdateStandalonePipeline) ensurePersistentVolumeClaim(ctx context.Context) error {
	quantity, err := resource.ParseQuantity(u.instance.Spec.Parameters.Resources.StorageCapacity.String())
	if err != nil {
		return err
	}
	updatePVCFunc := func() error {
		u.persistentVolumeClaim.Spec.Resources.Requests = map[corev1.ResourceName]resource.Quantity{corev1.ResourceStorage: quantity}
		return nil
	}
	_, err = controllerutil.CreateOrUpdate(ctx, u.client, u.persistentVolumeClaim, updatePVCFunc)
	return err
}

func (u *UpdateStandalonePipeline) restartPostgresqlPod(ctx context.Context) error {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PostgresqlPodName,
			Namespace: u.instance.Status.HelmChart.DeploymentNamespace,
		},
	}

	// Restart the pod so that the new PVC storage takes effect
	return u.client.Delete(ctx, pod)
}

func (u *UpdateStandalonePipeline) isStorageUpdated(_ context.Context) bool {
	quantity, err := resource.ParseQuantity(u.instance.Spec.Parameters.Resources.StorageCapacity.String())
	if err != nil {
		return false
	}
	if *u.persistentVolumeClaim.Spec.Resources.Requests.Storage() != quantity {
		return true
	}
	return false
}

func (u *UpdateStandalonePipeline) fetchPVC(ctx context.Context) error {
	persistentVolumeClaim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getPVCName(),
			Namespace: u.instance.Status.HelmChart.DeploymentNamespace,
		},
	}

	err := getClientFromContext(ctx).Get(ctx, types.NamespacedName{Name: persistentVolumeClaim.Name, Namespace: persistentVolumeClaim.Namespace}, persistentVolumeClaim)
	if err != nil {
		return err
	}
	u.persistentVolumeClaim = persistentVolumeClaim
	return nil
}

// isHelmReleaseReady returns true if the ModifiedTime is non-zero.
//
// Note: This only works for first-time deployments. In the future another mechanism might be better.
// This step requires that fetchHelmRelease has run before.
func (u *UpdateStandalonePipeline) isHelmReleaseReady(_ context.Context) bool {
	if u.instance.Status.HelmChart != nil && !u.instance.Status.HelmChart.ModifiedTime.IsZero() {
		return true
	}
	if u.helmRelease.Status.Synced {
		if readyCondition := FindCrossplaneCondition(u.helmRelease.Status.Conditions, crossplanev1.TypeReady); readyCondition != nil && readyCondition.Status == corev1.ConditionTrue {
			u.instance.Status.HelmChart.ModifiedTime = readyCondition.LastTransitionTime
			return true
		}
	}
	return false
}

// isPodRunning returns true if postgresql instance pod is running.
//
// Note: The pod may have been restarted during RunInitialUpdatePipeline thus
// it is necessary to check if the pod is running
func (u *UpdateStandalonePipeline) isPodRunning(ctx context.Context) bool {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PostgresqlPodName,
			Namespace: u.instance.Status.HelmChart.DeploymentNamespace,
		},
	}
	err := u.client.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pod)
	if err == nil || isStatusConditionTrue(pod.Status.Conditions, conditions.TypeReady) {
		return true
	}
	return false
}

// markInstanceAsProgressing marks an instance as progressing by updating the status conditions.
func (u *UpdateStandalonePipeline) markInstanceAsProgressing(ctx context.Context) error {
	meta.SetStatusCondition(
		&u.instance.Status.Conditions,
		conditions.Builder().
			With(conditions.Progressing()).
			WithGeneration(u.instance).
			Build(),
	)
	meta.SetStatusCondition(
		&u.instance.Status.Conditions,
		conditions.Builder().
			With(conditions.Ready(metav1.ConditionFalse)).
			WithGeneration(u.instance).
			Build(),
	)
	return u.client.Status().Update(ctx, u.instance)
}

// markInstanceAsReady marks an instance immediately as ready by updating the status conditions.
func (u *UpdateStandalonePipeline) markInstanceAsReady(ctx context.Context) error {
	meta.SetStatusCondition(
		&u.instance.Status.Conditions,
		conditions.Builder().
			With(conditions.Ready(metav1.ConditionTrue)).
			WithGeneration(u.instance).
			Build(),
	)
	meta.RemoveStatusCondition(&u.instance.Status.Conditions, conditions.TypeProgressing)
	return u.client.Status().Update(ctx, u.instance)
}

//  fetchConnectionSecret gets the connection secret for this instance.
func (u *UpdateStandalonePipeline) fetchConnectionSecret(ctx context.Context) error {
	connectionSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      u.instance.Spec.WriteConnectionSecretToRef.Name,
		Namespace: u.instance.Namespace,
	}}
	err := u.client.Get(ctx, types.NamespacedName{Name: connectionSecret.Name, Namespace: connectionSecret.Namespace}, connectionSecret)
	if err != nil {
		return err
	}
	u.connectionSecret = connectionSecret
	return nil
}

// isRootUserNotConsistentWithSecretData checks whether data in secret is consistent with EnableSuperUser value
func (u *UpdateStandalonePipeline) isRootUserNotConsistentWithSecretData(_ context.Context) bool {
	enabledUser := u.instance.Spec.Parameters.EnableSuperUser
	rootUserExists := u.connectionSecret.Data["POSTGRESQL_POSTGRES_PASSWORD"] != nil
	// Changed only when both variables have different values
	if (enabledUser || rootUserExists) && !(enabledUser && rootUserExists) {
		return true
	}
	return false
}

// isStatusConditionTrue checks whether pod condition type is active from the pod conditions array
func isStatusConditionTrue(podConditions []corev1.PodCondition, conditionType corev1.PodConditionType) bool {
	for _, condition := range podConditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
