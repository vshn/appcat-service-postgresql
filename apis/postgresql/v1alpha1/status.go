package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GenerationStatus struct {
	// ObservedGeneration is the meta.generation number this resource was last reconciled with.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// SetObservedGeneration sets the ObservedGeneration from the given ObjectMeta.
func (in *GenerationStatus) SetObservedGeneration(obj client.Object) {
	in.ObservedGeneration = obj.GetGeneration()
}
