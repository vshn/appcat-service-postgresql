package conditions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConditionBuilder builds Conditions using various properties.
type ConditionBuilder struct {
	condition metav1.Condition
}

// Builder returns a new ConditionBuilder instance.
func Builder() *ConditionBuilder {
	return &ConditionBuilder{}
}

// With initializes the condition with the given value.
// Returns itself for convenience.
func (b *ConditionBuilder) With(condition metav1.Condition) *ConditionBuilder {
	b.condition = condition
	return b
}

// WithMessage sets the condition message.
// Returns itself for convenience.
func (b *ConditionBuilder) WithMessage(message string) *ConditionBuilder {
	b.condition.Message = message
	return b
}

// WithGeneration sets ObservedGeneration from the given object.
// Returns itself for convenience.
func (b *ConditionBuilder) WithGeneration(object client.Object) *ConditionBuilder {
	b.condition.ObservedGeneration = object.GetGeneration()
	return b
}

// Build returns the condition.
func (b *ConditionBuilder) Build() metav1.Condition {
	return b.condition
}
