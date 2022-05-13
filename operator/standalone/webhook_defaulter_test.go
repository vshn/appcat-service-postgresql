package standalone

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPostgresqlStandaloneDefaulter_Default(t *testing.T) {
	tests := map[string]struct {
		givenInstance    *v1alpha1.PostgresqlStandalone
		expectedInstance *v1alpha1.PostgresqlStandalone
		expectedError    string
	}{
		"GivenEmptyWriteConnectionSecretToRef_ThenExpectInstanceName": {
			givenInstance: &v1alpha1.PostgresqlStandalone{
				ObjectMeta: metav1.ObjectMeta{Name: "my-instance"},
			},
			expectedInstance: &v1alpha1.PostgresqlStandalone{
				ObjectMeta: metav1.ObjectMeta{Name: "my-instance"},
				Spec: v1alpha1.PostgresqlStandaloneSpec{
					ConnectableInstance: v1alpha1.ConnectableInstance{
						WriteConnectionSecretToRef: v1alpha1.ConnectionSecretRef{Name: "my-instance"},
					},
				},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d := &PostgresqlStandaloneDefaulter{}
			err := d.Default(nil, tc.givenInstance)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError, "defaulter error")
				return
			}
			require.NoError(t, err, "defaulter error")
			assert.Equal(t, tc.expectedInstance, tc.givenInstance)
		})
	}
}
