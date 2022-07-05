package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

type NamespaceSuite struct {
	operatortest.Suite
}

func TestCreateStandalonePipeline(t *testing.T) {
	suite.Run(t, new(NamespaceSuite))
}

func (ts *NamespaceSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	SetClientInContext(ts.Context, ts.Client)
}

func (ts *NamespaceSuite) Test_EnsureDeploymentNamespace() {
	// Arrange
	instance := newInstance("test-ensure-namespace", "my-app")
	SetInstanceInContext(ts.Context, instance)

	// Act
	err := EnsureNamespace("sv-postgresql-s-merry-vigilante-7b16", labels.Set{
		"app.kubernetes.io/instance":           instance.Name,
		"app.kubernetes.io/instance-namespace": instance.Namespace,
	})(ts.Context)
	ts.Require().NoError(err, "create namespace func")

	// Assert
	ns := &corev1.Namespace{}
	ts.FetchResource(types.NamespacedName{Name: "sv-postgresql-s-merry-vigilante-7b16"}, ns)
	ts.Assert().Equal(ns.Labels["app.kubernetes.io/instance"], instance.Name)
	ts.Assert().Equal(ns.Labels["app.kubernetes.io/instance-namespace"], instance.Namespace)
}

func (ts *NamespaceSuite) Test_DeleteNamespace() {
	tests := map[string]struct {
		prepare        func(namespace string)
		givenNamespace string
	}{
		"GivenNonExistingNamespace_WhenDeleting_ThenExpectNoFurtherAction": {
			prepare:        func(namespace string) {},
			givenNamespace: "non-existing-namespace",
		},
		"GivenExistingNamespace_WhenDeleting_ThenExpectNoFurtherAction": {
			prepare:        func(namespace string) { ts.EnsureNS(namespace) },
			givenNamespace: "existing-namespace",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			instance := NewInstanceBuilder("instance", "test-delete").setDeploymentNamespace(tc.givenNamespace).getInstance()
			SetInstanceInContext(ts.Context, instance)
			tc.prepare(tc.givenNamespace)

			// Act
			err := DeleteNamespaceFn()(ts.Context)
			ts.Require().NoError(err)

			// Assert
			resultNs := &corev1.Namespace{}
			err = ts.Client.Get(
				ts.Context,
				types.NamespacedName{Name: tc.givenNamespace},
				resultNs,
			)
			AssertResourceNotExists(ts.T(), resultNs.GetDeletionTimestamp(), err)
		})
	}
}
