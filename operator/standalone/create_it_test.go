//go:build integration

package standalone

import (
	"context"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/helmvalues"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type CreateStandalonePipelineSuite struct {
	operatortest.Suite
}

func TestCreateStandalonePipeline(t *testing.T) {
	suite.Run(t, new(CreateStandalonePipelineSuite))
}

func (ts *CreateStandalonePipelineSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	setClientInContext(ts.Context, ts.Client)
	ts.RegisterScheme(helmv1beta1.SchemeBuilder.AddToScheme)
	ts.RegisterScheme(k8upv1.SchemeBuilder.AddToScheme)
}

func (ts *CreateStandalonePipelineSuite) Test_FetchOperatorConfig() {
	tests := map[string]struct {
		prepare            func()
		givenNamespace     string
		expectedConfigName string
		expectedError      string
	}{
		"GivenNoExistingConfig_WhenFetching_ThenExpectError": {
			prepare:        func() {},
			givenNamespace: "nonexisting",
			expectedError:  "no PostgresqlStandaloneOperatorConfig found with label 'map[postgresql.appcat.vshn.io/major-version:v14]' in namespace 'nonexisting'",
		},
		"GivenExistingConfig_WhenLabelsMatch_ThenExpectSingleEntry": {
			prepare: func() {
				ts.EnsureNS("single-entry")
				cfg := newPostgresqlStandaloneOperatorConfig("config", "single-entry")
				cfg.Labels = map[string]string{
					v1alpha1.PostgresqlMajorVersionLabelKey: v1alpha1.PostgresqlVersion14.String(),
				}
				ts.EnsureResources(cfg)
			},
			givenNamespace:     "single-entry",
			expectedConfigName: "config",
		},
		"GivenMultipleExistingConfigs_WhenLabelsMatch_ThenExpectError": {
			prepare: func() {
				ts.EnsureNS("multiple-entries")
				cfg1 := newPostgresqlStandaloneOperatorConfig("first", "multiple-entries")
				cfg1.Labels = map[string]string{
					v1alpha1.PostgresqlMajorVersionLabelKey: v1alpha1.PostgresqlVersion14.String(),
				}
				cfg2 := newPostgresqlStandaloneOperatorConfig("second", "multiple-entries")
				cfg2.Labels = cfg1.Labels
				ts.EnsureResources(cfg1, cfg2)
			},
			givenNamespace: "multiple-entries",
			expectedError:  "multiple versions of PostgresqlStandaloneOperatorConfig found with label 'map[postgresql.appcat.vshn.io/major-version:v14]' in namespace 'multiple-entries'",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			p := &CreateStandalonePipeline{
				operatorNamespace: tc.givenNamespace,
				client:            ts.Client,
				instance:          newInstance("instance", "my-app"),
			}
			tc.prepare()
			err := p.fetchOperatorConfig(ts.Context)
			if tc.expectedError != "" {
				ts.Require().EqualError(err, tc.expectedError)
				ts.Assert().Nil(p.config)
				return
			}
			ts.Assert().NoError(err)
		})
	}
}

func (ts *CreateStandalonePipelineSuite) Test_EnsureDeploymentNamespace() {
	// Arrange
	p := &CreateStandalonePipeline{
		instance: newInstance("test-ensure-namespace", "my-app"),
		client:   ts.Client,
	}
	currentRand := namegeneratorRNG
	defer func() {
		namegeneratorRNG = currentRand
	}()
	namegeneratorRNG = rand.New(rand.NewSource(1))
	// Act
	err := p.ensureDeploymentNamespace(ts.Context)
	ts.Require().NoError(err, "create namespace func")

	// Assert
	ns := &corev1.Namespace{}
	ts.FetchResource(types.NamespacedName{Name: "sv-postgresql-s-merry-vigilante-7b16"}, ns)
	ts.Assert().Equal(ns.Labels["app.kubernetes.io/instance"], p.instance.Name)
	ts.Assert().Equal(ns.Labels["app.kubernetes.io/instance-namespace"], p.instance.Namespace)
}

