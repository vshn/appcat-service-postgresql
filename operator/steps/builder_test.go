package steps

import (
	"github.com/stretchr/testify/require"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// AssertResourceNotExists checks if the given resource is not existing or is existing with a deletion timestamp.
// Test fails if the resource exists or there's another error.
func AssertResourceNotExists(t *testing.T, deletionTime *metav1.Time, err error) {
	if err != nil {
		require.True(t, apierrors.IsNotFound(err))
	} else {
		require.False(t, deletionTime.IsZero(), "deletion timestamp")
	}
}

type PostgresqlStandaloneBuilder struct {
	*v1alpha1.PostgresqlStandalone
}

func NewInstanceBuilder(name, namespace string) *PostgresqlStandaloneBuilder {
	return &PostgresqlStandaloneBuilder{newInstance(name, namespace)}
}

func (b *PostgresqlStandaloneBuilder) getInstance() *v1alpha1.PostgresqlStandalone {
	return b.PostgresqlStandalone
}

func (b *PostgresqlStandaloneBuilder) setDeploymentNamespace(namespace string) *PostgresqlStandaloneBuilder {
	b.Status.PostgresqlStandaloneObservation = v1alpha1.PostgresqlStandaloneObservation{
		HelmChart: &v1alpha1.ChartMetaStatus{
			DeploymentNamespace: namespace,
		},
	}
	return b
}

func (b *PostgresqlStandaloneBuilder) setConditions(conditions ...metav1.Condition) *PostgresqlStandaloneBuilder {
	b.Status.Conditions = conditions
	return b
}

func (b *PostgresqlStandaloneBuilder) setBackupEnabled(enabled bool) *PostgresqlStandaloneBuilder {
	b.Spec.Backup.Enabled = enabled
	return b
}
