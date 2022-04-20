package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PostgresqlStandaloneParameters are the configurable fields of a PostgresqlStandalone.
type PostgresqlStandaloneParameters struct {
	//+kubebuilder:validation:Enum=Chart
	//+kubebuilder:default=Chart

	// DeploymentStrategy identifies how the Postgresql instance is deployed.
	DeploymentStrategy string `json:"deploymentStrategy,omitempty"`
	// Chart is a DeploymentStrategy that uses Helm chart to deploy PostgresqlStandalone instance.
	Chart                     *ChartMeta `json:"chart,omitempty"`
	BackupEnabledInstance     `json:",inline"`
	MonitoringEnabledInstance `json:",inline"`
	DelayableMaintenance      `json:",inline"`
}

// PostgresqlStandaloneObservation are the observable fields of a PostgresqlStandalone.
type PostgresqlStandaloneObservation struct {
	// Chart is the Helm chart meta that is last observed on a deployed instance.
	Chart *ChartMeta `json:"chart,omitempty"`
}

// A PostgresqlStandaloneSpec defines the desired state of a PostgresqlStandalone.
type PostgresqlStandaloneSpec struct {
	ForProvider PostgresqlStandaloneParameters `json:"forProvider"`
}

// A PostgresqlStandaloneStatus represents the observed state of a PostgresqlStandalone.
type PostgresqlStandaloneStatus struct {
	GenerationStatus `json:",inline"`
	Conditions       []metav1.Condition              `json:"conditions,omitempty"`
	AtProvider       PostgresqlStandaloneObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="Synced",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="External Name",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,appcat,postgresql}
// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-postgresql-appcat-vshn-io-v1alpha1-postgresqlstandalone,mutating=false,failurePolicy=fail,groups=postgresql.appcat.vshn.io,resources=postgresqlstandalones,versions=v1alpha1,name=postgresqlstandalones.postgresql.appcat.vshn.io,sideEffects=None,admissionReviewVersions=v1

// A PostgresqlStandalone is an example API type.
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