func (ts *CreateStandalonePipelineSuite) Test_EnsureCredentialSecret() {
	// Arrange
	ns := ServiceNamespacePrefix + "my-app-instance"
	ts.EnsureNS(ns)
	p := &CreateStandalonePipeline{
		instance:            newInstance("instance", "my-app"),
		client:              ts.Client,
		deploymentNamespace: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
	}

	// Act
	err := p.ensureCredentialsSecret(ts.Context)
	ts.Require().NoError(err)

	// Assert
	result := &corev1.Secret{}
	ts.FetchResource(types.NamespacedName{Namespace: ns, Name: "postgresql-credentials"}, result)
	ts.Assert().Equal("instance", result.Labels["app.kubernetes.io/instance"], "instance label")
	// Note: Even though we access "Data", the content is not encoded in base64 in envtest.
	ts.Assert().Len(result.Data["password"], 40, "password length")
}

func (ts *CreateStandalonePipelineSuite) Test_EnsureHelmRelease() {
	// Arrange
	p := &CreateStandalonePipeline{
		instance:   newInstance("instance", "my-app"),
		client:     ts.Client,
		helmChart:  &v1alpha1.ChartMeta{Repository: "https://host/path", Version: "version", Name: "postgres"},
		helmValues: helmvalues.V{"key": "value"},
		config:     newPostgresqlStandaloneOperatorConfig("config", "postgresql-system"),
	}
	targetNs := ServiceNamespacePrefix + "my-app-" + p.instance.Name
	p.deploymentNamespace = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: targetNs}}

	// Act
	err := p.ensureHelmRelease(ts.Context)
	ts.Require().NoError(err)

	// Assert
	result := &helmv1beta1.Release{}
	ts.FetchResource(types.NamespacedName{Name: targetNs}, result)
	ts.Assert().Equal(result.Spec.ForProvider.Namespace, targetNs, "target namespace")
	ts.Assert().JSONEq(`{"key":"value"}`, string(result.Spec.ForProvider.Values.Raw))
}

func (ts *CreateStandalonePipelineSuite) Test_EnrichStatus() {
	// Arrange
	p := &CreateStandalonePipeline{
		instance:            newInstance("enrich-status", "my-app"),
		client:              ts.Client,
		helmChart:           &v1alpha1.ChartMeta{Repository: "https://host/path", Version: "version", Name: "postgres"},
		deploymentNamespace: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: generateClusterScopedNameForInstance()}},
	}
	ts.EnsureNS(p.instance.Namespace)
	ts.EnsureResources(p.instance)

	// Act
	err := p.enrichStatus(ts.Context)
	ts.Require().NoError(err)

	// Assert
	result := &v1alpha1.PostgresqlStandalone{}
	ts.FetchResource(types.NamespacedName{Name: p.instance.Name, Namespace: p.instance.Namespace}, result)
	ts.Assert().Equal(v1alpha1.StrategyHelmChart, result.Status.DeploymentStrategy, "deployment strategy")
	ts.Assert().Equal(p.helmChart.Name, result.Status.HelmChart.Name, "helm chart name")
	ts.Assert().Equal(p.helmChart.Repository, result.Status.HelmChart.Repository, "helm chart repo")
	ts.Assert().Equal(p.helmChart.Version, result.Status.HelmChart.Version, "helm chart version")
	ts.Assert().Equal(p.deploymentNamespace.Name, result.Status.HelmChart.DeploymentNamespace, "deployment namespace")
	ts.Assert().True(result.Status.HelmChart.ModifiedTime.IsZero(), "modification date comes later")
}

