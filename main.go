package main

import (
	"context"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/retagger/cmd"
)

func main() {
	err := mainE(context.Background())
	if err != nil {
		panic(microerror.Stack(err))
	}
}

func mainE(ctx context.Context) error {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var rootCommand *cobra.Command
	{
		c := cmd.Config{
			Logger: logger,

			GitCommit: "%",
			Source:    "%",
		}
		rootCommand, err = cmd.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = rootCommand.Execute()
	if err != nil {
		logger.LogCtx(ctx, "level", "error", "message", "failed to execute command", "stack", microerror.Stack(err))
		os.Exit(1)
	}

	return nil
}
