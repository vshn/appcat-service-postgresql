//go:build integration

package standalone

import (
	"context"
	"fmt"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	"math/rand"
	"strings"
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
	tests := map[string]struct {
		prepare                 func()
		givenBackup             bool
		givenInstance           *v1alpha1.PostgresqlStandalone
		givenS3BucketSecret     *corev1.Secret
		givenConfig             *v1alpha1.PostgresqlStandaloneOperatorConfig
		expectedAccessKeyRef    *corev1.SecretKeySelector
		expectedSecretKeyRef    *corev1.SecretKeySelector
		expectedBucket          string
		expectedEndpoint        string
		expectedResticKey       string
		expectedResticRef       string
		expectedBackupFrequency string
		expectedOtherFrequency  string
		expectedError           string
	}{
		"GivenPostgresqlStandaloneCRD_WhenBackupTrue_ThenExpectSchedule": {
			prepare: func() {
				ts.EnsureNS("scheduled-namespace")
			},
			givenBackup: true,
			givenInstance: newBuilderInstance("instance", "postgresql-instance").
				setDeploymentNamespace("scheduled-namespace").
				setBackup(true).
				get(),
			givenS3BucketSecret: &corev1.Secret{
				Data: map[string][]byte{
					"accessKey": []byte("access"),
					"secretKey": []byte("secret"),
					"bucket":    []byte("k8up-bucket"),
					"endpoint":  []byte("http://minio:9000"),
				},
			},
			givenConfig: &v1alpha1.PostgresqlStandaloneOperatorConfig{
				Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
					BackupConfigSpec: v1alpha1.BackupConfigSpec{
						S3BucketSecret: v1alpha1.S3BucketConfigSpec{
							AccessKeyRef: corev1.SecretKeySelector{
								Key:                  "accessKey",
								LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"},
							},
							SecretKeyRef: corev1.SecretKeySelector{
								Key:                  "secretKey",
								LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"},
							},
							BucketRef: corev1.SecretKeySelector{
								Key:                  "bucket",
								LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"},
							},
							EndpointRef: corev1.SecretKeySelector{
								Key:                  "endpoint",
								LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"},
							},
						},
					},
				},
			},
			expectedAccessKeyRef: &corev1.SecretKeySelector{
				Key:                  "accessKey",
				LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"}},
			expectedSecretKeyRef: &corev1.SecretKeySelector{
				Key:                  "secretKey",
				LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"}},
			expectedBucket:          "k8up-bucket",
			expectedEndpoint:        "http://minio:9000",
			expectedResticKey:       "repository",
			expectedResticRef:       "postgresql-restic",
			expectedBackupFrequency: "@daily-random",
			expectedOtherFrequency:  "@weekly-random",
			expectedError:           "",
		},
		"GivenPostgresqlStandaloneCRD_WhenBackupFalse_ThenExpectNoSchedule": {
			prepare: func() {
				ts.EnsureNS("scheduled-namespace")
			},
			givenBackup: false,
			givenInstance: newBuilderInstance("instance", "postgresql-instance").
				setDeploymentNamespace("scheduled-namespace").
				setBackup(false).
				get(),
			givenS3BucketSecret: nil,
			givenConfig:         nil,
			expectedError:       "schedules.k8up.io \"postgresql\" not found",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// given
			p := &CreateStandalonePipeline{
				operatorNamespace: "operator-namespace",
				client:            ts.Client,
				instance:          tc.givenInstance,
				s3BucketSecret:    tc.givenS3BucketSecret,
				config:            tc.givenConfig,
			}
			tc.prepare()

			// act
			err := p.ensureK8upSchedule(ts.Context)

			// assert
			ts.Assert().NoError(err)

			actualSchedule := k8upv1.Schedule{}
			err = ts.Client.Get(ts.Context, types.NamespacedName{Name: "postgresql", Namespace: "scheduled-namespace"}, &actualSchedule)

			if tc.givenBackup == false {
				ts.Assert().EqualError(err, tc.expectedError)
				return
			}

			ts.Assert().Equal(actualSchedule.Spec.Backup.ScheduleCommon.Schedule.String(), tc.expectedBackupFrequency)
			ts.Assert().Equal(actualSchedule.Spec.Archive.ScheduleCommon.Schedule.String(), tc.expectedOtherFrequency)
			ts.Assert().Equal(actualSchedule.Spec.Prune.ScheduleCommon.Schedule.String(), tc.expectedOtherFrequency)
			ts.Assert().Equal(actualSchedule.Spec.Check.ScheduleCommon.Schedule.String(), tc.expectedOtherFrequency)
			ts.Assert().Equal(actualSchedule.Spec.Backend.RepoPasswordSecretRef.Key, tc.expectedResticKey)
			ts.Assert().Equal(actualSchedule.Spec.Backend.RepoPasswordSecretRef.LocalObjectReference.Name, tc.expectedResticRef)
			ts.Assert().Equal(actualSchedule.Spec.Backend.S3.Endpoint, tc.expectedEndpoint)
			ts.Assert().Equal(actualSchedule.Spec.Backend.S3.Bucket, tc.expectedBucket)
			ts.Assert().Equal(actualSchedule.Spec.Backend.S3.AccessKeyIDSecretRef, tc.expectedAccessKeyRef)
			ts.Assert().Equal(actualSchedule.Spec.Backend.S3.SecretAccessKeySecretRef, tc.expectedSecretKeyRef)
		})
	}
}

