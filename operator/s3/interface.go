package s3

import (
	"context"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
)

type ObjectUser struct {
	Endpoint  string
	AccessKey string
	SecretKey string
}

type Bucket struct {
	ObjectUser
	Name string
}

type Provider interface {
	CreateObjectUser(ctx context.Context, config *v1alpha1.PostgresqlStandaloneOperatorConfig, instance *v1alpha1.PostgresqlStandalone) (*ObjectUser, error)
	CreateBucket(ctx context.Context, user *ObjectUser, instance *v1alpha1.PostgresqlStandalone) (*Bucket, error)
	DeleteBucket(ctx context.Context, bucket *Bucket) error
	DeleteObjectUser(ctx context.Context, user *ObjectUser) error
}

var SupportedBucketProviders = map[string]Provider{}
