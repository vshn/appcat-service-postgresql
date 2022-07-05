//go:build integration

package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

type SecretSuite struct {
	operatortest.Suite
}

func TestSecretSuite(t *testing.T) {
	suite.Run(t, new(SecretSuite))
}

func (ts *SecretSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	SetClientInContext(ts.Context, ts.Client)
}

func (ts *SecretSuite) Test_EnsureResticRepositorySecret() {
	ts.Run("GivenNonExistingSecret_WhenCreatingSecret_ThenExpectNewGeneratedPassword", func() {
		// Arrange
		SetClientInContext(ts.Context, ts.Client)
		deploymentNamespace := "new-restic-repo-secret"
		SetInstanceInContext(ts.Context, NewInstanceBuilder("instance", "new-restic-repo-secret").setDeploymentNamespace(deploymentNamespace).getInstance())
		ts.EnsureNS(deploymentNamespace)

		// Act
		err := EnsureResticRepositorySecretFn(labels.Set{})(ts.Context)
		ts.Require().NoError(err)

		// Assert
		result := &corev1.Secret{}
		err = ts.Client.Get(ts.Context, types.NamespacedName{Name: getResticRepositorySecretName(), Namespace: deploymentNamespace}, result)
		ts.Require().NoError(err)

		ts.Require().NotEmpty(result.Data, "secret data")
		ts.Assert().NotEmpty(result.Data["repository"], "repo password (generated)")
	})

	ts.Run("GivenExistingSecret_WhenUpdatingSecret_ThenLeaveExistingPasswordUntouched", func() {
		// Arrange
		SetClientInContext(ts.Context, ts.Client)
		deploymentNamespace := "existing-restic-repo-secret"
		SetInstanceInContext(ts.Context, NewInstanceBuilder("instance", "existing-restic-repo-secret").setDeploymentNamespace(deploymentNamespace).getInstance())
		ts.EnsureNS(deploymentNamespace)
		existingSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: getResticRepositorySecretName(), Namespace: deploymentNamespace},
			StringData: map[string]string{
				"repository":  "some-generated-password",
				"foreign-key": "should-be-removed",
			},
		}
		ts.EnsureResources(existingSecret)

		// Act
		err := EnsureResticRepositorySecretFn(labels.Set{
			"app.kubernetes.io/managed-by": v1alpha1.Group,
		})(ts.Context)
		ts.Require().NoError(err)

		// Assert
		result := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: getResticRepositorySecretName(), Namespace: deploymentNamespace},
		}
		ts.FetchResource(client.ObjectKeyFromObject(result), result)
		ts.Require().Len(result.Data, 1, "secret data")
		ts.Assert().Equal("some-generated-password", string(result.Data["repository"]), "content unchanged")
		ts.Assert().Equal(v1alpha1.Group, result.ObjectMeta.Labels["app.kubernetes.io/managed-by"], "label")
	})
}

func (ts *SecretSuite) Test_EnsureCredentialSecret() {
	// Arrange
	ns := "credential-secret"
	instance := NewInstanceBuilder("instance", "my-app").setDeploymentNamespace(ns).getInstance()
	SetInstanceInContext(ts.Context, instance)
	pipeline.StoreInContext(ts.Context, DeploymentNamespaceKey{}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})

	ts.EnsureNS(ns)

	// Act
	err := EnsureCredentialsSecretFn(labels.Set{
		"app.kubernetes.io/instance": instance.Name,
	})(ts.Context)
	ts.Require().NoError(err)

	// Assert
	result := &corev1.Secret{}
	ts.FetchResource(types.NamespacedName{Namespace: ns, Name: "postgresql-credentials"}, result)
	ts.Assert().Equal("instance", result.Labels["app.kubernetes.io/instance"], "instance label")
	// Note: Even though we access "Data", the content is not encoded in base64 in envtest.
	ts.Assert().Len(result.Data["password"], 40, "password length")
	ts.Assert().Len(result.Data["postgres-password"], 40, "password length")
	ts.Assert().Len(result.Data["replication-password"], 40, "password length")
}

