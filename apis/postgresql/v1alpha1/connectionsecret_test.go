package v1alpha1

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestPostgresqlStandalone_GetConnectionSecretName(t *testing.T) {
	instance := PostgresqlStandalone{
		ObjectMeta: metav1.ObjectMeta{Name: "instance"},
	}
	t.Run("GivenEmptySecretName_ThenExpectInstanceName", func(t *testing.T) {
		instance.Spec.WriteConnectionSecretToRef.Name = ""

		result := instance.GetConnectionSecretName()
		assert.Equal(t, "instance", result)
	})
	t.Run("GivenExplicitSecretName_ThenExpectGivenName", func(t *testing.T) {
		instance.Spec.WriteConnectionSecretToRef.Name = "my-secret"

		result := instance.GetConnectionSecretName()
		assert.Equal(t, "my-secret", result)
	})
}
