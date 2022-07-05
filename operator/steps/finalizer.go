package steps

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// AddFinalizerFn returns a func that immediately updates the instance with the given finalizer.
func AddFinalizerFn(obj client.Object, finalizer string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)

		if controllerutil.AddFinalizer(obj, finalizer) {
			return kube.Update(ctx, obj)
		}
		return nil
	}
}

// RemoveFinalizerFn removes the finalizer from the PostgresqlStandalone instance and updates it if there was a finalizer present.
func RemoveFinalizerFn(finalizer string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)

		if controllerutil.RemoveFinalizer(instance, finalizer) {
			return kube.Update(ctx, instance)
		}
		return nil
	}
}
