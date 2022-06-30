package standalone

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/stretchr/testify/assert"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func Test_IsHelmReleaseReady(t *testing.T) {
	initialTime := metav1.Time{Time: time.Unix(10000, 0)}
	timeOne := metav1.Time{Time: time.Unix(20000, 0)}
	timeTwo := metav1.Time{Time: time.Unix(30000, 0)}
	tests := map[string]struct {
		prepare              func(*v1alpha1.PostgresqlStandalone)
		givenInstance        *v1alpha1.PostgresqlStandalone
		givenHelmRelease     *helmv1beta1.Release
		expectedInstance     *v1alpha1.PostgresqlStandalone
		expectedModifiedTime metav1.Time
		expectedReleaseReady bool
	}{
		"GivenAnUpdatedInstance_WhenHelmReleaseStatusNotSynced_ThenExpectHelmReleaseNotReady": {
			givenInstance: newInstanceBuilder("release-one-instance", "release-one-namespace").
				setHelmChartModifiedTime(initialTime).
				getInstance(),
			givenHelmRelease: newHelmReleaseBuilder("release-one").
				setSynced(false).
				getRelease(),
			expectedModifiedTime: initialTime,
			expectedReleaseReady: false,
		},
		"GivenAnUpdatedInstance_WhenHelmReleaseStatusSyncedAndWithoutConditionReady_ThenExpectHelmReleaseNotReady": {
			givenInstance: newInstanceBuilder("release-two-instance", "release-two-namespace").
				setHelmChartModifiedTime(initialTime).
				getInstance(),
			givenHelmRelease: newHelmReleaseBuilder("release-two").
				setSynced(true).
				setConditions(
					crossplanev1.Condition{
						Type:               crossplanev1.TypeSynced,
						Status:             v1.ConditionTrue,
						Reason:             crossplanev1.ReasonUnavailable,
						LastTransitionTime: timeTwo,
					},
				).
				getRelease(),
			expectedModifiedTime: initialTime,
			expectedReleaseReady: false,
		},
		"GivenAnUpdatedInstance_WhenHelmReleaseStatusSyncedAndWithConditionReadyFalse_ThenExpectHelmReleaseNotReady": {
			givenInstance: newInstanceBuilder("release-two-instance", "release-two-namespace").
				setHelmChartModifiedTime(initialTime).
				getInstance(),
			givenHelmRelease: newHelmReleaseBuilder("release-two").
				setSynced(true).
				setConditions(
					crossplanev1.Condition{
						Type:               crossplanev1.TypeSynced,
						Status:             v1.ConditionTrue,
						Reason:             crossplanev1.ReasonUnavailable,
						LastTransitionTime: timeOne,
					},
					crossplanev1.Condition{
						Type:               crossplanev1.TypeReady,
						Status:             v1.ConditionFalse,
						Reason:             crossplanev1.ReasonUnavailable,
						LastTransitionTime: timeTwo,
					},
				).
				getRelease(),
			expectedModifiedTime: initialTime,
			expectedReleaseReady: false,
		},
		"GivenAnUpdatedInstance_WhenHelmReleaseStatusSyncedAndWithConditionReadyTrue_ThenExpectHelmReleaseReady": {
			givenInstance: newInstanceBuilder("release-two-instance", "release-two-namespace").
				setHelmChartModifiedTime(initialTime).
				getInstance(),
			givenHelmRelease: newHelmReleaseBuilder("release-two").
				setSynced(true).
				setConditions(
					crossplanev1.Condition{
						Type:               crossplanev1.TypeSynced,
						Status:             v1.ConditionFalse,
						Reason:             crossplanev1.ReasonUnavailable,
						LastTransitionTime: timeOne,
					},
					crossplanev1.Condition{
						Type:               crossplanev1.TypeReady,
						Status:             v1.ConditionTrue,
						Reason:             crossplanev1.ReasonUnavailable,
						LastTransitionTime: timeTwo,
					},
				).
				getRelease(),
			expectedModifiedTime: timeTwo,
			expectedReleaseReady: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Given
			p := UpdateStandalonePipeline{}
			ctx := pipeline.MutableContext(context.Background())
			setInstanceInContext(ctx, tc.givenInstance)
			setHelmReleaseInContext(ctx, tc.givenHelmRelease)

			//When
			actualReleaseReady := p.isHelmReleaseReady(ctx)

			//Then
			assert.Equal(t, tc.expectedReleaseReady, actualReleaseReady)
			assert.Equal(t, tc.expectedModifiedTime, getInstanceFromContext(ctx).Status.HelmChart.ModifiedTime)
		})
	}
}
