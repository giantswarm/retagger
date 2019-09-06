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

	"github.com/giantswarm/retagger/pkg/config"
	"github.com/giantswarm/retagger/pkg/registry"
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

	var conf *config.Config
	{
		conf, err = config.FromFile(r.flag.ConfigFile)
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

	for _, image := range conf.Images {
		for _, tag := range image.Tags {
			imageName := image.Name
			if image.OverrideRepoName != "" {
				r.logger.LogCtx(ctx, "message", "level", "debug", fmt.Sprintf("Override Name specified. Using %s as mirrored image name", image.OverrideRepoName))
				imageName = image.OverrideRepoName
			}
			r.logger.LogCtx(ctx, "message", "level", "debug", fmt.Sprintf("managing: %v, %v, %v", imageName, tag.Sha, tag.Tag))

			for _, customImage := range tag.CustomImages {
				ok, err := destRegistry.CheckImageTagExists(imageName, tag.Tag)
				if ok {
					r.logger.LogCtx(ctx, "message", "level", "debug", fmt.Sprintf("rebuilt image %q with tag %q already exists, skipping", imageName, fmt.Sprintf("%s-%s", tag.Tag, customImage.TagSuffix)))
					continue
				} else if err != nil {
					return microerror.Maskf(err, "could not check image %q and tag %q: %v", imageName, tag.Tag, err)
				} else {
					r.logger.LogCtx(ctx, "message", "level", "debug", fmt.Sprintf("rebuilt image %q with tag %q does not exists", imageName, fmt.Sprintf("%s-%s", tag.Tag, customImage.TagSuffix)))
				}
				rebuiltImageTag, err := destRegistry.Rebuild(imageName, tag.Tag, customImage)
				if err != nil {
					return microerror.Maskf(err, "could not rebuild image")
				}

				r.logger.LogCtx(ctx, "message", "level", "debug", fmt.Sprintf("pushing rebuilt custom image %s-%s", tag.Tag, customImage.TagSuffix))
				push := exec.Command("docker", "push", rebuiltImageTag)
				if err := Run(push); err != nil {
					return microerror.Maskf(err, "could not push image")
				}
			}

			ok, err := destRegistry.CheckImageTagExists(imageName, tag.Tag)
			if ok {
				r.logger.LogCtx(ctx, "message", "level", "debug", fmt.Sprintf("retagged image %q with tag %q already exists, skipping", imageName, tag.Tag))
				continue
			} else if err != nil {
				return microerror.Maskf(err, "could not check image %q and tag %q: %v", imageName, tag.Tag, err)
			} else {
				r.logger.LogCtx(ctx, "message", "level", "debug", fmt.Sprintf("retagged image %q with tag %q does not exist", imageName, tag.Tag))
			}

			shaName := config.ShaName(image.Name, tag.Sha)

			r.logger.LogCtx(ctx, "message", "level", "debug", fmt.Sprintf("pulling original image"))
			pullOriginal := exec.Command("docker", "pull", shaName)
			if err := Run(pullOriginal); err != nil {
				return microerror.Maskf(err, "could not pull image")
			}

			retaggedNameWithTag, err := destRegistry.Retag(imageName, shaName, tag.Tag)
			if err != nil {
				return microerror.Maskf(err, "could not retag image")
			}

			r.logger.LogCtx(ctx, "message", "level", "debug", fmt.Sprintf("pushing retagged image"))
			push := exec.Command("docker", "push", retaggedNameWithTag)
			if err := Run(push); err != nil {
				return microerror.Maskf(err, "could not push image")
			}
		}
	}

	return nil
}

func Run(c *exec.Cmd) error {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
