package standalone

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DeleteStandalonePipeline is a pipeline that deletes an instance from the target deployment namespace.
type DeleteStandalonePipeline struct {
	client             client.Client
	instance           *v1alpha1.PostgresqlStandalone
	helmReleaseDeleted bool
}

// NewDeleteStandalonePipeline creates a new delete pipeline with the required dependencies.
func NewDeleteStandalonePipeline(client client.Client, instance *v1alpha1.PostgresqlStandalone) *DeleteStandalonePipeline {
	return &DeleteStandalonePipeline{
		instance: instance,
		client:   client,
	}
}

// RunPipeline executes the pipeline with configured business logic steps.
// The pipeline requires multiple reconciliations due to asynchronous deletion of resources in background
// The Helm Release step requires a complete removal of its resources before moving to the next step
func (d *DeleteStandalonePipeline) RunPipeline(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("delete connection secret", d.deleteConnectionSecret),
			pipeline.NewStepFromFunc("delete helm release", d.deleteHelmRelease),
			pipeline.If(d.isHelmReleaseDeleted,
				pipeline.NewPipeline().WithNestedSteps("finalize",
					pipeline.NewStepFromFunc("delete namespace", d.deleteNamespace),
					pipeline.NewStepFromFunc("remove finalizer", d.removeFinalizer),
				),
			),
		).
		RunWithContext(ctx).Err()
}

// deleteHelmRelease removes the Helm Release from the cluster
// We may reconcile multiple times until Release is completely deleted hence we use helmReleaseDeleted variable
func (d *DeleteStandalonePipeline) deleteHelmRelease(ctx context.Context) error {
	helmRelease := &helmv1beta1.Release{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.instance.Status.HelmChart.DeploymentNamespace,
		},
	}
	err := d.client.Delete(ctx, helmRelease)
	if err != nil && apierrors.IsNotFound(err) {
		d.helmReleaseDeleted = true
	}
	return client.IgnoreNotFound(err)
}

// deleteNamespace removes the namespace of the PostgreSQL instance
// We delete the namespace only if the Helm Release has been deleted
func (d *DeleteStandalonePipeline) deleteNamespace(ctx context.Context) error {
	deploymentNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.instance.Status.HelmChart.DeploymentNamespace,
		},
	}
	propagation := metav1.DeletePropagationBackground
	deleteOptions := &client.DeleteOptions{
		PropagationPolicy: &propagation,
	}
	return client.IgnoreNotFound(d.client.Delete(ctx, deploymentNamespace, deleteOptions))
}

// deleteConnectionSecret removes the connection secret of the PostgreSQL instance
func (d *DeleteStandalonePipeline) deleteConnectionSecret(ctx context.Context) error {
	connectionSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.instance.Spec.WriteConnectionSecretToRef.Name,
			Namespace: d.instance.Namespace,
		},
	}
	return client.IgnoreNotFound(d.client.Delete(ctx, connectionSecret))
}

// removeFinalizer removes the finalizer from the PostgreSQL CRD
func (d *DeleteStandalonePipeline) removeFinalizer(ctx context.Context) error {
	if controllerutil.RemoveFinalizer(d.instance, finalizer) {
		return d.client.Update(ctx, d.instance)
	}
	return nil
}

// isHelmReleaseDeleted checks whether the Release was completely deleted
func (d *DeleteStandalonePipeline) isHelmReleaseDeleted(_ context.Context) bool {
	return d.helmReleaseDeleted
}
