package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// A PostgresqlStandaloneConfigSpec defines the desired state of a PostgresqlStandaloneConfig.
type PostgresqlStandaloneConfigSpec struct {
}

// A PostgresqlStandaloneConfigStatus reflects the observed state of a PostgresqlStandaloneConfig.
type PostgresqlStandaloneConfigStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster

// A PostgresqlStandaloneConfig configures a PostgresqlStandalone provider.
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
