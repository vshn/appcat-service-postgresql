//go:build integration

package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/helmvalues"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

type HelmReleaseSuite struct {
	operatortest.Suite
}

func TestHelmRelease(t *testing.T) {
	suite.Run(t, new(HelmReleaseSuite))
}

func (ts *HelmReleaseSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	SetClientInContext(ts.Context, ts.Client)
	ts.RegisterScheme(helmv1beta1.SchemeBuilder.AddToScheme)
}

func (ts *HelmReleaseSuite) Test_EnsureHelmRelease() {
	tests := map[string]struct {
		prepare             func(releaseName string)
		givenReleaseName    string
		givenTemplateValues helmvalues.V
		expectedExtraValues helmvalues.V // a basic set of values is merged
	}{
		"GivenNewHelmRelease_WhenCreating_ThenExpectValuesFromTemplate": {
			givenReleaseName:    "create-release",
			prepare:             func(releaseName string) {},
			givenTemplateValues: helmvalues.V{"key": "value"},
			expectedExtraValues: helmvalues.V{"key": "value"},
		},
		"GivenExistingHelmRelease_WhenUpdating_ThenExpectMergedValuesFromExisting": {
			givenReleaseName:    "update-release",
			givenTemplateValues: helmvalues.V{"key": "template"},
			prepare: func(releaseName string) {
				release := &helmv1beta1.Release{
					ObjectMeta: metav1.ObjectMeta{Name: releaseName},
					Spec: helmv1beta1.ReleaseSpec{
						ForProvider: helmv1beta1.ReleaseParameters{
							ValuesSpec: helmv1beta1.ValuesSpec{
								Values: helmvalues.MustMarshal(helmvalues.V{"key": "existing"})}}}}
				ts.EnsureResources(release)
			},
			expectedExtraValues: helmvalues.V{"key": "existing"},
		},
	}

	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			deploymentNamespace := tc.givenReleaseName
			instance := NewInstanceBuilder("instance", "my-app").getInstance()
			config := newPostgresqlStandaloneOperatorConfig("config", "postgresql-system")
			config.Spec.HelmReleaseTemplate = &v1alpha1.HelmReleaseConfig{
				Values: helmvalues.MustMarshal(tc.givenTemplateValues),
			}

			pipeline.StoreInContext(ts.Context, DeploymentNamespaceKey{}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: deploymentNamespace}})
			SetInstanceInContext(ts.Context, instance)
			pipeline.StoreInContext(ts.Context, ConfigKey{}, config)
			tc.prepare(tc.givenReleaseName)

			// Act
			err := EnsureHelmReleaseFn(labels.Set{
				"test": "label",
			})(ts.Context)
			ts.Require().NoError(err)

			// Assert
			result := &helmv1beta1.Release{}
			ts.FetchResource(types.NamespacedName{Name: deploymentNamespace}, result)

			helmvalues.Merge(testValues, &tc.expectedExtraValues)

			ts.Assert().Equal(deploymentNamespace, result.Spec.ForProvider.Namespace, "target namespace")
			ts.Assert().Equal(deploymentNamespace, result.Name, "metadata.name")
			ts.Assert().Equal("label", result.Labels["test"])
			ts.Assert().JSONEq(string(helmvalues.MustMarshal(tc.expectedExtraValues).Raw), string(result.Spec.ForProvider.Values.Raw))
		})
	}
}

func (ts *HelmReleaseSuite) Test_EnrichStatus() {
	// Arrange
	instance := newInstance("enrich-status", "my-app")
	deploymentNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "enrich-status"}}
	helmRelease := &helmv1beta1.Release{Spec: helmv1beta1.ReleaseSpec{ForProvider: helmv1beta1.ReleaseParameters{
		Namespace: deploymentNamespace.Name,
		Chart:     helmv1beta1.ChartSpec{Name: "postgres", Repository: "https://host/path", Version: "version"},
	}}}
	SetInstanceInContext(ts.Context, instance)
	pipeline.StoreInContext(ts.Context, DeploymentNamespaceKey{}, deploymentNamespace)
	pipeline.StoreInContext(ts.Context, HelmReleaseKey{}, helmRelease)

	ts.EnsureNS(instance.Namespace)
	ts.EnsureResources(instance)

	// Act
	err := EnrichStatusWithHelmChartMetaFn()(ts.Context)
	ts.Require().NoError(err)

	// Assert
	result := &v1alpha1.PostgresqlStandalone{}
	chart := helmRelease.Spec.ForProvider.Chart
	ts.FetchResource(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, result)
	ts.Assert().Equal(v1alpha1.StrategyHelmChart, result.Status.DeploymentStrategy, "deployment strategy")
	ts.Assert().Equal(chart.Name, result.Status.HelmChart.Name, "helm chart name")
	ts.Assert().Equal(chart.Repository, result.Status.HelmChart.Repository, "helm chart repo")
	ts.Assert().Equal(chart.Version, result.Status.HelmChart.Version, "helm chart version")
	ts.Assert().Equal(deploymentNamespace.Name, result.Status.HelmChart.DeploymentNamespace, "deployment namespace")
}

func (ts *HelmReleaseSuite) Test_DeleteHelmRelease() {
	tests := map[string]struct {
		prepare          func(releaseNameString string)
		givenReleaseName string
	}{
		"GivenNonExistingHelmRelease_WhenDeleting_ThenExpectNoError": {
			prepare:          func(releaseName string) {},
			givenReleaseName: "postgresql-release",
		},
		"GivenAnExistingHelmRelease_WhenHelmReleaseStillExists_ThenExpectNoError": {
			prepare: func(releaseName string) {
				release := &helmv1beta1.Release{
					ObjectMeta: metav1.ObjectMeta{Name: "release-existing"},
				}
				ts.EnsureResources(release)
			},
			givenReleaseName: "postgresql-release",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			instance := NewInstanceBuilder("instance", "namespace").setDeploymentNamespace(tc.givenReleaseName).getInstance()
			SetInstanceInContext(ts.Context, instance)
			tc.prepare(tc.givenReleaseName)

			// Act
			err := DeleteHelmReleaseFn()(ts.Context)
			ts.Require().NoError(err)

			// Assert
			resultRelease := &helmv1beta1.Release{}
			err = ts.Client.Get(
				ts.Context,
				client.ObjectKey{Name: tc.givenReleaseName},
				resultRelease,
			)
			AssertResourceNotExists(ts.T(), resultRelease.GetDeletionTimestamp(), err)
		})
	}
}
