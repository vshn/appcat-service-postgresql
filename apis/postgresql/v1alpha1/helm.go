package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ChartMeta contains the metadata to a Helm chart.
type ChartMeta struct {
	// Repository is the Helm chart repository URL.
	Repository string `json:"repository"`
	// Version is the Helm chart version identifier.
	Version string `json:"version"`
	// Name is the Helm chart name within the repository.
	Name string `json:"name"`
}

// ChartMetaStatus contains metadata to a deployed Helm chart.
type ChartMetaStatus struct {
	ChartMeta  `json:",inline"`
	ModifiedAt *metav1.Time `json:"modifiedAt,omitempty"`
}
