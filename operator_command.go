package main

import (
	"os"
	"sync"

	"github.com/urfave/cli/v2"
)

type operatorCommand struct {
}

var operatorCommandName = "operator"

func newOperatorCommand() *cli.Command {
	command := &operatorCommand{}
	return &cli.Command{
		Name:   operatorCommandName,
		Usage:  "Start provider in operator mode",
		Before: command.validate,
		Action: command.execute,
	}
}

func (c *operatorCommand) validate(context *cli.Context) error {
	_ = LogMetadata(context)
	log := AppLogger(context).WithName(operatorCommandName)
	log.V(1).Info("validating config")
	return nil
}

func (c *operatorCommand) execute(context *cli.Context) error {
	log := AppLogger(context).WithName(operatorCommandName)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		// Shutdown hook. Can be used to gracefully shutdown listeners or pre-shutdown cleanup.
		// Can be removed if not needed.
		// Please note that this example is incomplete and doesn't cover all cases when properly implementing shutdowns.
		defer wg.Done()
		<-context.Done()
		err := c.shutdown(context)
		if err != nil {
			log.Error(err, "cannot properly shut down")
			os.Exit(2)
		}
	}()
	log.Info("Hello from operator command!", "config", c)
	wg.Wait()
	return nil
}

func (c *operatorCommand) shutdown(context *cli.Context) error {
	log := AppLogger(context).WithName(operatorCommandName)
	log.Info("Shutting down operator command")
	return nil
}
