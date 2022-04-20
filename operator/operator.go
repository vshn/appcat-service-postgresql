package operator

import (
	"github.com/vshn/appcat-service-postgresql/operator/config"
	"github.com/vshn/appcat-service-postgresql/operator/standalone"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// SetupControllers creates all Postgresql controllers with the supplied logger and adds them to the supplied manager.
func SetupControllers(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		config.Setup,
		standalone.SetupController,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhooks creates all Postgresql webhooks with the supplied logger and adds them to the supplied manager.
func SetupWebhooks(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		standalone.SetupWebhook,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
