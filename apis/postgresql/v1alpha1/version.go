package v1alpha1

import "fmt"

// MajorVersion identifies a major version of a service instance.
type MajorVersion string

const (
	// PostgresqlVersion14 identifies PostgreSQL v14.
	PostgresqlVersion14 MajorVersion = "v14"
)

var (
	// PostgresqlMajorVersionLabelKey is the label key to add for selecting major version
	PostgresqlMajorVersionLabelKey = fmt.Sprintf("%s/major-version", Group)
)

// String implements fmt.Stringer.
func (v MajorVersion) String() string {
	return string(v)
}
