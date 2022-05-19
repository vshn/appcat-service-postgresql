//go:build generate
// +build generate

// Clean samples dir
//go:generate rm -rf package/samples/*

// Generate sample files
//go:generate go run gen_sample.go package/samples

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/v1beta1"
	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/vshn/appcat-service-postgresql/apis"
	"github.com/vshn/appcat-service-postgresql/apis/conditions"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	serializerjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/json"
)

var scheme = runtime.NewScheme()

func main() {
	failIfError(apis.AddToScheme(scheme))
	generatePostgresStandaloneConfigSample()
	generatePostgresStandaloneSample()

	generatePostgresqlStandaloneAdmissionRequest()

	generateProviderHelmConfigSample()
}

func generatePostgresStandaloneConfigSample() {
	spec := &v1alpha1.PostgresqlStandaloneOperatorConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.PostgresqlStandaloneOperatorConfigGroupVersionKind.GroupVersion().String(),
			Kind:       v1alpha1.PostgresqlStandaloneOperatorConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "platform-config-v14",
			Namespace: "postgresql-system",
			Labels: map[string]string{
				v1alpha1.PostgresqlMajorVersionLabelKey: v1alpha1.PostgresqlVersion14.String(),
			}},
		Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
			DeploymentStrategy: v1alpha1.StrategyHelmChart,
			ResourceMinima: v1alpha1.Resources{
				ComputeResources: v1alpha1.ComputeResources{MemoryLimit: parseResource("512Mi")},
				StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("5Gi")},
			},
			ResourceMaxima: v1alpha1.Resources{
				ComputeResources: v1alpha1.ComputeResources{MemoryLimit: parseResource("6Gi")},
				StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("500Gi")},
			},
			HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
				Chart: v1alpha1.ChartMeta{
					Repository: "https://charts.bitnami.com/bitnami",
					Version:    "11.1.23",
					Name:       "postgresql",
				},
				Values: runtime.RawExtension{Raw: toRawJSON(map[string]interface{}{
					"key": "value",
				})},
			},
			HelmReleases: []v1alpha1.HelmReleaseConfig{
				{
					Chart: v1alpha1.ChartMeta{Version: "11.1.23"},
					Values: runtime.RawExtension{Raw: toRawJSON(map[string]interface{}{
						"key":    "overridden",
						"newKey": "newValue",
					})},
					MergeValuesFromTemplate: true,
				},
			},
			HelmProviderConfigReference: "provider-helm",
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
				ModifiedTime: modified,
			},
		},
	}
	serialize(spec, true)
}

func newPostgresqlStandaloneSample() *v1alpha1.PostgresqlStandalone {
	return &v1alpha1.PostgresqlStandalone{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.PostgresqlStandaloneGroupVersionKind.GroupVersion().String(),
			Kind:       v1alpha1.PostgresqlStandaloneKind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "my-instance", Namespace: "default", Generation: 1},
		Spec: v1alpha1.PostgresqlStandaloneSpec{
			Parameters: v1alpha1.PostgresqlStandaloneParameters{
				Resources: v1alpha1.Resources{
					ComputeResources: v1alpha1.ComputeResources{MemoryLimit: parseResource("256Mi")},
					StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("1Gi")},
				},
				MajorVersion:    v1alpha1.PostgresqlVersion14,
				EnableSuperUser: true,
			},
		},
		Status: v1alpha1.PostgresqlStandaloneStatus{},
	}
}

func generatePostgresqlStandaloneAdmissionRequest() {
	spec := newPostgresqlStandaloneSample()
	gvk := metav1.GroupVersionKind{Group: v1alpha1.Group, Version: v1alpha1.Version, Kind: v1alpha1.PostgresqlStandaloneKind}
	gvr := metav1.GroupVersionResource{Group: v1alpha1.Group, Version: v1alpha1.Version, Resource: v1alpha1.PostgresqlStandaloneKind}
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

func generateProviderHelmConfigSample() {
	spec := &helmv1beta1.ProviderConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: helmv1beta1.ProviderConfigGroupVersionKind.GroupVersion().String(), Kind: helmv1beta1.ProviderConfigKind},
		ObjectMeta: metav1.ObjectMeta{
			Name: "provider-helm",
		},
		Spec: helmv1beta1.ProviderConfigSpec{
			Credentials: helmv1beta1.ProviderCredentials{Source: crossplanev1.CredentialsSourceInjectedIdentity},
		},
	}
	serialize(spec, true)
}

func failIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func toRawJSON(vals map[string]interface{}) []byte {
	b, err := json.Marshal(vals)
	failIfError(err)
	return b
}

func serialize(object runtime.Object, useYaml bool) {
	gvk := object.GetObjectKind().GroupVersionKind()
	fileExt := "json"
	if useYaml {
		fileExt = "yaml"
	}
	fileName := fmt.Sprintf("%s_%s.%s", strings.ToLower(gvk.Group), strings.ToLower(gvk.Kind), fileExt)
	f := prepareFile(fileName)
	err := serializerjson.NewSerializerWithOptions(serializerjson.DefaultMetaFactory, scheme, scheme, serializerjson.SerializerOptions{Yaml: useYaml, Pretty: true}).Encode(object, f)
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
