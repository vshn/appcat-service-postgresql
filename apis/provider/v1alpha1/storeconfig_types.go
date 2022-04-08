package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// A StoreConfigSpec defines the desired state of a ProviderConfig.
type StoreConfigSpec struct {
}

// A StoreConfigStatus represents the status of a StoreConfig.
type StoreConfigStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Default-Scope",type="string",JSONPath=".spec.defaultScope"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,store,gcp}
// +kubebuilder:subresource:status

// A StoreConfig configures how GCP controller should store connection details.
type StoreConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoreConfigSpec   `json:"spec"`
	Status StoreConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StoreConfigList contains a list of StoreConfig
type StoreConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StoreConfig `json:"items"`
}

// GetCondition of this StoreConfig.
func (in *StoreConfig) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return in.Status.GetCondition(ct)
}

// SetConditions of this StoreConfig.
func (in *StoreConfig) SetConditions(c ...xpv1.Condition) {
	in.Status.SetConditions(c...)
}

// StoreConfig type metadata.
var (
	StoreConfigKind             = reflect.TypeOf(StoreConfig{}).Name()
	StoreConfigGroupKind        = schema.GroupKind{Group: Group, Kind: StoreConfigKind}.String()
	StoreConfigKindAPIVersion   = StoreConfigKind + "." + SchemeGroupVersion.String()
	StoreConfigGroupVersionKind = SchemeGroupVersion.WithKind(StoreConfigKind)
)

func init() {
	SchemeBuilder.Register(&StoreConfig{}, &StoreConfigList{})
}
