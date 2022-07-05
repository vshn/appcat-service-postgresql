package steps

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// EnsurePvcFn ensures that the PVC is created.
func EnsurePvcFn(labelSet labels.Set) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)
		config := GetConfigFromContext(ctx)
		deploymentNamespace := getFromContextOrPanic(ctx, DeploymentNamespaceKey{}).(*corev1.Namespace)

		persistentVolumeClaim := newPVC(deploymentNamespace.Name)
		persistentVolumeClaim.Spec.AccessModes = config.Spec.Persistence.AccessModes
		persistentVolumeClaim.Spec.StorageClassName = config.Spec.Persistence.StorageClassName

		_, err := controllerutil.CreateOrUpdate(ctx, kube, persistentVolumeClaim, func() error {
			persistentVolumeClaim.Labels = labels.Merge(persistentVolumeClaim.Labels, labelSet)
			persistentVolumeClaim.Spec.Resources.Requests[corev1.ResourceStorage] = *instance.Spec.Parameters.Resources.StorageCapacity
			return nil
		})
		return err
	}
}

// DeletePvcFn deletes the corev1.PersistentVolumeClaim from the deployment namespace.
// Ignore "not found" error and returns nil if deployment namespace is unknown.
func DeletePvcFn() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)

		if instance.Status.HelmChart == nil || instance.Status.HelmChart.DeploymentNamespace == "" {
			// instance might have never been properly deployed
			return nil
		}

		pvc := newPVC(instance.Status.HelmChart.DeploymentNamespace)
		err := kube.Delete(ctx, pvc)
		return client.IgnoreNotFound(err)
	}
}

func newPVC(ns string) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getPVCName(),
			Namespace: ns,
		},
	}
	pvc.Spec.Resources.Requests = map[corev1.ResourceName]resource.Quantity{}
	return pvc
}
