package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/helmvalues"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
	"time"
)

func TestCompileHelmValues(t *testing.T) {
	// Arrange
	config := &v1alpha1.PostgresqlStandaloneOperatorConfig{Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
		HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
			Values: runtime.RawExtension{Raw: []byte(`{"key":"value"}`)},
			Chart:  v1alpha1.ChartMeta{Repository: "https://host/path", Name: "postgresql", Version: "1.0"},
		},
	}}
	instance := &v1alpha1.PostgresqlStandalone{
		ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "my-app"},
		Spec: v1alpha1.PostgresqlStandaloneSpec{
			Parameters: v1alpha1.PostgresqlStandaloneParameters{
				MajorVersion:    v1alpha1.PostgresqlVersion14,
				EnableSuperUser: true,
				Resources: v1alpha1.Resources{
					ComputeResources: v1alpha1.ComputeResources{
						MemoryLimit: parseResource("2Gi")},
					StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("1Gi")}}}},
	}

	// Act
	values, chart, err := compileHelmValues(config, instance)
	require.NoError(t, err)

	// Assert
	expectedChart := &v1alpha1.ChartMeta{
		Repository: "https://host/path",
		Name:       "postgresql",
		Version:    "1.0",
	}

	expectedValues := helmvalues.V{
		"key": "value",
	}
	helmvalues.Merge(testValues, &expectedValues)
	assert.Equal(t, expectedChart, chart)
	assert.Equal(t, expectedValues, values)
}

func TestOverrideTemplateValues(t *testing.T) {
	tests := map[string]struct {
		givenSpec      v1alpha1.PostgresqlStandaloneOperatorConfigSpec
		expectedValues helmvalues.V
		expectedChart  v1alpha1.ChartMeta
		expectedError  string
	}{
		"GivenSpecificReleaseExists_WhenMergeEnabled_ThenMergeWithExistingValues": {
			givenSpec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
				HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
					Values: helmvalues.MustMarshal(helmvalues.V{"key": "value", "existing": "untouched"}),
					Chart:  v1alpha1.ChartMeta{Version: "version", Name: "postgres", Repository: "url"},
				},
				HelmReleases: []v1alpha1.HelmReleaseConfig{
					{
						MergeValuesFromTemplate: true,
						Chart:                   v1alpha1.ChartMeta{Version: "version"},
						Values:                  helmvalues.MustMarshal(helmvalues.V{"key": map[string]interface{}{"nested": "value"}, "merged": "newValue"}),
					},
				},
			},
			expectedValues: helmvalues.V{
				"key": map[string]interface{}{
					"nested": "value",
				},
				"merged":   "newValue",
				"existing": "untouched",
			},
			expectedChart: v1alpha1.ChartMeta{Repository: "url", Name: "postgres", Version: "version"},
		},
		"GivenSpecificReleaseExists_WhenMergeDisabled_ThenOverwriteExistingValues": {
			givenSpec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
				HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
					Values: helmvalues.MustMarshal(helmvalues.V{"key": "value", "existing": "untouched"}),
					Chart:  v1alpha1.ChartMeta{Version: "version", Name: "postgres", Repository: "url"},
				},
				HelmReleases: []v1alpha1.HelmReleaseConfig{
					{
						Chart:  v1alpha1.ChartMeta{Version: "version", Name: "alternative", Repository: "fork"},
						Values: helmvalues.MustMarshal(helmvalues.V{"key": helmvalues.V{"nested": "value"}, "merged": "newValue"}),
					},
				},
			},
			expectedValues: helmvalues.V{
				"key": map[string]interface{}{
					"nested": "value",
				},
				"merged": "newValue",
			},
			expectedChart: v1alpha1.ChartMeta{Repository: "fork", Name: "alternative", Version: "version"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			vals := helmvalues.V{}
			helmvalues.MustUnmarshal(tc.givenSpec.HelmReleaseTemplate.Values, &vals)
			resultValues, resultChart, err := overrideTemplateValues(&v1alpha1.PostgresqlStandaloneOperatorConfig{Spec: tc.givenSpec}, vals)
			if tc.expectedError != "" {
				require.EqualError(t, err, tc.expectedError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedValues, resultValues)
			assert.Equal(t, &tc.expectedChart, resultChart)
		})
	}
}

