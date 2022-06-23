package v1alpha1

import corev1 "k8s.io/api/core/v1"

// BackupEnabledInstance is the composable type for enabling instance backups.
type BackupEnabledInstance struct {
	// Backup configures the settings related to backing up the instance.
	Backup BackupSpec `json:"backup,omitempty"`
}

// BackupSpec contains the backup settings.
type BackupSpec struct {
	// Enabled configures whether instances are generally being backed up.
	Enabled bool `json:"enabled,omitempty"`
}

// BackupConfigSpec contains settings for configuring backups for all instances.
type BackupConfigSpec struct {
	// S3BucketSecret configures the bucket settings for backup buckets.
	S3BucketSecret S3BucketConfigSpec `json:"s3BucketSecret,omitempty"`
}

// S3BucketConfigSpec contains references to configure bucket properties.
type S3BucketConfigSpec struct {
	// EndpointRef selects the secret and key for retrieving the endpoint name.
	EndpointRef corev1.SecretKeySelector `json:"endpointRef,omitempty"`
	// BucketRef selects the secret and key for retrieving the bucket name.
	BucketRef corev1.SecretKeySelector `json:"bucketRef,omitempty"`
	// AccessKeyRef selects the access key credential for the bucket.
	AccessKeyRef corev1.SecretKeySelector `json:"accessKeyRef,omitempty"`
	// SecretKeyRef selects the secret key credential for the bucket.
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}
