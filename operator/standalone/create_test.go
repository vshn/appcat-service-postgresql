package standalone

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateStandalonePipeline_OverrideTemplateValues(t *testing.T) {
	tests := map[string]struct {
		givenSpec      v1alpha1.PostgresqlStandaloneOperatorConfigSpec
		expectedValues HelmValues
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
		},
		"GivenSpecificReleaseExists_WhenMergeDisabled_ThenOverwriteExistingValues": {
			givenSpec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
				HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
					Values: HelmValues{"key": "value", "existing": "untouched"}.MustMarshal(),
					Chart:  v1alpha1.ChartMeta{Version: "version"},
				},
				HelmReleases: []v1alpha1.HelmReleaseConfig{
					{
						Chart:  v1alpha1.ChartMeta{Version: "version"},
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
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			vals := HelmValues{}
			vals.MustUnmarshal(tc.givenSpec.HelmReleaseTemplate.Values)
			p := &CreateStandalonePipeline{
				config:     &v1alpha1.PostgresqlStandaloneOperatorConfig{Spec: tc.givenSpec},
				helmValues: vals,
			}
			err := p.OverrideTemplateValues(nil)
			if tc.expectedError != "" {
				require.EqualError(t, err, tc.expectedError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedValues, p.helmValues)
		})
	}
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
	}
}
func newInstance() *v1alpha1.PostgresqlStandalone {
	return &v1alpha1.PostgresqlStandalone{
		Spec: v1alpha1.PostgresqlStandaloneSpec{
			Parameters: v1alpha1.PostgresqlStandaloneParameters{
				MajorVersion: v1alpha1.PostgresqlVersion14,
			},
		},
	}
}
