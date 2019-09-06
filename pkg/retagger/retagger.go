package retagger

import (
	"fmt"
	"os/exec"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/retagger/pkg/config"
	"github.com/giantswarm/retagger/pkg/registry"
)

type Config struct {
	Logger              micrologger.Logger
	DestinationRegistry *registry.Registry
}

type Retagger struct {
	logger              micrologger.Logger
	destinationRegistry *registry.Registry
}

func New(config Config) (*Retagger, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.DestinationRegistry == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.DestinationRegistry must not be empty", config)
	}

	r := &Retagger{
		logger:              config.Logger,
		destinationRegistry: config.DestinationRegistry,
	}
	return r, nil
}

func (r *Retagger) RetagImages(images []config.Image) error {
	for _, image := range images {
		err := r.RetagImage(image)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Retagger) RetagImage(image config.Image) error {
	for _, tag := range image.Tags {
		err := r.handleImageTag(image, tag)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Retagger) handleImageTag(image config.Image, tag config.Tag) error {
	imageName := image.Name
	if image.OverrideRepoName != "" {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("Override Name specified. Using %s as mirrored image name", image.OverrideRepoName))
		imageName = image.OverrideRepoName
	}
	r.logger.Log("level", "debug", "message", fmt.Sprintf("managing: %v, %v, %v", imageName, tag.Sha, tag.Tag))

	for _, customImage := range tag.CustomImages {
		ok, err := r.destinationRegistry.CheckImageTagExists(imageName, tag.Tag)
		if ok {
			r.logger.Log("message", "level", "debug", fmt.Sprintf("rebuilt image %q with tag %q already exists, skipping", imageName, fmt.Sprintf("%s-%s", tag.Tag, customImage.TagSuffix)))
			continue
		} else if err != nil {
			return microerror.Maskf(err, "could not check image %q and tag %q: %v", imageName, tag.Tag, err)
		} else {
			r.logger.Log("level", "debug", "message", fmt.Sprintf("rebuilt image %q with tag %q does not exists", imageName, fmt.Sprintf("%s-%s", tag.Tag, customImage.TagSuffix)))
		}
		rebuiltImageTag, err := r.destinationRegistry.Rebuild(imageName, tag.Tag, customImage)
		if err != nil {
			return microerror.Maskf(err, "could not rebuild image")
		}

		r.logger.Log("level", "debug", "message", fmt.Sprintf("pushing rebuilt custom image %s-%s", tag.Tag, customImage.TagSuffix))
		push := exec.Command("docker", "push", rebuiltImageTag)
		if err := Run(push); err != nil {
			return microerror.Maskf(err, "could not push image")
		}
	}

	ok, err := r.destinationRegistry.CheckImageTagExists(imageName, tag.Tag)
	if ok {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("retagged image %q with tag %q already exists, skipping", imageName, tag.Tag))
		return nil
	} else if err != nil {
		return microerror.Maskf(err, "could not check image %q and tag %q: %v", imageName, tag.Tag, err)
	} else {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("retagged image %q with tag %q does not exist", imageName, tag.Tag))
	}

	shaName := config.ShaName(image.Name, tag.Sha)

	r.logger.Log("level", "debug", "message", fmt.Sprintf("pulling original image"))
	pullOriginal := exec.Command("docker", "pull", shaName)
	if err := Run(pullOriginal); err != nil {
		return microerror.Maskf(err, "could not pull image")
	}

	retaggedNameWithTag, err := r.destinationRegistry.Retag(imageName, shaName, tag.Tag)
	if err != nil {
		return microerror.Maskf(err, "could not retag image")
	}

	r.logger.Log("message", "level", "debug", "message", fmt.Sprintf("pushing retagged image"))
	push := exec.Command("docker", "push", retaggedNameWithTag)
	if err := Run(push); err != nil {
		return microerror.Maskf(err, "could not push image")
	}

	return nil
}
