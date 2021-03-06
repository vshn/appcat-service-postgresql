// +kubebuilder:object:generate=true
// +groupName=postgresql.appcat.vshn.io
// +versionName=v1alpha1

// Package v1alpha1 contains the v1alpha1 group postgresql.appcat.vshn.io resources of the PostgreSQL provider.
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "postgresql.appcat.vshn.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)
