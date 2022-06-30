//go:build integration

package standalone

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/helmvalues"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

type PipelineSuite struct {
	operatortest.Suite
}

func TestPipeline(t *testing.T) {
	suite.Run(t, new(PipelineSuite))
}

func (ts *PipelineSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	setClientInContext(ts.Context, ts.Client)
	ts.RegisterScheme(helmv1beta1.SchemeBuilder.AddToScheme)
	ts.RegisterScheme(k8upv1.SchemeBuilder.AddToScheme)
}

func (ts *PipelineSuite) Test_FetchOperatorConfig() {
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
			setClientInContext(ts.Context, ts.Client)
			setInstanceInContext(ts.Context, newInstance("instance", "my-app"))
			tc.prepare()
			err := fetchOperatorConfigF(tc.givenNamespace)(ts.Context)
			if tc.expectedError != "" {
				ts.Require().EqualError(err, tc.expectedError)
				return
			}
			ts.Assert().NoError(err)
		})
	}
}

func (ts *PipelineSuite) Test_EnsureHelmRelease() {
	// Arrange
	deploymentNamespace := "ensure-helm-release"
	instance := newInstanceBuilder("instance", "my-app").setDeploymentNamespace(deploymentNamespace).getInstance()
	setInstanceInContext(ts.Context, instance)
	setClientInContext(ts.Context, ts.Client)
	setHelmValuesInContext(ts.Context, helmvalues.V{"key": "value"})
	setHelmChartInContext(ts.Context, &v1alpha1.ChartMeta{Repository: "https://host/path", Version: "version", Name: "postgres"})
	setConfigInContext(ts.Context, newPostgresqlStandaloneOperatorConfig("config", "postgresql-system"))

	// Act
	err := ensureHelmRelease(ts.Context)
	ts.Require().NoError(err)

	// Assert
	result := &helmv1beta1.Release{}
	ts.FetchResource(types.NamespacedName{Name: deploymentNamespace}, result)
	ts.Assert().Equal(result.Spec.ForProvider.Namespace, deploymentNamespace, "target namespace")
	ts.Assert().JSONEq(`{"key":"value"}`, string(result.Spec.ForProvider.Values.Raw))
}

func (ts *PipelineSuite) Test_FetchHelmRelease() {
	// Arrange
	instance := newInstance("fetch-release", "my-app")
	setInstanceInContext(ts.Context, instance)
	setClientInContext(ts.Context, ts.Client)
	instance.Status.HelmChart = &v1alpha1.ChartMetaStatus{
		DeploymentNamespace: generateClusterScopedNameForInstance(),
	}
	helmRelease := &helmv1beta1.Release{
		ObjectMeta: metav1.ObjectMeta{Name: instance.Status.HelmChart.DeploymentNamespace},
	}
	ts.EnsureResources(helmRelease)

	// Act
	err := fetchHelmRelease(ts.Context)
	ts.Require().NoError(err)

	// Assert
	ts.Assert().Equal(helmRelease, getHelmReleaseFromContext(ts.Context))
}
