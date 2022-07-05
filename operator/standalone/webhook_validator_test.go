package standalone

import (
	"k8s.io/apimachinery/pkg/api/resource"
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
		"GivenMajorVersion_WhenVersionSame_ThenExpectNil": {
			givenOldSpec: &v1alpha1.PostgresqlStandalone{
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					Parameters: v1alpha1.PostgresqlStandaloneParameters{
						MajorVersion: v1alpha1.PostgresqlVersion14,
						Resources: v1alpha1.Resources{
							StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("1Gi")},
						}},
				},
			},
			givenNewSpec: &v1alpha1.PostgresqlStandalone{
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					Parameters: v1alpha1.PostgresqlStandaloneParameters{
						MajorVersion: v1alpha1.PostgresqlVersion14,
						Resources: v1alpha1.Resources{
							StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("1Gi")},
						}},
				},
			},
		},
		"GivenMajorVersion_WhenVersionDifferent_ThenExpectError": {
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
		"GivenStorageCapacity_WhenIncreased_ThenExpectNil": {
			givenOldSpec: &v1alpha1.PostgresqlStandalone{
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					Parameters: v1alpha1.PostgresqlStandaloneParameters{Resources: v1alpha1.Resources{
						StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("1Gi")},
					}},
				},
			},
			givenNewSpec: &v1alpha1.PostgresqlStandalone{
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					Parameters: v1alpha1.PostgresqlStandaloneParameters{Resources: v1alpha1.Resources{
						StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("1.1Gi")},
					}},
				},
			},
		},
		"GivenStorageCapacity_WhenDecreased_ThenExpectError": {
			givenOldSpec: &v1alpha1.PostgresqlStandalone{
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					Parameters: v1alpha1.PostgresqlStandaloneParameters{Resources: v1alpha1.Resources{
						StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("1Gi")},
					}},
				},
			},
			givenNewSpec: &v1alpha1.PostgresqlStandalone{
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					Parameters: v1alpha1.PostgresqlStandaloneParameters{Resources: v1alpha1.Resources{
						StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("0.9Gi")},
					}},
				},
			},
			expectedError: "storage capacity cannot be decreased",
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

func parseResource(value string) *resource.Quantity {
	parsed := resource.MustParse(value)
	return &parsed
}
