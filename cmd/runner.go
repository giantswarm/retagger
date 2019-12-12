package cmd

import (
	"context"
	"fmt"
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

	var img images.Images
	{
		img, err = images.FromFile(r.flag.ConfigFile)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var newRegistry *registry.Registry
	{
		c := registry.Config{
			Host:         r.flag.Host,
			Organisation: r.flag.Organization,
			Password:     r.flag.Password,
			Username:     r.flag.Username,
			LogFunc:      nil,
			Logger:       r.logger,
		}
		newRegistry, err = registry.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = newRegistry.Login()
	if err != nil {
		return microerror.Mask(err)
	}

	var newRetagger *retagger.Retagger
	{
		c := retagger.Config{
			Logger:   r.logger,
			Registry: newRegistry,
			DryRun:   r.flag.DryRun,
		}
		newRetagger, err = retagger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	n, err := newRetagger.LoadImages(img)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.Log("level", "debug", "message", fmt.Sprintf("loaded %d jobs from YAML", n))

	err = newRetagger.ExecuteJobs()
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