func (ts *SecretSuite) Test_EnsureConnectionSecret() {
	// Arrange
	ns := "service-ns"
	instance := NewInstanceBuilder("instance", "connection-secret").getInstance()
	SetInstanceInContext(ts.Context, instance)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "postgresql", Namespace: ns},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 5432}},
		},
	}
	credentialSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: getCredentialSecretName(), Namespace: ns},
		StringData: map[string]string{
			"password":          "test",
			"postgres-password": "superuser",
		},
	}
	pipeline.StoreInContext(ts.Context, ServiceKey{}, service)
	pipeline.StoreInContext(ts.Context, CredentialSecretKey{}, credentialSecret)

	ts.EnsureNS(ns)
	ts.EnsureNS("connection-secret")
	ts.EnsureResources(service, credentialSecret, instance)

	// Act
	err := EnsureConnectionSecretFn(labels.Set{"test": "label"})(ts.Context)
	ts.Require().NoError(err)

	// Assert
	result := &corev1.Secret{}
	ts.FetchResource(types.NamespacedName{Name: instance.GetConnectionSecretName(), Namespace: "connection-secret"}, result)

	ts.Require().Len(result.Data, 7, "data field")
	ts.Assert().Equal("label", result.Labels["test"])
	ts.Assert().Equal("postgresql.service-ns.svc.cluster.local", string(result.Data["POSTGRESQL_SERVICE_NAME"]), "service name")
	ts.Assert().Equal("postgresql://postgresql.service-ns.svc.cluster.local:5432", string(result.Data["POSTGRESQL_SERVICE_URL"]), "service url")
	ts.Assert().Equal("5432", string(result.Data["POSTGRESQL_SERVICE_PORT"]), "service port")
	ts.Assert().Equal("test", string(result.Data["POSTGRESQL_PASSWORD"]), "password")
	ts.Assert().Equal("superuser", string(result.Data["POSTGRESQL_POSTGRES_PASSWORD"]), "superuser password")
	ts.Assert().Equal("instance", string(result.Data["POSTGRESQL_DATABASE"]), "database name")
	ts.Assert().Equal("instance", string(result.Data["POSTGRESQL_USER"]), "user name")
	ts.Assert().Equal("instance", result.OwnerReferences[0].Name)

}

func (ts *SecretSuite) Test_FetchS3BucketSecret() {
	// Arrange
	operatorConfig := &v1alpha1.PostgresqlStandaloneOperatorConfig{
		Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
			BackupConfigSpec: v1alpha1.BackupConfigSpec{
				S3BucketSecret: v1alpha1.S3BucketConfigSpec{
					BucketRef: corev1.SecretKeySelector{
						Key:                  "bucket",
						LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"},
					},
				},
			},
		},
	}
	pipeline.StoreInContext(ts.Context, ConfigKey{}, operatorConfig)
	SetInstanceInContext(ts.Context, NewInstanceBuilder("instance", "postgresql-instance").
		setDeploymentNamespace("secret-namespace").
		getInstance())

	ts.Run("GivenPostgresqlStandaloneCRD_WhenFetchS3BucketMissingSecret_ThenExpectError", func() {
		// Act
		err := FetchS3BucketSecretFn()(ts.Context)

		// Assert
		ts.Assert().EqualError(err, "secrets \"s3-credentials\" not found")
	})

	ts.Run("GivenExistingSecret_WhenFetchingSecret_ThenReturnSecret", func() {
		// Arrange
		ts.EnsureNS("secret-namespace")
		ts.EnsureResources(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "s3-credentials", Namespace: "secret-namespace"},
			Data: map[string][]byte{
				"bucket": []byte("some-bucket"),
			},
		})

		// Act
		err := FetchS3BucketSecretFn()(ts.Context)
		ts.Require().NoError(err)

		// Assert
		result := getFromContextOrPanic(ts.Context, BucketSecretKey{}).(*corev1.Secret)
		ts.Assert().NotEmpty(result.Data)
		ts.Assert().Equal("some-bucket", string(result.Data["bucket"]))
	})
}

func (ts *SecretSuite) Test_DeleteConnectionSecret() {
	tests := map[string]struct {
		prepare        func(name, namespace string)
		givenNamespace string
		givenSecret    string
		expectedError  string
	}{
		"GivenNonExistingSecret_WhenDeleting_ThenExpectNoError": {
			prepare:        func(name, namespace string) {},
			givenNamespace: "test-namespace",
			givenSecret:    "non-existing-secret",
			expectedError:  "",
		},
		"GivenExistingSecret_WhenDeleting_ThenExpectNoError": {
			prepare: func(name, namespace string) {
				ts.EnsureNS(namespace)
				ts.EnsureResources(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}})
			},
			givenNamespace: "test-namespace",
			givenSecret:    "existing-secret",
			expectedError:  "",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			instance := NewInstanceBuilder("instance", tc.givenNamespace).getInstance()
			instance.Spec.ConnectableInstance.WriteConnectionSecretToRef.Name = tc.givenSecret
			SetInstanceInContext(ts.Context, instance)
			tc.prepare(tc.givenSecret, tc.givenNamespace)

			// Act
			err := DeleteConnectionSecretFn()(ts.Context)

			// Assert
			if tc.expectedError != "" {
				ts.Assert().EqualError(err, tc.expectedError)
				return
			}
			ts.Require().NoError(err)
			resultSecret := &corev1.Secret{}
			err = ts.Client.Get(
				ts.Context,
				types.NamespacedName{Name: tc.givenSecret, Namespace: tc.givenNamespace},
				resultSecret,
			)
			AssertResourceNotExists(ts.T(), resultSecret.GetDeletionTimestamp(), err)
		})
	}
}
