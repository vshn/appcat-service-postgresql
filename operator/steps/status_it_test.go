//go:build integration

package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/conditions"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

type StatusSuite struct {
	operatortest.Suite
}

func TestStatusSuite(t *testing.T) {
	suite.Run(t, new(StatusSuite))
}

func (ts *StatusSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	SetClientInContext(ts.Context, ts.Client)
}

func (ts *StatusSuite) Test_MarkInstanceAsReady() {
	tests := map[string]struct {
		prepare          func(*v1alpha1.PostgresqlStandalone)
		givenInstance    *v1alpha1.PostgresqlStandalone
		expectedInstance *v1alpha1.PostgresqlStandalone
	}{
		"GivenInstanceWithProgressingCondition_WhenMarkAsReady_ThenExpectReadyCondition": {
			prepare: func(instance *v1alpha1.PostgresqlStandalone) {
				ts.EnsureNS("instance-second-namespace")
				ts.EnsureResources(instance)
			},
			givenInstance: NewInstanceBuilder("my-second-instance", "instance-second-namespace").
				setConditions(
					metav1.Condition{
						Type:               conditions.TypeProgressing,
						Status:             metav1.ConditionTrue,
						Reason:             "ProgressingResource",
						ObservedGeneration: 1,
					},
				).
				getInstance(),
			expectedInstance: NewInstanceBuilder("my-second-instance", "instance-second-namespace").
				setConditions(
					metav1.Condition{
						Type:               conditions.TypeReady,
						Status:             metav1.ConditionTrue,
						Reason:             "Available",
						ObservedGeneration: 1,
					},
				).getInstance(),
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			SetInstanceInContext(ts.Context, tc.givenInstance)
			tc.prepare(tc.givenInstance)

			// Act
			err := MarkInstanceAsReadyFn()(ts.Context)
			ts.Require().NoError(err)

			// Assert
			actualInstance := &v1alpha1.PostgresqlStandalone{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-second-instance",
					Namespace: "instance-second-namespace",
				},
			}
			ts.FetchResource(client.ObjectKeyFromObject(actualInstance), actualInstance)
			expectedStatus := tc.expectedInstance.Status
			ts.Assert().Len(actualInstance.Status.Conditions, 1, "amount of conditions")
			ts.Assert().Equal(expectedStatus.Conditions[0].Status, actualInstance.Status.Conditions[0].Status)
			ts.Assert().Equal(expectedStatus.Conditions[0].Type, actualInstance.Status.Conditions[0].Type)
			ts.Assert().Equal(expectedStatus.Conditions[0].Reason, actualInstance.Status.Conditions[0].Reason)
			ts.Assert().Equal(expectedStatus.Conditions[0].ObservedGeneration, actualInstance.Status.Conditions[0].ObservedGeneration)
		})
	}
}

func (ts *StatusSuite) Test_MarkInstanceAsProgressing() {
	tests := map[string]struct {
		prepare          func(*v1alpha1.PostgresqlStandalone)
		givenInstance    *v1alpha1.PostgresqlStandalone
		expectedInstance *v1alpha1.PostgresqlStandalone
	}{
		"GivenInstanceIsUpdated_WhenInitialUpdateReconcileFinished_ThenExpectStatusProgressingTrue": {
			prepare: func(instance *v1alpha1.PostgresqlStandalone) {
				ts.EnsureNS("instance-first-namespace")
				ts.EnsureResources(instance)
			},
			givenInstance: newInstance("my-first-instance", "instance-first-namespace"),
			expectedInstance: NewInstanceBuilder("my-first-instance", "instance-first-namespace").
				setConditions(
					metav1.Condition{
						Type:               conditions.TypeProgressing,
						Status:             metav1.ConditionTrue,
						Reason:             "ProgressingResource",
						ObservedGeneration: 1,
					},
					metav1.Condition{
						Type:               conditions.TypeReady,
						Status:             metav1.ConditionFalse,
						Reason:             "NotAvailable",
						ObservedGeneration: 1,
					},
				).getInstance(),
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			SetClientInContext(ts.Context, ts.Client)
			SetInstanceInContext(ts.Context, tc.givenInstance)
			tc.prepare(tc.givenInstance)

			// Act
			err := MarkInstanceAsProgressingFn()(ts.Context)
			ts.Require().NoError(err)

			// Assert
			actualInstance := &v1alpha1.PostgresqlStandalone{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-first-instance",
					Namespace: "instance-first-namespace",
				},
			}
			ts.FetchResource(client.ObjectKeyFromObject(actualInstance), actualInstance)
			expectedStatus := tc.expectedInstance.Status
			ts.Assert().Len(actualInstance.Status.Conditions, 2, "amount of conditions")
			ts.Assert().Equal(expectedStatus.Conditions[0].Status, actualInstance.Status.Conditions[0].Status)
			ts.Assert().Equal(expectedStatus.Conditions[0].Type, actualInstance.Status.Conditions[0].Type)
			ts.Assert().Equal(expectedStatus.Conditions[0].Reason, actualInstance.Status.Conditions[0].Reason)
			ts.Assert().Equal(expectedStatus.Conditions[0].ObservedGeneration, actualInstance.Status.Conditions[0].ObservedGeneration)
			ts.Assert().Equal(expectedStatus.Conditions[1].Status, actualInstance.Status.Conditions[1].Status)
			ts.Assert().Equal(expectedStatus.Conditions[1].Type, actualInstance.Status.Conditions[1].Type)
			ts.Assert().Equal(expectedStatus.Conditions[1].Reason, actualInstance.Status.Conditions[1].Reason)
			ts.Assert().Equal(expectedStatus.Conditions[1].ObservedGeneration, actualInstance.Status.Conditions[1].ObservedGeneration)
		})
	}
}
