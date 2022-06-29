package standalone

import (
	pipeline "github.com/ccremer/go-command-pipeline"
	"golang.org/x/net/context"
	"testing"
	"time"

	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/stretchr/testify/assert"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateStandalonePipeline_IsHelmReleaseReady(t *testing.T) {
	ctx := pipeline.MutableContext(context.Background())
	instance := newInstance("release-ready", "my-app")
	setInstanceInContext(ctx, instance)
	instance.Status.HelmChart = &v1alpha1.ChartMetaStatus{}

	modifiedDate := metav1.Date(2022, 05, 17, 17, 52, 35, 0, time.Local)
	helmRelease := &helmv1beta1.Release{
		ObjectMeta: metav1.ObjectMeta{Name: generateClusterScopedNameForInstance()},
	}
	setHelmReleaseInContext(ctx, helmRelease)

	// Arrange
	p := CreateStandalonePipeline{}
	t.Run("check non-ready release", func(t *testing.T) {
		// Act
		result := p.isHelmReleaseReady(ctx)

		// Assert
		assert.False(t, result)
		assert.True(t, instance.Status.HelmChart.ModifiedTime.IsZero())
	})

	t.Run("check ready release", func(t *testing.T) {
		helmRelease.Status = helmv1beta1.ReleaseStatus{
			ResourceStatus: crossplanev1.ResourceStatus{
				ConditionedStatus: crossplanev1.ConditionedStatus{Conditions: []crossplanev1.Condition{
					{
						Type:               crossplanev1.TypeReady,
						Status:             corev1.ConditionTrue,
						LastTransitionTime: modifiedDate,
					},
				}},
			},
			Synced: true,
		}

		// Act
		result := p.isHelmReleaseReady(ctx)

		// Assert
		assert.Equal(t, modifiedDate, instance.Status.HelmChart.ModifiedTime)
		assert.True(t, result)
	})
}