func (ts *CreateStandalonePipelineSuite) Test_FetchHelmRelease() {
	// Arrange
	p := &CreateStandalonePipeline{
		instance: newInstance("fetch-release", "my-app"),
		client:   ts.Client,
	}
	p.instance.Status.HelmChart = &v1alpha1.ChartMetaStatus{
		DeploymentNamespace: generateClusterScopedNameForInstance(),
	}
	helmRelease := &helmv1beta1.Release{
		ObjectMeta: metav1.ObjectMeta{Name: p.instance.Status.HelmChart.DeploymentNamespace},
	}
	ts.EnsureResources(helmRelease)

	// Act
	err := p.fetchHelmRelease(ts.Context)
	ts.Require().NoError(err)

	// Assert
	ts.Assert().Equal(helmRelease, p.helmRelease)
}

func (ts *CreateStandalonePipelineSuite) Test_FetchCredentialSecret() {
	// Arrange
	p := CreateStandalonePipeline{
		instance: newInstance("fetch-credentials", "my-app"),
	}
	credentialSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql-credentials",
			Namespace: generateClusterScopedNameForInstance(),
		},
		Data: map[string][]byte{
			"password":          []byte("user-password"),
			"postgres-password": []byte("superuser-password"),
		},
	}
	ts.EnsureNS("my-app")
	ts.EnsureNS(credentialSecret.Namespace)
	ts.EnsureResources(credentialSecret)
	p.instance.Status.HelmChart = &v1alpha1.ChartMetaStatus{DeploymentNamespace: credentialSecret.Namespace}

	// Act
	err := p.fetchCredentialSecret(ts.Context)
	ts.Require().NoError(err)

	// Assert
	ts.Require().Len(p.connectionSecret.Data, 2, "data field")
	ts.Require().Len(p.connectionSecret.StringData, 2, "stringData field")
	ts.Assert().Equal("user-password", string(p.connectionSecret.Data["POSTGRESQL_PASSWORD"]))
	ts.Assert().Equal("superuser-password", string(p.connectionSecret.Data["POSTGRESQL_POSTGRES_PASSWORD"]))
	ts.Assert().Equal("fetch-credentials", p.connectionSecret.StringData["POSTGRESQL_USER"])
	ts.Assert().Equal("fetch-credentials", p.connectionSecret.StringData["POSTGRESQL_DATABASE"])
}

func (ts *CreateStandalonePipelineSuite) Test_FetchService() {
	// Arrange
	p := CreateStandalonePipeline{
		instance: newInstance("fetch-service", "my-app"),
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: "service-ns",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 5432}},
		},
	}
	ts.EnsureNS("my-app")
	ts.EnsureNS(service.Namespace)
	ts.EnsureResources(service)
	p.instance.Status.HelmChart = &v1alpha1.ChartMetaStatus{DeploymentNamespace: service.Namespace}

	// Act
	err := p.fetchService(ts.Context)
	ts.Require().NoError(err)

	// Assert
	ts.Require().Len(p.connectionSecret.StringData, 3, "stringData field")
	ts.Assert().Equal("postgresql.service-ns.svc.cluster.local", p.connectionSecret.StringData["POSTGRESQL_SERVICE_NAME"], "service name")
	ts.Assert().Equal("postgresql://postgresql.service-ns.svc.cluster.local:5432", p.connectionSecret.StringData["POSTGRESQL_SERVICE_URL"], "service url")
	ts.Assert().Equal("5432", p.connectionSecret.StringData["POSTGRESQL_SERVICE_PORT"], "service port")
}