func (ts *CreateStandalonePipelineSuite) Test_EnsureResticRepositorySecret() {
	tests := map[string]struct {
		prepare                  func(deplymentNamespace string)
		givenInstance            *v1alpha1.PostgresqlStandalone
		givenDeploymentNamespace string
		expectedName             string
		expectedLabels           map[string]string
		expectedDataKey          string
		expectedError            string
	}{
		"GivenPostgresqlStandaloneCRD_WhenEnsureResticRepositorySecret_ThenExpectSecret": {
			prepare: func(deploymentNamespace string) {
				ts.EnsureNS(deploymentNamespace)
			},
			givenInstance:            newBuilderInstance("instance", "postgresql-instance").get(),
			givenDeploymentNamespace: "instance-namespace",
			expectedName:             "postgresql-restic",
			expectedLabels: map[string]string{
				"app.kubernetes.io/instance":   "instance",
				"app.kubernetes.io/managed-by": v1alpha1.Group,
				"app.kubernetes.io/created-by": fmt.Sprintf("controller-%s", strings.ToLower(v1alpha1.PostgresqlStandaloneKind)),
			},
			expectedDataKey: "repository",
			expectedError:   "",
		},
		"GivenPostgresqlStandaloneCRDWithoutDeploymentNamespace_WhenEnsureResticRepositorySecret_ThenExpectError": {
			prepare: func(deploymentNamespace string) {
				ts.EnsureNS("another-namespace")
			},
			givenInstance:            newBuilderInstance("instance", "postgresql-instance").get(),
			givenDeploymentNamespace: "",
			expectedName:             "",
			expectedLabels:           map[string]string{},
			expectedDataKey:          "",
			expectedError:            "an empty namespace may not be set when a resource name is provided",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// given
			p := &CreateStandalonePipeline{
				operatorNamespace:   "operator-namespace",
				client:              ts.Client,
				instance:            tc.givenInstance,
				deploymentNamespace: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: tc.givenDeploymentNamespace}},
			}
			setInstanceInContext(ts.Context, tc.givenInstance)
			tc.prepare(tc.givenDeploymentNamespace)

			// act
			err := p.ensureResticRepositorySecret(ts.Context)

			// assert
			if tc.expectedError != "" {
				ts.Assert().EqualError(err, tc.expectedError)
				return
			}

			ts.Assert().NoError(err)
			actualSecret := corev1.Secret{}
			err = ts.Client.Get(ts.Context, types.NamespacedName{Name: tc.expectedName, Namespace: tc.givenDeploymentNamespace}, &actualSecret)

			ts.Assert().NoError(err)
			ts.Assert().Equal(tc.expectedLabels, actualSecret.Labels)
			ts.Assert().Contains(actualSecret.Data, tc.expectedDataKey)
			ts.Assert().NotEmpty(actualSecret.Data[tc.expectedDataKey])
		})
	}
}

func (ts *CreateStandalonePipelineSuite) Test_FetchS3BucketSecret() {
	tests := map[string]struct {
		prepare                  func(deplymentNamespace string)
		givenInstance            *v1alpha1.PostgresqlStandalone
		givenConfig              *v1alpha1.PostgresqlStandaloneOperatorConfig
		givenDeploymentNamespace string
		expectedName             string
		expectedBucket           []uint8
		expectedError            string
	}{
		"GivenPostgresqlStandaloneCRD_WhenFetchS3BucketSecret_ThenExpectSecret": {
			prepare: func(deploymentNamespace string) {
				ts.EnsureNS(deploymentNamespace)
				ts.EnsureResources(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: "s3-credentials", Namespace: deploymentNamespace},
					Data: map[string][]byte{
						"bucket": []byte("some-bucket"),
					},
				})
			},
			givenInstance: newBuilderInstance("instance", "postgresql-instance").
				setDeploymentNamespace("secret-namespace").
				get(),
			givenDeploymentNamespace: "secret-namespace",
			givenConfig: &v1alpha1.PostgresqlStandaloneOperatorConfig{
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
			},
			expectedName:   "s3-credentials",
			expectedBucket: []uint8("some-bucket"),
			expectedError:  "",
		},
		"GivenPostgresqlStandaloneCRD_WhenFetchS3BucketMissingSecret_ThenExpectError": {
			prepare: func(deploymentNamespace string) {
				ts.EnsureNS(deploymentNamespace)
			},
			givenInstance: newBuilderInstance("instance", "postgresql-instance").
				setDeploymentNamespace("secret-namespace").
				get(),
			givenDeploymentNamespace: "secret-namespace",
			givenConfig: &v1alpha1.PostgresqlStandaloneOperatorConfig{
				Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
					BackupConfigSpec: v1alpha1.BackupConfigSpec{
						S3BucketSecret: v1alpha1.S3BucketConfigSpec{
							BucketRef: corev1.SecretKeySelector{
								Key:                  "bucket",
								LocalObjectReference: corev1.LocalObjectReference{Name: "missing-s3-credentials"},
							},
						},
					},
				},
			},
			expectedName:   "",
			expectedBucket: []uint8(""),
			expectedError:  "secrets \"missing-s3-credentials\" not found",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// given
			p := &CreateStandalonePipeline{
				operatorNamespace: "operator-namespace",
				client:            ts.Client,
				instance:          tc.givenInstance,
				config:            tc.givenConfig,
			}
			tc.prepare(tc.givenDeploymentNamespace)

			// act
			err := p.fetchS3BucketSecret(ts.Context)

			// assert
			if tc.expectedError != "" {
				ts.Assert().EqualError(err, tc.expectedError)
				return
			}
			ts.Assert().NoError(err)
			actualSecret := corev1.Secret{}
			err = ts.Client.Get(ts.Context, types.NamespacedName{Name: tc.expectedName, Namespace: tc.givenDeploymentNamespace}, &actualSecret)
			ts.Assert().NoError(err)
			ts.Assert().NotEmpty(actualSecret.Data)
			ts.Assert().Equal(tc.expectedBucket, actualSecret.Data["bucket"])
		})
	}
}
