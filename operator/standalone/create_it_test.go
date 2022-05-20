//go:build integration

package standalone

import (
	"context"
	"math/rand"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
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
	ts.RegisterScheme(helmv1beta1.SchemeBuilder.AddToScheme)
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
				instance:          newInstance("instance"),
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
		instance: newInstance("test-ensure-namespace"),
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
		instance:            newInstance("instance"),
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
		instance:   newInstance("instance"),
		client:     ts.Client,
		helmChart:  &v1alpha1.ChartMeta{Repository: "https://host/path", Version: "version", Name: "postgres"},
		helmValues: HelmValues{"key": "value"},
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
