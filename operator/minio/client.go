package minio

import (
	"context"
	"github.com/minio/madmin-go"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/s3"
	"k8s.io/apimachinery/pkg/util/rand"
)

type provider struct {
}

func NewMinioProvider() s3.Provider {
	return &provider{}
}

func (p *provider) CreateObjectUser(ctx context.Context, config *v1alpha1.PostgresqlStandaloneOperatorConfig, instance *v1alpha1.PostgresqlStandalone) (*s3.ObjectUser, error) {
	//endpoint := "minio-server.minio-system.svc.cluster.local:9000"
	endpoint := "localhost:9000"
	accessKey := "bszB5aCNYxYCboWR5uq7"
	secretKey := "rdRjzQ3BFodxdXWnAKstGyRyIYDVqWTbphoHQQao"

	mdmClient, err := madmin.New(endpoint, accessKey, secretKey, false)
	if err != nil {
		return nil, err
	}

	userAccessKey := instance.Status.HelmChart.DeploymentNamespace
	userSecretKey := rand.String(40)

	err = mdmClient.AddUser(ctx, userAccessKey, userSecretKey)
	if err != nil {
		return nil, err
	}
	return &s3.ObjectUser{
		Endpoint:  endpoint,
		AccessKey: userAccessKey,
		SecretKey: userSecretKey,
	}, nil
}

func (p *provider) CreateBucket(ctx context.Context, user *s3.ObjectUser, instance *v1alpha1.PostgresqlStandalone) (*s3.Bucket, error) {

	client, err := minio.New(user.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(user.AccessKey, user.SecretKey, ""),
	})
	if err != nil {
		return nil, err
	}

	bucketName := instance.Status.HelmChart.DeploymentNamespace
	location := "use-east-1"

	bucketInfo := &s3.Bucket{
		ObjectUser: *user,
		Name:       bucketName,
	}

	err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region: location,
	})
	if err != nil {
		// Check to see if we already own this bucket (which happens if we run this twice)
		exists, errBucketExists := client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			return bucketInfo, nil
		} else {
			return nil, err
		}
	}

	return bucketInfo, err
}

func (p *provider) DeleteBucket(ctx context.Context, bucket *s3.Bucket) error {
	client, err := minio.New(bucket.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(bucket.AccessKey, bucket.SecretKey, ""),
	})
	if err != nil {
		return err
	}
	return client.RemoveBucketWithOptions(ctx, bucket.Name, minio.RemoveBucketOptions{ForceDelete: true})
}

func (p *provider) DeleteObjectUser(ctx context.Context, user *s3.ObjectUser) error {
	//endpoint := "minio-server.minio-system.svc.cluster.local:9000"
	endpoint := "localhost:9000"
	accessKey := "bszB5aCNYxYCboWR5uq7"
	secretKey := "rdRjzQ3BFodxdXWnAKstGyRyIYDVqWTbphoHQQao"

	mdmClient, err := madmin.New(endpoint, accessKey, secretKey, false)
	if err != nil {
		return err
	}

	return mdmClient.RemoveUser(ctx, user.AccessKey)
}
