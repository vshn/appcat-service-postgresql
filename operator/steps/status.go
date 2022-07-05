package steps

import (
	"context"
	"github.com/vshn/appcat-service-postgresql/apis/conditions"
	"k8s.io/apimachinery/pkg/api/meta"
)

// MarkInstanceAsReadyFn marks an instance as ready by updating the status conditions.
func MarkInstanceAsReadyFn() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)

		meta.SetStatusCondition(
			&instance.Status.Conditions,
			conditions.Builder().
				With(conditions.Ready()).
				WithGeneration(instance).
				Build(),
		)
		meta.RemoveStatusCondition(&instance.Status.Conditions, conditions.TypeProgressing)
		return kube.Status().Update(ctx, instance)
	}
}

// MarkInstanceAsProgressingFn marks an instance as progressing by updating the status conditions.
func MarkInstanceAsProgressingFn() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)

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
		return kube.Status().Update(ctx, instance)
	}
}
