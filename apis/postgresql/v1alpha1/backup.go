package v1alpha1

// BackupEnabledInstance is a reusable type meant for API spec composition.
type BackupEnabledInstance struct {
	// Backup controls the backup settings of an instance.
	Backup Backup `json:"backup,omitempty"`
}

// Backup contains the backup settings.
type Backup struct {
	//+kubebuilder:default=true

	// Enabled controls whether to activate backups.
	Enabled bool `json:"enabled"`
}
