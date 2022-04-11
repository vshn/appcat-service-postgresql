package v1alpha1

// ChartMeta contains the metadata to a Helm chart.
type ChartMeta struct {
	// Repository is the Helm chart repository URL.
	Repository string `json:"repository"`
	// Version is the Helm chart version identifier.
	Version string `json:"version"`
	// Name is the Helm chart name within the repository.
	Name string `json:"name"`
}
