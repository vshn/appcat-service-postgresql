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

	"github.com/vshn/appcat-service-postgresql/apis"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	authv1 "k8s.io/api/authentication/v1"
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
		ObjectMeta: metav1.ObjectMeta{Name: "provider-config"},
		Spec:       v1alpha1.PostgresqlStandaloneConfigSpec{},
	}
	serialize(spec, true)
}

func generatePostgresStandaloneSample() {
	spec := newPostgresqlStandaloneSample()
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
			Resources: v1alpha1.Resources{
				ComputeResources: v1alpha1.ComputeResources{
					MemoryLimit: "2Gi",
				},
				StorageResources: v1alpha1.StorageResources{
					Size: "8Gi",
				},
			},
			MonitoringEnabledInstance: v1alpha1.MonitoringEnabledInstance{
				Monitoring: v1alpha1.Monitoring{
					SLA: v1alpha1.SlaBestEffort,
				},
			},
			DeferrableMaintenance: v1alpha1.DeferrableMaintenance{
				UpdatePolicy: v1alpha1.UpdatePolicy{
					MaintenanceWindow: v1alpha1.MaintenanceWindow{
						Start: v1alpha1.MaintenanceWindowSelector{
							//Weekday: "Monday",
						},
					},
				},
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
