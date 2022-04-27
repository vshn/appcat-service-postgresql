package v1alpha1

// MajorVersion identifies a major version of a service instance.
type MajorVersion string

const (
	// PostgresqlVersion14 identifies PostgreSQL v14.
	PostgresqlVersion14 MajorVersion = "v14"
)

// String implements fmt.Stringer.
func (v *MajorVersion) String() string {
	if v == nil {
		return ""
	}
	return string(*v)
}
