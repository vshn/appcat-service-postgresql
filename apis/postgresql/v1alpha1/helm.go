package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ChartMeta contains the metadata to a Helm chart.
type ChartMeta struct {
	// Repository is the Helm chart repository URL.
	Repository string `json:"repository,omitempty"`
	// Version is the Helm chart version identifier.
	Version string `json:"version,omitempty"`
	// Name is the Helm chart name within the repository.
	Name string `json:"name,omitempty"`
}

// ChartMetaStatus contains metadata to a deployed Helm chart.
type ChartMetaStatus struct {
	ChartMeta `json:",inline"`
	// ModifiedAt is the timestamp when the helm release has been last modified.
	ModifiedAt metav1.Time `json:"modifiedAt,omitempty"`
	// DeploymentNamespace is the observed namespace name where the instance is deployed.
	DeploymentNamespace string `json:"deploymentNamespace,omitempty"`
}
