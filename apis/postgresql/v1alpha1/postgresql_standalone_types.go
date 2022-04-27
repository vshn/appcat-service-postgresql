package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PostgresqlStandaloneObservation are the observable fields of a PostgresqlStandalone.
type PostgresqlStandaloneObservation struct {
	DeploymentStrategy DeploymentStrategy `json:"deploymentStrategy,omitempty"`
	HelmChart          *ChartMetaStatus   `json:"helmChart,omitempty"`
}

// A PostgresqlStandaloneSpec defines the desired state of a PostgresqlStandalone.
type PostgresqlStandaloneSpec struct {
	BackupEnabledInstance     `json:",inline"`
	MonitoringEnabledInstance `json:",inline"`
	DeferrableMaintenance     `json:",inline"`
	Resources                 Resources `json:"resources,omitempty"`
}

// A PostgresqlStandaloneStatus represents the observed state of a PostgresqlStandalone.
type PostgresqlStandaloneStatus struct {
	GenerationStatus                `json:",inline"`
	Conditions                      []metav1.Condition `json:"conditions,omitempty"`
	PostgresqlStandaloneObservation `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={appcat,postgresql}
// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-postgresql-appcat-vshn-io-v1alpha1-postgresqlstandalone,mutating=false,failurePolicy=fail,groups=postgresql.appcat.vshn.io,resources=postgresqlstandalones,versions=v1alpha1,name=postgresqlstandalones.postgresql.appcat.vshn.io,sideEffects=None,admissionReviewVersions=v1

// A PostgresqlStandalone is the user-facing and consumer-friendly API that abstracts the provisioning of standalone Postgresql service instances.
type PostgresqlStandalone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresqlStandaloneSpec   `json:"spec"`
	Status PostgresqlStandaloneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PostgresqlStandaloneList contains a list of PostgresqlStandalone
type PostgresqlStandaloneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgresqlStandalone `json:"items"`
}

// PostgresqlStandalone type metadata.
var (
	PostgresStandaloneKind             = reflect.TypeOf(PostgresqlStandalone{}).Name()
	PostgresStandaloneGroupKind        = schema.GroupKind{Group: Group, Kind: PostgresStandaloneKind}.String()
	PostgresStandaloneKindAPIVersion   = PostgresStandaloneKind + "." + SchemeGroupVersion.String()
	PostgresStandaloneGroupVersionKind = SchemeGroupVersion.WithKind(PostgresStandaloneKind)
)

func init() {
	SchemeBuilder.Register(&PostgresqlStandalone{}, &PostgresqlStandaloneList{})
}
