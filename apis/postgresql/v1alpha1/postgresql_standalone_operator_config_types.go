package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// A PostgresqlStandaloneOperatorConfigSpec defines the desired state of a PostgresqlStandaloneOperatorConfig.
type PostgresqlStandaloneOperatorConfigSpec struct {
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

	// HelmProviderConfigReference is the name of the ProviderConfig CR from crossplane-contrib/provider-helm.
	// Used when DeploymentStrategy is StrategyHelmChart.
	HelmProviderConfigReference string `json:"helmProviderConfigReference,omitempty"`

	// Persistence contains default PVC settings.
	Persistence PersistenceSpec `json:"persistence,omitempty"`

	// BackupConfigSpec defines settings for instance backups.
	BackupConfigSpec BackupConfigSpec `json:"backupConfigSpec,omitempty"`
}

// HelmReleaseConfig describes a Helm chart release.
type HelmReleaseConfig struct {
	// Chart sets the scope of this config to a specific version.
	// At least chart version is required in order for this HelmReleaseConfig to take effect.
	Chart ChartMeta `json:"chart,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields

	// Values override PostgresqlStandaloneOperatorConfigSpec.HelmReleaseTemplate.
	// Set MergeValuesFromTemplate to true to deep-merge values instead of replacing them all.
	Values runtime.RawExtension `json:"values,omitempty"`
	// MergeValuesFromTemplate sets the merge behaviour for Values.
	MergeValuesFromTemplate bool `json:"mergeValuesFromTemplate,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced

// A PostgresqlStandaloneOperatorConfig configures a PostgresqlStandalone provider on a cluster level.
// This API isn't meant for consumers.
// It contains defaults and platform-specific configuration values that influence how instances are provisioned.
// There should be a PostgresqlStandaloneOperatorConfig for each major version in use.
type PostgresqlStandaloneOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresqlStandaloneOperatorConfigSpec `json:"spec"`
	Status PostgresqlStandaloneConfigStatus       `json:"status,omitempty"`
}

// PersistenceSpec contains default PVC settings.
type PersistenceSpec struct {
	// storageClassName is the name of the StorageClass required by the claim.
	StorageClassName *string                             `json:"storageClassName,omitempty"`
	AccessModes      []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
}

// +kubebuilder:object:root=true

// PostgresqlStandaloneOperatorConfigList contains a list of PostgresqlStandaloneOperatorConfig.
type PostgresqlStandaloneOperatorConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgresqlStandaloneOperatorConfig `json:"items"`
}

// A PostgresqlStandaloneConfigStatus reflects the observed state of a PostgresqlStandaloneOperatorConfig.
type PostgresqlStandaloneConfigStatus struct {
}

// PostgresqlStandaloneOperatorConfig type metadata.
var (
	PostgresqlStandaloneOperatorConfigKind             = reflect.TypeOf(PostgresqlStandaloneOperatorConfig{}).Name()
	PostgresqlStandaloneOperatorConfigGroupKind        = schema.GroupKind{Group: Group, Kind: PostgresqlStandaloneOperatorConfigKind}.String()
	PostgresqlStandaloneOperatorConfigKindAPIVersion   = PostgresqlStandaloneOperatorConfigKind + "." + SchemeGroupVersion.String()
	PostgresqlStandaloneOperatorConfigGroupVersionKind = SchemeGroupVersion.WithKind(PostgresqlStandaloneOperatorConfigKind)
)

func init() {
	SchemeBuilder.Register(&PostgresqlStandaloneOperatorConfig{}, &PostgresqlStandaloneOperatorConfigList{})
}
