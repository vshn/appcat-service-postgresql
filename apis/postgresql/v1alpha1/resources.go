package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// Resources is the common set of high-level scalable resources for an instance.
type Resources struct {
	ComputeResources `json:",inline"`
	StorageResources `json:",inline"`
}

// ComputeResources contains the high-level scalable compute resources for an instance.
type ComputeResources struct {
	// MemoryLimit defines the maximum memory limit designated for the instance.
	// It can be freely scaled up or down within the operator-configured limits.
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
}

// StorageResources contains the high-level scalable storage resources for an instance.
type StorageResources struct {
	// StorageCapacity is the reserved storage size for a PersistentVolume.
	// It can only grow and never shrink.
	// Attempt to shrink the size will throw a validation error.
	// Minimum and Maximum is defined on an operator level.
	StorageCapacity *resource.Quantity `json:"storageCapacity,omitempty"`
}
