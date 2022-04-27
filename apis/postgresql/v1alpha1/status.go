package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GenerationStatus struct {
	// ObservedGeneration is the meta.generation number this resource was last reconciled with.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// SetObservedGeneration sets the ObservedGeneration from the given ObjectMeta.
func (in *GenerationStatus) SetObservedGeneration(meta metav1.ObjectMeta) {
	in.ObservedGeneration = meta.Generation
}
