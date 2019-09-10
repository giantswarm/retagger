package cmd

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/retagger/pkg/images"
	"github.com/giantswarm/retagger/pkg/registry"
	"github.com/giantswarm/retagger/pkg/retagger"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	stdout io.Writer
	stderr io.Writer
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	var conf *images.Config
	{
		conf, err = images.FromFile(r.flag.ConfigFile)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var destRegistry *registry.Registry
	{
		c := registry.Config{
			Host:         r.flag.Host,
			Organisation: r.flag.Organization,
			Password:     r.flag.Password,
			Username:     r.flag.Username,
			LogFunc:      nil,
		}
		destRegistry, err = registry.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = destRegistry.Login()
	if err != nil {
		return microerror.Mask(err)
	}

	var newRetagger *retagger.Retagger
	{
		c := retagger.Config{
			Logger:              r.logger,
			DestinationRegistry: destRegistry,
		}
		newRetagger, err = retagger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = newRetagger.RetagImages(conf.Images)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func Run(c *exec.Cmd) error {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
