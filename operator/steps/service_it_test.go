package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type ServiceSuite struct {
	operatortest.Suite
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (ts *ServiceSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	SetClientInContext(ts.Context, ts.Client)
}

func (ts *ServiceSuite) Test_FetchService() {
	// Arrange
	instance := newInstance("fetch-service", "my-app")
	SetInstanceInContext(ts.Context, instance)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: "service-ns",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 5432}},
		},
	}
	ts.EnsureNS("my-app")
	ts.EnsureNS(service.Namespace)
	ts.EnsureResources(service)
	instance.Status.HelmChart = &v1alpha1.ChartMetaStatus{DeploymentNamespace: service.Namespace}

	// Act
	err := FetchServiceFn()(ts.Context)
	ts.Require().NoError(err)

	// Assert
	result := getFromContextOrPanic(ts.Context, ServiceKey{}).(*corev1.Service)
	ts.Assert().Equal(service, result)
}
