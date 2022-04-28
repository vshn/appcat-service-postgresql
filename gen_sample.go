//go:build generate
// +build generate

// Clean samples dir
//go:generate rm -rf chart/samples/*

// Generate sample files
//go:generate go run gen_sample.go chart/samples

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vshn/appcat-service-postgresql/apis"
	"github.com/vshn/appcat-service-postgresql/apis/conditions"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var scheme = runtime.NewScheme()

func main() {
	failIfError(apis.AddToScheme(scheme))
	generatePostgresStandaloneConfigSample()
	generatePostgresStandaloneSample()
	generatePostgresqlStandaloneAdmissionRequest()
}

func generatePostgresqlStandaloneAdmissionRequest() {
	spec := newPostgresqlStandaloneSample()
	gvk := metav1.GroupVersionKind{Group: v1alpha1.Group, Version: v1alpha1.Version, Kind: v1alpha1.PostgresStandaloneKind}
	gvr := metav1.GroupVersionResource{Group: v1alpha1.Group, Version: v1alpha1.Version, Resource: v1alpha1.PostgresStandaloneKind}
	admission := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
		Request: &admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{
				Object: spec,
			},
			Kind:            gvk,
			Resource:        gvr,
			RequestKind:     &gvk,
			RequestResource: &gvr,
			Name:            spec.Name,
			Operation:       admissionv1.Create,
			UserInfo: authv1.UserInfo{
				Username: "admin",
				Groups:   []string{"system:authenticated"},
			},
		},
	}
	serialize(admission, false)
}

func generatePostgresStandaloneConfigSample() {
	spec := &v1alpha1.PostgresqlStandaloneConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.PostgresqlStandaloneConfigGroupVersionKind.GroupVersion().String(),
			Kind:       v1alpha1.PostgresqlStandaloneConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "platform-config"},
		Spec: v1alpha1.PostgresqlStandaloneConfigSpec{
			ResourceMinima: v1alpha1.Resources{
				ComputeResources: v1alpha1.ComputeResources{MemoryLimit: parseResource("512Mi")},
				StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("5Gi")},
			},
			ResourceMaxima: v1alpha1.Resources{
				ComputeResources: v1alpha1.ComputeResources{MemoryLimit: parseResource("6Gi")},
				StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("500Gi")},
			},
		},
	}
	serialize(spec, true)
}

func generatePostgresStandaloneSample() {
	spec := newPostgresqlStandaloneSample()
	modified := metav1.Date(2022, time.April, 27, 15, 20, 13, 0, time.UTC)
	cond := conditions.Ready()
	cond.LastTransitionTime = modified
	spec.Status = v1alpha1.PostgresqlStandaloneStatus{
		Conditions: []metav1.Condition{cond},
		PostgresqlStandaloneObservation: v1alpha1.PostgresqlStandaloneObservation{
			DeploymentStrategy: v1alpha1.StrategyHelmChart,
			HelmChart: &v1alpha1.ChartMetaStatus{
				ChartMeta: v1alpha1.ChartMeta{
					Repository: "https://charts.bitnami.com/bitnami",
					Version:    "11.1.23",
					Name:       "postgresql",
				},
				ModifiedAt: &modified,
			},
		},
	}
	serialize(spec, true)
}

func newPostgresqlStandaloneSample() *v1alpha1.PostgresqlStandalone {
	return &v1alpha1.PostgresqlStandalone{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.PostgresStandaloneGroupVersionKind.GroupVersion().String(),
			Kind:       v1alpha1.PostgresStandaloneKind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "standalone", Generation: 1},
		Spec: v1alpha1.PostgresqlStandaloneSpec{
			Parameters: v1alpha1.PostgresqlStandaloneParameters{
				Resources: v1alpha1.Resources{
					ComputeResources: v1alpha1.ComputeResources{},
					StorageResources: v1alpha1.StorageResources{},
				},
				MajorVersion:    v1alpha1.PostgresqlVersion14,
				EnableSuperUser: true,
			},
		},
		Status: v1alpha1.PostgresqlStandaloneStatus{},
	}
}

func failIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func serialize(object runtime.Object, useYaml bool) {
	gvk := object.GetObjectKind().GroupVersionKind()
	fileExt := "json"
	if useYaml {
		fileExt = "yaml"
	}
	fileName := fmt.Sprintf("%s_%s.%s", strings.ToLower(gvk.Group), strings.ToLower(gvk.Kind), fileExt)
	f := prepareFile(fileName)
	err := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: useYaml, Pretty: true}).Encode(object, f)
	failIfError(err)
}

func prepareFile(file string) io.Writer {
	dir := os.Args[1]
	err := os.MkdirAll(os.Args[1], 0775)
	failIfError(err)
	f, err := os.Create(filepath.Join(dir, file))
	failIfError(err)
	return f
}

func parseResource(value string) *resource.Quantity {
	parsed := resource.MustParse(value)
	return &parsed
}
