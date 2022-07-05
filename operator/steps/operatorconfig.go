package steps

import (
	"context"
	"fmt"
	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchOperatorConfigFn fetches a matching v1alpha1.PostgresqlStandaloneOperatorConfig from the OperatorNamespace.
// The Major version specified in v1alpha1.PostgresqlStandalone is used to filter the correct config by the v1alpha1.PostgresqlMajorVersionLabelKey label.
// If there is none or multiple found, it returns an error.
func FetchOperatorConfigFn(operatorNamespace string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)

		list := &v1alpha1.PostgresqlStandaloneOperatorConfigList{}
		labelMatch := client.MatchingLabels{
			v1alpha1.PostgresqlMajorVersionLabelKey: instance.Spec.Parameters.MajorVersion.String(),
		}
		ns := client.InNamespace(operatorNamespace)
		err := kube.List(ctx, list, labelMatch, ns)
		if err != nil {
			return err
		}
		if len(list.Items) == 0 {
			return fmt.Errorf("no %s found with label '%s' in namespace '%s'", v1alpha1.PostgresqlStandaloneOperatorConfigKind, labelMatch, ns)
		}
		if len(list.Items) > 1 {
			return fmt.Errorf("multiple versions of %s found with label '%s' in namespace '%s'", v1alpha1.PostgresqlStandaloneOperatorConfigKind, labelMatch, ns)
		}

		pipeline.StoreInContext(ctx, ConfigKey{}, &list.Items[0])
		return nil
	}
}