func (ts *CreateStandalonePipelineSuite) Test_EnsureK8upSchedule() {
	type testCase struct {
		prepare       func(testCase)
		givenInstance *v1alpha1.PostgresqlStandalone
	}
	tests := map[string]testCase{
		"GivenScheduleDoesNotExist_WhenCreatingSchedule_ThenExpectNewSchedule": {
			prepare: func(tc testCase) {
				ts.EnsureNS("new-schedule")
			},
			givenInstance: newInstanceBuilder("instance", "postgresql-instance").
				setDeploymentNamespace("new-schedule").
				setBackup(true).
				get(),
		},
		"GiveExistingSchedule_WhenUpdatingSchedule_ThenRevertSpecOfOldSchedule": {
			prepare: func(tc testCase) {
				ts.EnsureNS("existing-schedule")
				existingSchedule := newK8upSchedule(tc.givenInstance)
				existingSchedule.Spec.Archive = &k8upv1.ArchiveSchedule{
					ArchiveSpec: k8upv1.ArchiveSpec{
						RestoreSpec: &k8upv1.RestoreSpec{Tags: []string{"tag"}}, // this should be removed
					},
				}
				ts.EnsureResources(existingSchedule)
			},
			givenInstance: newInstanceBuilder("instance", "postgresql-instance").
				setDeploymentNamespace("existing-schedule").
				setBackup(true).
				get(),
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			p := &CreateStandalonePipeline{
				client:   ts.Client,
				instance: tc.givenInstance,
				s3BucketSecret: &corev1.Secret{
					Data: map[string][]byte{
						"accessKey": []byte("access"),
						"secretKey": []byte("secret"),
						"bucket":    []byte("k8up-bucket"),
						"endpoint":  []byte("http://minio:9000"),
					},
				},
				config: &v1alpha1.PostgresqlStandaloneOperatorConfig{
					Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
						BackupConfigSpec: v1alpha1.BackupConfigSpec{
							S3BucketSecret: v1alpha1.S3BucketConfigSpec{
								AccessKeyRef: corev1.SecretKeySelector{Key: "accessKey", LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"}},
								SecretKeyRef: corev1.SecretKeySelector{Key: "secretKey", LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"}},
								BucketRef:    corev1.SecretKeySelector{Key: "bucket", LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"}},
								EndpointRef:  corev1.SecretKeySelector{Key: "endpoint", LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"}},
							},
						},
					},
				},
			}
			tc.prepare(tc)

			// Act
			err := p.ensureK8upSchedule(ts.Context)
			ts.Require().NoError(err)

			// Assert
			result := &k8upv1.Schedule{}
			err = ts.Client.Get(ts.Context, client.ObjectKeyFromObject(newK8upSchedule(tc.givenInstance)), result)

			ts.Assert().Nil(result.Spec.Archive.RestoreSpec, "restore spec")
			ts.Assert().Equal("@daily-random", result.Spec.Backup.ScheduleCommon.Schedule.String(), "backup schedule")
			ts.Assert().Equal("@weekly-random", result.Spec.Archive.ScheduleCommon.Schedule.String(), "archive schedule")
			ts.Assert().Equal("@weekly-random", result.Spec.Prune.ScheduleCommon.Schedule.String(), "prune schedule")
			ts.Assert().Equal("@weekly-random", result.Spec.Check.ScheduleCommon.Schedule.String(), "check schedule")
			ts.Assert().Equal("repository", result.Spec.Backend.RepoPasswordSecretRef.Key, "repo encryption key ref")
			ts.Assert().Equal(2, *result.Spec.FailedJobsHistoryLimit, "failed jobs history limit")
			ts.Assert().Equal(2, *result.Spec.SuccessfulJobsHistoryLimit, "successful jobs history limit")
			ts.Assert().Equal("postgresql-restic", result.Spec.Backend.RepoPasswordSecretRef.LocalObjectReference.Name, "repo encryption name ref")
			ts.Assert().Equal("http://minio:9000", result.Spec.Backend.S3.Endpoint, "s3 endpoint")
			ts.Assert().Equal("k8up-bucket", result.Spec.Backend.S3.Bucket, "s3 bucket name")
			ts.Assert().Equal("accessKey", result.Spec.Backend.S3.AccessKeyIDSecretRef.Key)
			ts.Assert().Equal("s3-credentials", result.Spec.Backend.S3.AccessKeyIDSecretRef.Name)
			ts.Assert().Equal("secretKey", result.Spec.Backend.S3.SecretAccessKeySecretRef.Key)
			ts.Assert().Equal("s3-credentials", result.Spec.Backend.S3.SecretAccessKeySecretRef.Name)
		})
	}
}

func (ts *CreateStandalonePipelineSuite) Test_EnsureResticRepositorySecret() {
	ts.Run("GivenNonExistingSecret_WhenCreatingSecret_ThenExpectNewGeneratedPassword", func() {
		// Arrange
		p := &CreateStandalonePipeline{
			client:              ts.Client,
			instance:            newInstanceBuilder("instance", "new-restic-repo-secret").get(),
			deploymentNamespace: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "new-restic-repo-secret"}},
		}
		setInstanceInContext(ts.Context, p.instance)
		ts.EnsureNS(p.deploymentNamespace.Name)

		// Act
		err := p.ensureResticRepositorySecret(ts.Context)
		ts.Require().NoError(err)

		// Assert
		result := &corev1.Secret{}
		err = ts.Client.Get(ts.Context, types.NamespacedName{Name: getResticRepositorySecretName(), Namespace: p.deploymentNamespace.Name}, result)
		ts.Require().NoError(err)

		ts.Require().NotEmpty(result.Data, "secret data")
		ts.Assert().NotEmpty(result.Data["repository"], "repo password (generated)")
	})

	ts.Run("GivenExistingSecret_WhenUpdatingSecret_ThenLeaveExistingPasswordUntouched", func() {
		// Arrange
		p := &CreateStandalonePipeline{
			client:              ts.Client,
			instance:            newInstanceBuilder("instance", "existing-restic-repo-secret").get(),
			deploymentNamespace: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "existing-restic-repo-secret"}},
		}
		setInstanceInContext(ts.Context, p.instance)
		ts.EnsureNS(p.deploymentNamespace.Name)
		existingSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: getResticRepositorySecretName(), Namespace: p.deploymentNamespace.Name},
			StringData: map[string]string{
				"repository":  "some-generated-password",
				"foreign-key": "should-be-removed",
			},
		}
		ts.EnsureResources(existingSecret)

		// Act
		err := p.ensureResticRepositorySecret(ts.Context)
		ts.Require().NoError(err)

		// Assert
		result := &corev1.Secret{}
		err = ts.Client.Get(ts.Context, types.NamespacedName{Name: getResticRepositorySecretName(), Namespace: p.deploymentNamespace.Name}, result)
		ts.Require().NoError(err)

		ts.Require().Len(result.Data, 1, "secret data")
		ts.Assert().Equal("some-generated-password", string(result.Data["repository"]), "content unchanged")
		ts.Assert().Equal(v1alpha1.Group, result.ObjectMeta.Labels["app.kubernetes.io/managed-by"], "label")
	})
}

