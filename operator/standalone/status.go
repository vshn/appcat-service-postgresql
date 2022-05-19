package standalone

import (
	"context"

	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// FindCrossplaneCondition finds the conditionType in conditions.
func FindCrossplaneCondition(conditions []crossplanev1.Condition, conditionType crossplanev1.ConditionType) *crossplanev1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}

	return nil
}

// setConditionFn returns a func that immediately updates the instance with the given status condition.
// The condition's LastTransitionTime is set to Now() just before updating.
func setConditionFn(obj client.Object, conditions *[]metav1.Condition, condition metav1.Condition) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		clt := getClientFromContext(ctx)
		condition.LastTransitionTime = metav1.Now()
		meta.SetStatusCondition(conditions, condition)
		return clt.Status().Update(ctx, obj)
	}
}

// addFinalizerFn returns a func that immediately updates the instance with the given finalizer.
func addFinalizerFn(obj client.Object, finalizer string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		clt := getClientFromContext(ctx)
		if controllerutil.AddFinalizer(obj, finalizer) {
			return clt.Update(ctx, obj)
		}
		return nil
	}
}
