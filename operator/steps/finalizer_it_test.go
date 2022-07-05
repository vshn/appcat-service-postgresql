package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

type FinalizerSuite struct {
	operatortest.Suite
}

func TestFinalizerSuite(t *testing.T) {
	suite.Run(t, new(FinalizerSuite))
}

func (ts *FinalizerSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	SetClientInContext(ts.Context, ts.Client)
}

func (ts *FinalizerSuite) Test_RemoveFinalizer() {
	tests := map[string]struct {
		prepare        func(instance *v1alpha1.PostgresqlStandalone)
		givenInstance  string
		givenNamespace string
		assert         func(previousInstance, result *v1alpha1.PostgresqlStandalone)
	}{
		"GivenInstanceWithFinalizer_WhenDeletingFinalizer_ThenExpectInstanceUpdatedWithRemovedFinalizer": {
			prepare: func(instance *v1alpha1.PostgresqlStandalone) {
				instance.Finalizers = []string{"finalizer"}
				ts.EnsureNS("remove-finalizer")
				ts.EnsureResources(instance)
				ts.Assert().NotEmpty(instance.Finalizers)
			},

			givenInstance:  "has-finalizer",
			givenNamespace: "remove-finalizer",
			assert: func(previousInstance, result *v1alpha1.PostgresqlStandalone) {
				ts.Assert().Empty(result.Finalizers)
				ts.Assert().NotEqual(previousInstance.ResourceVersion, result.ResourceVersion, "resource version should change")
			},
		},
		"GivenInstanceWithoutFinalizer_WhenDeletingFinalizer_ThenExpectInstanceUnchanged": {
			prepare: func(instance *v1alpha1.PostgresqlStandalone) {
				ts.EnsureNS("remove-finalizer")
				ts.EnsureResources(instance)
			},

			givenInstance:  "no-finalizer",
			givenNamespace: "remove-finalizer",
			assert: func(previousInstance, result *v1alpha1.PostgresqlStandalone) {
				ts.Assert().Empty(result.Finalizers)
				ts.Assert().Equal(previousInstance.ResourceVersion, result.ResourceVersion, "resource version should be equal")
			},
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			instance := NewInstanceBuilder(tc.givenInstance, tc.givenNamespace).getInstance()
			SetInstanceInContext(ts.Context, instance)
			tc.prepare(instance)
			previousVersion := instance.DeepCopy()

			// Act
			err := RemoveFinalizerFn("finalizer")(ts.Context)
			ts.Require().NoError(err)

			// Assert
			result := &v1alpha1.PostgresqlStandalone{}
			ts.FetchResource(client.ObjectKeyFromObject(instance), result)
			tc.assert(previousVersion, result)
		})
	}
}
