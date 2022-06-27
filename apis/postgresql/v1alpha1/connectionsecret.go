package v1alpha1

// ConnectableInstance is the composable type for enabling connection secret.
type ConnectableInstance struct {
	WriteConnectionSecretToRef ConnectionSecretRef `json:"writeConnectionSecretToRef,omitempty"`
}

// ConnectionSecretRef contains the reference where connection details should be made available.
type ConnectionSecretRef struct {
	// Name is the Secret name to where the connection details should be written to after creating an instance.
	Name string `json:"name,omitempty"`
}

// GetConnectionSecretName returns the name of the connection secret if set, otherwise it returns `metadata.name`.
func (in *PostgresqlStandalone) GetConnectionSecretName() string {
	if in.Spec.WriteConnectionSecretToRef.Name == "" {
		return in.Name
	}
	return in.Spec.WriteConnectionSecretToRef.Name
}
