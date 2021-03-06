package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PostgresqlStandaloneParameters defines the PostgreSQL specific settings.
type PostgresqlStandaloneParameters struct {

	// Resources contain the storage and compute resources.
	Resources Resources `json:"resources,omitempty"`

	//+kubebuilder:validation:Enum=v14
	//+kubebuilder:default=v14

	// MajorVersion is the supported major version of PostgreSQL.
	//
	// A version cannot be downgraded.
	// Once bumped to the next version, an upgrade process is started in the background.
	// During the upgrade the instance remains in maintenance mode until the upgrade went through successfully.
	MajorVersion MajorVersion `json:"majorVersion,omitempty"`

	//+kubebuilder:default=false

	// EnableSuperUser also provisions the 'postgres' superuser credentials for consumption.
	EnableSuperUser bool `json:"enableSuperUser,omitempty"`
}

// PostgresqlStandaloneSpec defines the desired state of a PostgresqlStandalone.
type PostgresqlStandaloneSpec struct {
	ConnectableInstance   `json:",inline"`
	BackupEnabledInstance `json:",inline"`

	// Parameters defines the PostgreSQL specific settings.
	Parameters PostgresqlStandaloneParameters `json:"forInstance,omitempty"`
}

// PostgresqlStandaloneStatus represents the observed state of a PostgresqlStandalone.
type PostgresqlStandaloneStatus struct {
	GenerationStatus                `json:",inline"`
	Conditions                      []metav1.Condition `json:"conditions,omitempty"`
	PostgresqlStandaloneObservation `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="Progressing",type="string",JSONPath=".status.conditions[?(@.type=='Progressing')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={appcat,postgresql}
// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-postgresql-appcat-vshn-io-v1alpha1-postgresqlstandalone,mutating=false,failurePolicy=fail,groups=postgresql.appcat.vshn.io,resources=postgresqlstandalones,versions=v1alpha1,name=postgresqlstandalones.postgresql.appcat.vshn.io,sideEffects=None,admissionReviewVersions=v1
// +kubebuilder:webhook:verbs=create;update,path=/mutate-postgresql-appcat-vshn-io-v1alpha1-postgresqlstandalone,mutating=true,failurePolicy=fail,groups=postgresql.appcat.vshn.io,resources=postgresqlstandalones,versions=v1alpha1,name=postgresqlstandalones.postgresql.appcat.vshn.io,sideEffects=None,admissionReviewVersions=v1

// PostgresqlStandalone is the user-facing and consumer-friendly API that abstracts the provisioning of standalone Postgresql service instances.
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
	PostgresqlStandaloneKind             = reflect.TypeOf(PostgresqlStandalone{}).Name()
	PostgresqlStandaloneGroupKind        = schema.GroupKind{Group: Group, Kind: PostgresqlStandaloneKind}.String()
	PostgresqlStandaloneKindAPIVersion   = PostgresqlStandaloneKind + "." + SchemeGroupVersion.String()
	PostgresqlStandaloneGroupVersionKind = SchemeGroupVersion.WithKind(PostgresqlStandaloneKind)
)

func init() {
	SchemeBuilder.Register(&PostgresqlStandalone{}, &PostgresqlStandaloneList{})
}
