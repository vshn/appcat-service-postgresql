//go:build generate
// +build generate

// Clean samples dir
//go:generate rm -rf chart/samples/*.yaml

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
			APIVersion: providerv1alpha1.ProviderConfigGroupVersionKind.GroupVersion().String(),
			Kind:       providerv1alpha1.ProviderConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "provider-config"},
		Spec:       providerv1alpha1.ProviderConfigSpec{},
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
				DeploymentStrategy: "",
				Chart: &v1alpha1.ChartMeta{
					Repository: "https://charts.bitnami.com/bitnami",
					Version:    "12.0",
					Name:       "postgres",
				},
				BackupEnabledInstance: v1alpha1.BackupEnabledInstance{
					Backup: v1alpha1.Backup{Enabled: true},
				},
				MonitoringEnabledInstance: v1alpha1.MonitoringEnabledInstance{
					Monitoring: v1alpha1.Monitoring{
						SLA: v1alpha1.SlaBestEffort.String(),
					},
				},
				DelayableMaintenance: v1alpha1.DelayableMaintenance{
					UpdatePolicy: v1alpha1.UpdatePolicy{
						Version: v1alpha1.VersionSelector{Major: "14", Minor: "0"},
						MaintenanceWindow: v1alpha1.MaintenanceWindow{
							Start: v1alpha1.MaintenanceWindowSelector{Weekday: "Wednesday", Hour: 15},
						},
					},
				},
			},
		},
		Status: v1alpha1.PostgresqlStandaloneStatus{
			AtProvider: v1alpha1.PostgresqlStandaloneObservation{
				Chart: &v1alpha1.ChartMeta{
					Repository: "https://charts.bitnami.com/bitnami",
					Version:    "12.0",
					Name:       "postgres",
				},
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
