//go:build generate
// +build generate

// Clean samples dir
//go:generate rm -r package/samples

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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/vshn/appcat-service-postgresql/apis"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	providerv1alpha1 "github.com/vshn/appcat-service-postgresql/apis/provider/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var scheme = runtime.NewScheme()

func main() {
	failIfError(apis.AddToScheme(scheme))
	generateProviderConfigSample()
	generatePostgresStandaloneSample()
}

func generateProviderConfigSample() {
	spec := &providerv1alpha1.ProviderConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: providerv1alpha1.ProviderConfigUsageGroupVersionKind.GroupVersion().String(),
			Kind:       providerv1alpha1.ProviderConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "provider-config"},
		Spec: providerv1alpha1.ProviderConfigSpec{
			Credentials: providerv1alpha1.ProviderCredentials{
				Source: xpv1.CredentialsSourceInjectedIdentity,
			},
		},
	}
	serialize(spec)
}

func generatePostgresStandaloneSample() {
	spec := &v1alpha1.PostgresqlStandalone{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.PostgresStandaloneGroupVersionKind.GroupVersion().String(),
			Kind:       v1alpha1.PostgresStandaloneKind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "standalone", Generation: 1},
		Spec: v1alpha1.PostgresqlStandaloneSpec{
			ForProvider: v1alpha1.PostgresqlStandaloneParameters{
				ConfigurableField: "sample",
			},
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{
					Name: "provider-config",
				},
			},
		},
		Status: v1alpha1.PostgresqlStandaloneStatus{
			ObservedGeneration: 1,
			AtProvider: v1alpha1.PostgresqlStandaloneObservation{
				ObservableField: "sample",
			},
		},
	}
	serialize(spec)
}

func failIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func serialize(object runtime.Object) {
	gvk := object.GetObjectKind().GroupVersionKind()
	fileName := fmt.Sprintf("%s_%s.yaml", strings.ToLower(gvk.Group), strings.ToLower(gvk.Kind))
	f := prepareFile(fileName)
	err := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: true}).Encode(object, f)
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
