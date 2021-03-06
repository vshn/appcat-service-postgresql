package main

import (
	"context"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
	helmreleasev1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	helmconfigv1beta1 "github.com/crossplane-contrib/provider-helm/apis/v1beta1"
	"github.com/urfave/cli/v2"
	"github.com/vshn/appcat-service-postgresql/apis"
	"github.com/vshn/appcat-service-postgresql/operator"
	"github.com/vshn/appcat-service-postgresql/operator/standalone"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type operatorCommand struct {
	SyncInterval          time.Duration
	LeaderElectionEnabled bool
	WebhookCertDir        string

	manager    manager.Manager
	kubeconfig *rest.Config
}

var operatorCommandName = "operator"

func newOperatorCommand() *cli.Command {
	command := &operatorCommand{}
	return &cli.Command{
		Name:   operatorCommandName,
		Usage:  "Start provider in operator mode",
		Before: command.validate,
		Action: command.execute,
		Flags: []cli.Flag{
			&cli.DurationFlag{Name: "sync-interval", Value: time.Hour, EnvVars: envVars("SYNC_INTERVAL"),
				Usage:       "How often all resources will be double-checked for drift from the desired state.",
				Destination: &command.SyncInterval,
			},
			&cli.BoolFlag{Name: "leader-election-enabled", Value: false, EnvVars: envVars("LEADER_ELECTION_ENABLED"),
				Usage:       "Use leader election for the controller manager.",
				Destination: &command.LeaderElectionEnabled,
			},
			&cli.StringFlag{Name: "webhook-tls-cert-dir", EnvVars: []string{"WEBHOOK_TLS_CERT_DIR"},
				Usage:       "Directory containing the certificates for the webhook server. If empty, the webhook server is not started.",
				Destination: &command.WebhookCertDir,
			},
			&cli.StringFlag{Name: "operator-namespace", EnvVars: []string{"OPERATOR_NAMESPACE"},
				Usage:       "OperatorNamespace name where the operator runs in.",
				Destination: &standalone.OperatorNamespace, Required: true,
			},
			&cli.StringFlag{Name: "service-namespace-prefix", EnvVars: envVars("SERVICE_NAMESPACE_PREFIX"),
				Usage: "Prefix of namespaces where the actual PostgreSQL deployments are deployed in.",
				Value: standalone.ServiceNamespacePrefix, Destination: &standalone.ServiceNamespacePrefix,
			},
		},
	}
}

func (c *operatorCommand) validate(ctx *cli.Context) error {
	_ = LogMetadata(ctx)
	log := AppLogger(ctx).WithName(operatorCommandName)
	log.V(1).Info("validating config")
	return nil
}

func (c *operatorCommand) execute(ctx *cli.Context) error {
	log := AppLogger(ctx).WithName(operatorCommandName)
	log.Info("Setting up controllers", "config", c)
	ctrl.SetLogger(log)

	p := pipeline.NewPipeline().WithBeforeHooks([]pipeline.Listener{
		func(step pipeline.Step) {
			log.V(1).Info(step.Name)
		},
	})
	p.AddStepFromFunc("get config", func(ctx context.Context) error {
		cfg, err := ctrl.GetConfig()
		c.kubeconfig = cfg
		return err
	})
	p.AddStepFromFunc("create manager", func(ctx context.Context) error {
		// configure client-side throttling
		c.kubeconfig.QPS = 100
		c.kubeconfig.Burst = 150 // more Openshift friendly

		mgr, err := ctrl.NewManager(c.kubeconfig, ctrl.Options{
			SyncPeriod: &c.SyncInterval,
			// controller-runtime uses both ConfigMaps and Leases for leader election by default.
			// Leases expire after 15 seconds, with a 10-second renewal deadline.
			// We've observed leader loss due to renewal deadlines being exceeded when under high load - i.e.
			//  hundreds of reconciles per second and ~200rps to the API server.
			// Switching to Leases only and longer leases appears to alleviate this.
			LeaderElection:             c.LeaderElectionEnabled,
			LeaderElectionID:           "crossplane-leader-election-provider-appcat-postgresql",
			LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
			LeaseDuration:              func() *time.Duration { d := 60 * time.Second; return &d }(),
			RenewDeadline:              func() *time.Duration { d := 50 * time.Second; return &d }(),
		})
		c.manager = mgr
		return err
	})
	p.AddStep(pipeline.NewPipeline().WithNestedSteps("register schemes",
		pipeline.NewStepFromFunc("register API schemes", func(ctx context.Context) error {
			return apis.AddToScheme(c.manager.GetScheme())
		}),
		pipeline.NewStepFromFunc("register helm config scheme", func(ctx context.Context) error {
			return helmconfigv1beta1.SchemeBuilder.AddToScheme(c.manager.GetScheme())
		}),
		pipeline.NewStepFromFunc("register helm release scheme", func(ctx context.Context) error {
			return helmreleasev1beta1.SchemeBuilder.AddToScheme(c.manager.GetScheme())
		}),
		pipeline.NewStepFromFunc("register k8up scheme", func(ctx context.Context) error {
			return k8upv1.AddToScheme(c.manager.GetScheme())
		}),
	))
	p.AddStepFromFunc("setup controllers", func(ctx context.Context) error {
		return operator.SetupControllers(c.manager)
	})
	p.AddStep(pipeline.ToStep("setup webhook server",
		func(ctx context.Context) error {
			ws := c.manager.GetWebhookServer()
			ws.CertDir = c.WebhookCertDir
			ws.TLSMinVersion = "1.3"
			return operator.SetupWebhooks(c.manager)
		},
		pipeline.Bool(c.WebhookCertDir != "")))
	p.AddStepFromFunc("run manager", func(ctx context.Context) error {
		log.Info("Starting manager")
		return c.manager.Start(ctx)
	})
	return p.RunWithContext(ctx.Context).Err()
}
