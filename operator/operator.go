package operator

import (
	"github.com/vshn/appcat-service-postgresql/operator/config"
	"github.com/vshn/appcat-service-postgresql/operator/standalone"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupControllers creates all Postgresql controllers with the supplied logger and adds them to the supplied manager.
func SetupControllers(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		config.SetupController,
		standalone.SetupController,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhooks creates all Postgresql webhooks with the supplied logger and adds them to the supplied manager.
func SetupWebhooks(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		standalone.SetupWebhook,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
