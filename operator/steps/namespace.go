package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// AppuioOrganizationLabelKey is the label key required for setting ownership of a namespace
const AppuioOrganizationLabelKey = "appuio.io/organization"

// EnsureNamespace creates the namespace with given name and labels.
func EnsureNamespace(name string, labelSet labels.Set) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instanceNamespace := getFromContextOrPanic(ctx, InstanceNamespaceKey{}).(*corev1.Namespace)
		instanceNsLabels := instanceNamespace.Labels
		copyLabels := labels.Set{}
		if org, exists := instanceNsLabels[AppuioOrganizationLabelKey]; exists {
			copyLabels[AppuioOrganizationLabelKey] = org
		}

		deploymentNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		}
		_, err := controllerutil.CreateOrUpdate(ctx, kube, deploymentNamespace, func() error {
			deploymentNamespace.Labels = labels.Merge(deploymentNamespace.Labels, labels.Merge(copyLabels, labelSet))
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

// FetchNamespaceFn fetches the namespace of the given name and stores it in the context with given key.
func FetchNamespaceFn(namespaceName string, contextKey any) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)

		ns := &corev1.Namespace{}
		err := kube.Get(ctx, types.NamespacedName{Name: namespaceName}, ns)
		pipeline.StoreInContext(ctx, contextKey, ns)
		return err
	}
}