func (ts *CreateStandalonePipelineSuite) Test_FetchS3BucketSecret() {
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

	ts.Run("GivenPostgresqlStandaloneCRD_WhenFetchS3BucketMissingSecret_ThenExpectError", func() {
		p := &CreateStandalonePipeline{
			client: ts.Client,
			instance: newInstanceBuilder("instance", "postgresql-instance").
				setDeploymentNamespace("secret-namespace").
				get(),
			config: operatorConfig,
		}
		err := p.fetchS3BucketSecret(ts.Context)
		ts.Assert().EqualError(err, "secrets \"s3-credentials\" not found")
	})

	ts.Run("GivenExistingSecret_WhenFetchingSecret_ThenReturnSecret", func() {
		// Arrange
		p := &CreateStandalonePipeline{
			client: ts.Client,
			instance: newInstanceBuilder("instance", "postgresql-instance").
				setDeploymentNamespace("secret-namespace").
				get(),
			config: operatorConfig,
		}
		ts.EnsureNS("secret-namespace")
		ts.EnsureResources(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "s3-credentials", Namespace: "secret-namespace"},
			Data: map[string][]byte{
				"bucket": []byte("some-bucket"),
			},
		})

		// Act
		err := p.fetchS3BucketSecret(ts.Context)
		ts.Require().NoError(err)

		// Assert
		result := p.s3BucketSecret
		ts.Assert().NotEmpty(result.Data)
		ts.Assert().Equal("some-bucket", string(result.Data["bucket"]))
	})
}
