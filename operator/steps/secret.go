package steps

import (
	"context"
	"fmt"
	pipeline "github.com/ccremer/go-command-pipeline"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// EnsureResticRepositorySecretFn returns a function that creates the restic repository secret required by K8up.
// The password is generated if it's a new resource, otherwise left unchanged.
func EnsureResticRepositorySecretFn(labelSet labels.Set) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)
		deploymentNamespace := instance.Status.HelmChart.DeploymentNamespace

		// NOTE: we should never delete the Restic Repository secret.
		// There could be cases where the user temporarily disables backups and then re-enables.
		// This case shouldn't result in a new encryption password that renders the previously created backups unusable.
		// All this is under assumption that the Bucket is not immediately removed when backups are disabled.
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getResticRepositorySecretName(),
				Namespace: deploymentNamespace,
			},
		}
		secretKey := "repository"
		_, err := controllerutil.CreateOrUpdate(ctx, kube, secret, func() error {
			secret.Labels = labels.Merge(secret.Labels, labelSet)
			if val, exists := secret.Data[secretKey]; exists && len(val) > 0 {
				// password already set, let's reset every other key
				secret.Data = map[string][]byte{
					secretKey: val,
				}
				return nil
			}
			password := generatePassword()
			if val, exists := secret.StringData[secretKey]; exists && val != "" {
				// only generate a password on new object, not existing ones
				password = val
			}
			secret.StringData = map[string]string{
				secretKey: password,
			}
			return nil
		})
		return err
	}
}

// EnsureCredentialsSecretFn creates the secret that contains the PostgreSQL secret.
// Passwords are generated if they are unset.
func EnsureCredentialsSecretFn(labelSet labels.Set) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		deploymentNamespace := getFromContextOrPanic(ctx, DeploymentNamespaceKey{}).(*corev1.Namespace)

		// https://github.com/bitnami/charts/tree/master/bitnami/postgresql
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getCredentialSecretName(),
				Namespace: deploymentNamespace.Name,
			},
		}
		_, err := controllerutil.CreateOrUpdate(ctx, kube, secret, func() error {
			secret.Labels = labels.Merge(secret.Labels, labelSet)
			if secret.Data == nil {
				secret.Data = map[string][]byte{}
			}
			if secret.StringData == nil {
				secret.StringData = map[string]string{}
			}
			for _, key := range []string{"postgres-password", "password", "replication-password"} {
				if _, exists := secret.Data[key]; !exists {
					secret.StringData[key] = generatePassword()
				}
			}
			return nil
		})
		pipeline.StoreInContext(ctx, CredentialSecretKey{}, secret)
		return err
	}
}

// EnsureConnectionSecretFn creates the connection secret in the instance's namespace.
func EnsureConnectionSecretFn(labelSet labels.Set) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)
		service := getFromContextOrPanic(ctx, ServiceKey{}).(*corev1.Service)
		credentialSecret := getFromContextOrPanic(ctx, CredentialSecretKey{}).(*corev1.Secret)

		secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: instance.GetConnectionSecretName(), Namespace: instance.Namespace}}
		_, err := controllerutil.CreateOrUpdate(ctx, kube, secret, func() error {
			secret.Labels = labels.Merge(secret.Labels, labelSet)
			if secret.Data == nil {
				secret.Data = map[string][]byte{}
			}
			if secret.StringData == nil {
				secret.StringData = map[string]string{}
			}
			secret.StringData["POSTGRESQL_SERVICE_NAME"] = fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace)
			secret.StringData["POSTGRESQL_SERVICE_URL"] = fmt.Sprintf("postgresql://%s.%s.svc.cluster.local:%d", service.Name, service.Namespace, service.Spec.Ports[0].Port)
			secret.StringData["POSTGRESQL_SERVICE_PORT"] = fmt.Sprintf("%d", service.Spec.Ports[0].Port)
			if instance.Spec.Parameters.EnableSuperUser {
				secret.Data["POSTGRESQL_POSTGRES_PASSWORD"] = credentialSecret.Data["postgres-password"]
			} else {
				delete(secret.Data, "POSTGRESQL_POSTGRES_PASSWORD")
			}
			secret.Data["POSTGRESQL_PASSWORD"] = credentialSecret.Data["password"]
			secret.StringData["POSTGRESQL_DATABASE"] = instance.Name
			secret.StringData["POSTGRESQL_USER"] = instance.Name
			return controllerutil.SetOwnerReference(instance, secret, kube.Scheme())
		})
		pipeline.StoreInContext(ctx, ConnectionSecretKey{}, secret)
		return err
	}
}

// FetchS3BucketSecretFn fetches a secret that contains the bucket configuration.
// It assumes that there is another provisioner that deploys S3 bucket ready for use.
func FetchS3BucketSecretFn() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)
		config := GetConfigFromContext(ctx)

		s3BucketSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
			Name:      config.Spec.BackupConfigSpec.S3BucketSecret.BucketRef.Name,
			Namespace: instance.Status.HelmChart.DeploymentNamespace,
		}}
		err := kube.Get(ctx, client.ObjectKeyFromObject(s3BucketSecret), s3BucketSecret)
		pipeline.StoreInContext(ctx, BucketSecretKey{}, s3BucketSecret)
		return err
	}
}

// DeleteConnectionSecretFn deletes the connection secret of the PostgreSQL instance.
// Ignores "not found" error.
func DeleteConnectionSecretFn() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)

		connectionSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.GetConnectionSecretName(),
				Namespace: instance.Namespace,
			},
		}
		err := kube.Delete(ctx, connectionSecret)
		return client.IgnoreNotFound(err)
	}
}

func getResticRepositorySecretName() string {
	return fmt.Sprintf("%s-restic", getDeploymentName())
}

func generatePassword() string {
	return utilrand.String(40)
}
