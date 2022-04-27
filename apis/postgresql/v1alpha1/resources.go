package v1alpha1

type Resources struct {
	ComputeResources `json:",inline"`
	StorageResources `json:",inline"`
}

type ComputeResources struct {
	MemoryLimit string `json:"memoryLimit,omitempty"`
}

type StorageResources struct {
	Size string `json:"size,omitempty"`
}
