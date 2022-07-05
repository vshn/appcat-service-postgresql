package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// EnsureNamespace creates the namespace with given name and labels.
func EnsureNamespace(name string, labelSet labels.Set) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)

		deploymentNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		}
		_, err := controllerutil.CreateOrUpdate(ctx, kube, deploymentNamespace, func() error {
			deploymentNamespace.Labels = labels.Merge(deploymentNamespace.Labels, labelSet)
			return nil
		})
		pipeline.StoreInContext(ctx, DeploymentNamespaceKey{}, deploymentNamespace)
		return err
	}
}

// DeleteNamespaceFn deletes the namespace where the instance is deployed.
// Ignore "not found" error and returns nil if deployment namespace is unknown.
func DeleteNamespaceFn() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)

		if instance.Status.GetDeploymentNamespace() == "" {
			// Namespace might not ever have existed, skip
			return nil
		}
		deploymentNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: instance.Status.GetDeploymentNamespace(),
			},
		}
		propagation := metav1.DeletePropagationBackground
		deleteOptions := &client.DeleteOptions{
			PropagationPolicy: &propagation,
		}
		err := kube.Delete(ctx, deploymentNamespace, deleteOptions)
		return client.IgnoreNotFound(err)
	}
}
