package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

type Resources struct {
	ComputeResources `json:",inline"`
	StorageResources `json:",inline"`
}

type ComputeResources struct {
	// MemoryLimit defines the maximum memory limit designated for the instance.
	// It can be freely scaled up or down within some limits.
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
}

type StorageResources struct {
	// StorageCapacity is the reserved storage size for a PersistentVolume.
	// It can only grow and never shrink.
	// Attempt to shrink the size will throw a validation error.
	StorageCapacity *resource.Quantity `json:"storageCapacity,omitempty"`
}