var testValues = helmvalues.V{
	"auth": helmvalues.V{
		"existingSecret":     "postgresql-credentials",
		"database":           "instance",
		"enablePostgresUser": true,
		"username":           "instance",
	},
	"primary": helmvalues.V{
		"persistence": helmvalues.V{
			"existingClaim": "postgresql-data",
		},
		"resources": helmvalues.V{
			"limits": helmvalues.V{
				"memory": "2Gi",
			},
		},
		"podAnnotations": helmvalues.V{
			"k8up.io/backupcommand":                      `sh -c 'PGUSER="postgres" PGPASSWORD="$POSTGRES_POSTGRES_PASSWORD" pg_dumpall --clean'`,
			"k8up.io/file-extension":                     ".sql",
			"postgresql.appcat.vshn.io/storage-capacity": "1Gi",
		},
	},
	"fullnameOverride": "postgresql",
	"networkPolicy": helmvalues.V{
		"enabled": true,
		"ingressRules": helmvalues.V{
			"primaryAccessOnlyFrom": helmvalues.V{
				"enabled": true,
				"namespaceSelector": helmvalues.V{
					"kubernetes.io/metadata.name": "my-app",
				},
			},
		},
	},
}

func TestApplyValuesFromInstance(t *testing.T) {
	instance := newInstance("instance", "my-app")
	result := applyValuesFromInstance(instance, helmvalues.V{})
	assert.Equal(t, testValues, result)
}

func TestIsHelmReleaseReady(t *testing.T) {
	// Arrange
	ctx := pipeline.MutableContext(context.Background())
	instance := newInstance("release-ready", "my-app")
	SetInstanceInContext(ctx, instance)
	instance.Status.HelmChart = &v1alpha1.ChartMetaStatus{}

	modifiedDate := metav1.Date(2022, 05, 17, 17, 52, 35, 0, time.Local)
	helmRelease := &helmv1beta1.Release{
		ObjectMeta: metav1.ObjectMeta{Name: "release-ready"},
	}
	pipeline.StoreInContext(ctx, HelmReleaseKey{}, helmRelease)

	t.Run("check non-ready release", func(t *testing.T) {
		// Act
		result := IsHelmReleaseReadyP()(ctx)

		// Assert
		assert.False(t, result)
	})

	t.Run("check ready release", func(t *testing.T) {
		helmRelease.Status = helmv1beta1.ReleaseStatus{
			ResourceStatus: crossplanev1.ResourceStatus{
				ConditionedStatus: crossplanev1.ConditionedStatus{Conditions: []crossplanev1.Condition{
					{
						Type:               crossplanev1.TypeReady,
						Status:             corev1.ConditionTrue,
						LastTransitionTime: metav1.NewTime(modifiedDate.Add(12 * time.Second)),
					},
				}},
			},
			Synced: true,
		}

		// Act
		result := IsHelmReleaseReadyP()(ctx)

		// Assert
		assert.True(t, result)
	})
}

func newInstance(name string, namespace string) *v1alpha1.PostgresqlStandalone {
	return &v1alpha1.PostgresqlStandalone{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Generation: 1},
		Spec: v1alpha1.PostgresqlStandaloneSpec{
			Parameters: v1alpha1.PostgresqlStandaloneParameters{
				MajorVersion:    v1alpha1.PostgresqlVersion14,
				EnableSuperUser: true,
				Resources: v1alpha1.Resources{
					ComputeResources: v1alpha1.ComputeResources{MemoryLimit: parseResource("2Gi")},
					StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("1Gi")},
				},
			},
		},
		Status: v1alpha1.PostgresqlStandaloneStatus{
			PostgresqlStandaloneObservation: v1alpha1.PostgresqlStandaloneObservation{
				HelmChart: &v1alpha1.ChartMetaStatus{},
			},
		},
	}
}

func parseResource(value string) *resource.Quantity {
	parsed := resource.MustParse(value)
	return &parsed
}
