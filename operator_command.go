package main

import (
	"context"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/urfave/cli/v2"
	"github.com/vshn/appcat-service-postgresql/apis"
	"github.com/vshn/appcat-service-postgresql/operator"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
			&cli.StringFlag{Name: "webhook-tls-cert-dir", EnvVars: []string{"WEBHOOK_TLS_CERT_DIR"}, // Env var is set by Crossplane
				Usage:       "Directory containing the certificates for the webhook server. If empty, the webhook server is not started.",
				Destination: &command.WebhookCertDir,
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
	p.AddStepFromFunc("register schemes", func(ctx context.Context) error {
		return apis.AddToScheme(c.manager.GetScheme())
	})
	p.AddStepFromFunc("setup controllers", func(ctx context.Context) error {
		o := controller.Options{
			Log: log,
		}
		return operator.SetupControllers(c.manager, o)
	})
	p.AddStep(pipeline.ToStep("setup webhook server",
		func(ctx context.Context) error {
			ws := c.manager.GetWebhookServer()
			ws.CertDir = c.WebhookCertDir
			ws.TLSMinVersion = "1.3"
			o := controller.Options{Log: log}
			return operator.SetupWebhooks(c.manager, o)
		},
		pipeline.Bool(c.WebhookCertDir != "")))
	p.AddStepFromFunc("run manager", func(ctx context.Context) error {
		log.Info("Starting manager")
		return c.manager.Start(ctx)
	})
	return p.RunWithContext(ctx.Context).Err()
}
