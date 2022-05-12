package standalone

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
)

func TestPostgresqlStandaloneValidator_ValidateUpdate(t *testing.T) {
	tests := map[string]struct {
		givenOldSpec  *v1alpha1.PostgresqlStandalone
		givenNewSpec  *v1alpha1.PostgresqlStandalone
		expectedError string
	}{
		"GivenSameMajorVersion_ThenExpectNoError": {
			givenOldSpec: &v1alpha1.PostgresqlStandalone{
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					Parameters: v1alpha1.PostgresqlStandaloneParameters{MajorVersion: v1alpha1.PostgresqlVersion14},
				},
			},
			givenNewSpec: &v1alpha1.PostgresqlStandalone{
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					Parameters: v1alpha1.PostgresqlStandaloneParameters{MajorVersion: v1alpha1.PostgresqlVersion14},
				},
			},
		},
		"GivenDifferentMajorVersion_ThenExceptError": {
			givenOldSpec: &v1alpha1.PostgresqlStandalone{
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					Parameters: v1alpha1.PostgresqlStandaloneParameters{MajorVersion: v1alpha1.PostgresqlVersion14},
				},
			},
			givenNewSpec: &v1alpha1.PostgresqlStandalone{
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					Parameters: v1alpha1.PostgresqlStandaloneParameters{MajorVersion: "v15"},
				},
			},
			expectedError: "major version cannot be changed once specified at creation time",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			v := PostgresqlStandaloneValidator{}
			err := v.ValidateUpdate(nil, tc.givenOldSpec, tc.givenNewSpec)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError, "validation error")
				return
			}
			require.NoError(t, err, "validation error")
		})
	}
}
