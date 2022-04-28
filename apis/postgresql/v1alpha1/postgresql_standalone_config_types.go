package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// A PostgresqlStandaloneConfigSpec defines the desired state of a PostgresqlStandaloneConfig.
type PostgresqlStandaloneConfigSpec struct {
	// DeploymentStrategy defines the DeploymentStrategy in case there isn't a 1:1 match.
	DeploymentStrategy DeploymentStrategy `json:"defaultDeploymentStrategy,omitempty"`

	// ResourceMinima defines the minimum supported resources an instance can have.
	ResourceMinima Resources `json:"resourceMinima,omitempty"`
	// ResourceMaxima defines the maximum supported resources an instance can have.
	ResourceMaxima Resources `json:"resourceMaxima,omitempty"`

	// HelmReleaseTemplate is the default release config that is used for all HelmReleases.
	// Changing values in this field affects also existing deployed Helm releases unless they are pinned in HelmReleases for a specific chart version.
	// New instances use this config unless there's a specific HelmReleaseConfig for a version that matches the version in this spec.
	HelmReleaseTemplate *HelmReleaseConfig `json:"helmReleaseTemplate,omitempty"`

	// HelmReleases allows to override settings for a specific deployable Helm chart.
	HelmReleases []HelmReleaseConfig `json:"helmReleases,omitempty"`
}

// HelmReleaseConfig describes a Helm chart release.
type HelmReleaseConfig struct {
	// Chart sets the scope of this config to a specific version.
	// At least chart version is required in order for this HelmReleaseConfig to take effect.
	Chart ChartMeta `json:"chart,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields

	// Values override PostgresqlStandaloneConfigSpec.HelmReleaseTemplate.
	// Set MergeValuesFromTemplate to true to deep-merge values instead of replacing them all.
	Values runtime.RawExtension `json:"values,omitempty"`
	// MergeValuesFromTemplate sets the merge behaviour for Values.
	MergeValuesFromTemplate bool `json:"mergeValuesFromTemplate,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced

// A PostgresqlStandaloneConfig configures a PostgresqlStandalone provider on a cluster level.
// This API isn't meant for consumers.
// It contains defaults and platform-specific configuration values that influence how instances are provisioned.
// There should be a PostgresqlStandaloneConfig for each major version in use.
type PostgresqlStandaloneConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresqlStandaloneConfigSpec   `json:"spec"`
	Status PostgresqlStandaloneConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PostgresqlStandaloneConfigList contains a list of PostgresqlStandaloneConfig.
type PostgresqlStandaloneConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgresqlStandaloneConfig `json:"items"`
}

// A PostgresqlStandaloneConfigStatus reflects the observed state of a PostgresqlStandaloneConfig.
type PostgresqlStandaloneConfigStatus struct {
}

// PostgresqlStandaloneConfig type metadata.
var (
	PostgresqlStandaloneConfigKind             = reflect.TypeOf(PostgresqlStandaloneConfig{}).Name()
	PostgresqlStandaloneConfigGroupKind        = schema.GroupKind{Group: Group, Kind: PostgresqlStandaloneConfigKind}.String()
	PostgresqlStandaloneConfigKindAPIVersion   = PostgresqlStandaloneConfigKind + "." + SchemeGroupVersion.String()
	PostgresqlStandaloneConfigGroupVersionKind = SchemeGroupVersion.WithKind(PostgresqlStandaloneConfigKind)
)

func init() {
	SchemeBuilder.Register(&PostgresqlStandaloneConfig{}, &PostgresqlStandaloneConfigList{})
}
