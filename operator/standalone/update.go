package standalone

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/vshn/appcat-service-postgresql/apis/conditions"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// PostgresqlPodName is the name of the PostgreSQL instance pod running in the deployment namespace
	PostgresqlPodName string = "postgresql-0"
)

// UpdateStandalonePipeline is a pipeline that updates an existing instance in the target deployment namespace.
type UpdateStandalonePipeline struct {
}

// NewUpdateStandalonePipeline creates a new pipeline with the required dependencies.
func NewUpdateStandalonePipeline() *UpdateStandalonePipeline {
	return &UpdateStandalonePipeline{}
}

// RunInitialUpdatePipeline executes the pipeline with configured business logic steps.
// This should only be executed once per pipeline as it stores intermediate results in the struct.
func (u *UpdateStandalonePipeline) RunInitialUpdatePipeline(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("fetch operator config", fetchOperatorConfig),
			pipeline.NewStepFromFunc("mark instance as progressing", u.markInstanceAsProgressing),
			pipeline.NewStepFromFunc("patch connection secret", u.patchConnectionSecret),
			pipeline.NewStepFromFunc("ensure persistent volume claim", ensurePVC),

			pipeline.NewPipeline().WithNestedSteps("compile helm values",
				pipeline.NewStepFromFunc("read template values", useTemplateValues),
				pipeline.NewStepFromFunc("override template values", overrideTemplateValues),
				pipeline.NewStepFromFunc("apply values from instance", applyValuesFromInstance),
			),
			pipeline.NewStepFromFunc("ensure helm release", ensureHelmRelease),
		).
		RunWithContext(ctx).Err()
}

// WaitUntilAllResourceReady runs a pipeline that verifies if all dependent resources are ready.
// It will add the conditions.TypeReady condition to the status field (and update it) if it's considered ready.
// No error is returned in case the instance is not considered ready.
func (u *UpdateStandalonePipeline) WaitUntilAllResourceReady(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("fetch helm release", fetchHelmRelease),
			pipeline.If(u.isHelmReleaseReady, pipeline.NewStepFromFunc("mark instance ready", u.markInstanceAsReady)),
		).
		RunWithContext(ctx).Err()
}

func (u *UpdateStandalonePipeline) patchConnectionSecret(ctx context.Context) error {
	client := getClientFromContext(ctx)
	instance := getInstanceFromContext(ctx)
	connectionSecret := newConnectionSecret(ctx)
	_, err := controllerutil.CreateOrUpdate(ctx, client, connectionSecret, func() error {
		if instance.Spec.Parameters.EnableSuperUser {
			credentialsSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Name:      getCredentialSecretName(),
				Namespace: instance.Status.HelmChart.DeploymentNamespace,
			}}
			err := client.Get(ctx, types.NamespacedName{Name: credentialsSecret.Name, Namespace: credentialsSecret.Namespace}, credentialsSecret)
			if err != nil {
				return err
			}
			connectionSecret.Data["POSTGRESQL_POSTGRES_PASSWORD"] = credentialsSecret.Data["postgres-password"]
		} else {
			delete(connectionSecret.Data, "POSTGRESQL_POSTGRES_PASSWORD")
		}
		return nil
	})

	return err
}

// isHelmReleaseReady returns true if the ModifiedTime is non-zero.
//
// Note: This only works for first-time deployments. In the future another mechanism might be better.
// This step requires that fetchHelmRelease has run before.
func (u *UpdateStandalonePipeline) isHelmReleaseReady(ctx context.Context) bool {
	instance := getInstanceFromContext(ctx)
	helmRelease := getHelmReleaseFromContext(ctx)
	if helmRelease.Status.Synced {
		if readyCondition := FindCrossplaneCondition(helmRelease.Status.Conditions, crossplanev1.TypeReady); readyCondition != nil && readyCondition.Status == corev1.ConditionTrue {
			instance.Status.HelmChart.ModifiedTime = readyCondition.LastTransitionTime
			return true
		}
	}
	return false
}

// markInstanceAsProgressing marks an instance as progressing by updating the status conditions.
func (u *UpdateStandalonePipeline) markInstanceAsProgressing(ctx context.Context) error {
	instance := getInstanceFromContext(ctx)
	meta.SetStatusCondition(
		&instance.Status.Conditions,
		conditions.Builder().
			With(conditions.Progressing()).
			WithGeneration(instance).
			Build(),
	)
	meta.SetStatusCondition(
		&instance.Status.Conditions,
		conditions.Builder().
			With(conditions.NotReady()).
			WithGeneration(instance).
			Build(),
	)
	instance.Status.SetObservedGeneration(instance)
	return getClientFromContext(ctx).Status().Update(ctx, instance)
}

// markInstanceAsReady marks an instance immediately as ready by updating the status conditions.
func (u *UpdateStandalonePipeline) markInstanceAsReady(ctx context.Context) error {
	instance := getInstanceFromContext(ctx)
	meta.SetStatusCondition(
		&instance.Status.Conditions,
		conditions.Builder().
			With(conditions.Ready()).
			WithGeneration(instance).
			Build(),
	)
	meta.RemoveStatusCondition(&instance.Status.Conditions, conditions.TypeProgressing)
	return getClientFromContext(ctx).Status().Update(ctx, instance)
}
