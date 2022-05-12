package standalone

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCreateStandalonePipeline_UseTemplateValues(t *testing.T) {
	p := &CreateStandalonePipeline{
		config: &v1alpha1.PostgresqlStandaloneOperatorConfig{Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
			HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
				Values: runtime.RawExtension{Raw: []byte(`{"key":"value"}`)},
				Chart:  v1alpha1.ChartMeta{Repository: "https://host/path", Name: "postgresql", Version: "1.0"},
			},
		}},
	}
	err := p.useTemplateValues(nil)
	assert.NoError(t, err)
	expectedValues := HelmValues{
		"key": "value",
	}
	expectedChart := &v1alpha1.ChartMeta{
		Repository: "https://host/path",
		Name:       "postgresql",
		Version:    "1.0",
	}
	assert.Equal(t, expectedValues, p.helmValues)
	assert.Equal(t, expectedChart, p.helmChart)
}

func TestCreateStandalonePipeline_OverrideTemplateValues(t *testing.T) {
	tests := map[string]struct {
		givenSpec      v1alpha1.PostgresqlStandaloneOperatorConfigSpec
		expectedValues HelmValues
		expectedChart  v1alpha1.ChartMeta
		expectedError  string
	}{
		"GivenSpecificReleaseExists_WhenMergeEnabled_ThenMergeWithExistingValues": {
			givenSpec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
				HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
					Values: HelmValues{"key": "value", "existing": "untouched"}.MustMarshal(),
					Chart:  v1alpha1.ChartMeta{Version: "version"},
				},
				HelmReleases: []v1alpha1.HelmReleaseConfig{
					{
						MergeValuesFromTemplate: true,
						Chart:                   v1alpha1.ChartMeta{Version: "version"},
						Values:                  HelmValues{"key": map[string]interface{}{"nested": "value"}, "merged": "newValue"}.MustMarshal(),
					},
				},
			},
			expectedValues: HelmValues{
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
					Values: HelmValues{"key": "value", "existing": "untouched"}.MustMarshal(),
					Chart:  v1alpha1.ChartMeta{Version: "version"},
				},
				HelmReleases: []v1alpha1.HelmReleaseConfig{
					{
						Chart:  v1alpha1.ChartMeta{Version: "version", Name: "alternative", Repository: "fork"},
						Values: HelmValues{"key": map[string]interface{}{"nested": "value"}, "merged": "newValue"}.MustMarshal(),
					},
				},
			},
			expectedValues: HelmValues{
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
			vals := HelmValues{}
			vals.MustUnmarshal(tc.givenSpec.HelmReleaseTemplate.Values)
			p := &CreateStandalonePipeline{
				config:     &v1alpha1.PostgresqlStandaloneOperatorConfig{Spec: tc.givenSpec},
				helmValues: vals,
				helmChart:  &v1alpha1.ChartMeta{Repository: "url", Name: "postgres", Version: "version"},
			}
			err := p.overrideTemplateValues(nil)
			if tc.expectedError != "" {
				require.EqualError(t, err, tc.expectedError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedValues, p.helmValues)
			assert.Equal(t, &tc.expectedChart, p.helmChart)
		})
	}
}

func TestCreateStandalonePipeline_ApplyValuesFromInstance(t *testing.T) {
	p := CreateStandalonePipeline{
		config:   newPostgresqlStandaloneOperatorConfig("cfg", "postgresql-system"),
		instance: newInstance("instance"),
	}
	p.instance.UID = "1aa230ee-63f7-4e7f-9ade-46818595e337"
	err := p.applyValuesFromInstance(nil)
	require.NoError(t, err)
	assert.Equal(t, HelmValues{
		"auth": HelmValues{
			"existingSecret":     "instance-credentials",
			"database":           "instance",
			"enablePostgresUser": true,
		},
		"primary": HelmValues{
			"persistence": HelmValues{
				"size": "1Gi",
			},
			"resources": HelmValues{
				"limits": HelmValues{
					"memory": "2Gi",
				},
			},
		},
	}, p.helmValues)
}

func newPostgresqlStandaloneOperatorConfig(name string, namespace string) *v1alpha1.PostgresqlStandaloneOperatorConfig {
	return &v1alpha1.PostgresqlStandaloneOperatorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.PostgresqlMajorVersionLabelKey: v1alpha1.PostgresqlVersion14.String(),
			},
		},
		Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
			HelmProviderConfigReference: "helm-provider",
		},
	}
}
func newInstance(name string) *v1alpha1.PostgresqlStandalone {
	return &v1alpha1.PostgresqlStandalone{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "my-app"},
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
	}
}
func parseResource(value string) *resource.Quantity {
	parsed := resource.MustParse(value)
	return &parsed
}
