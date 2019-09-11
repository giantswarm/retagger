package retagger

import (
	"fmt"
	"os/exec"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/retagger/pkg/images"
	"github.com/giantswarm/retagger/pkg/registry"
)

type Config struct {
	Logger   micrologger.Logger
	Registry *registry.Registry
}

type Retagger struct {
	logger   micrologger.Logger
	registry *registry.Registry

	jobs []Job
}

func New(config Config) (*Retagger, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Registry == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Registry must not be empty", config)
	}

	r := &Retagger{
		logger:   config.Logger,
		registry: config.Registry,
		jobs:     []Job{},
	}
	return r, nil
}

func (r *Retagger) LoadImages(images images.Images) (int, error) {
	jobs, err := FromConfig(images)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	r.jobs = append(r.jobs, jobs...)

	return len(jobs), nil
}

func (r *Retagger) RetagImages(images []images.Image) error {
	for _, image := range images {
		err := r.RetagImage(image)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Retagger) RetagImage(image images.Image) error {
	for _, tag := range image.Tags {
		err := r.handleImageTag(image, tag)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Retagger) handleImageTag(image images.Image, tag images.Tag) error {
	imageName := image.Name
	if image.OverrideRepoName != "" {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("Override Name specified. Using %s as mirrored image name", image.OverrideRepoName))
		imageName = image.OverrideRepoName
	}
	r.logger.Log("level", "debug", "message", fmt.Sprintf("managing: %v, %v, %v", imageName, tag.Sha, tag.Tag))

	for _, customImage := range tag.CustomImages {
		ok, err := r.registry.CheckImageTagExists(imageName, tag.Tag)
		if ok {
			r.logger.Log("level", "debug", "message", fmt.Sprintf("rebuilt image %q with tag %q already exists, skipping", imageName, fmt.Sprintf("%s-%s", tag.Tag, customImage.TagSuffix)))
			continue
		} else if err != nil {
			return microerror.Maskf(err, "could not check image %q and tag %q: %v", imageName, tag.Tag, err)
		} else {
			r.logger.Log("level", "debug", "message", fmt.Sprintf("rebuilt image %q with tag %q does not exists", imageName, fmt.Sprintf("%s-%s", tag.Tag, customImage.TagSuffix)))
		}
		rebuiltImageTag, err := r.registry.Rebuild(imageName, tag.Tag, customImage)
		if err != nil {
			return microerror.Maskf(err, "could not rebuild image")
		}

		r.logger.Log("level", "debug", "message", fmt.Sprintf("pushing rebuilt custom image %s-%s", tag.Tag, customImage.TagSuffix))
		push := exec.Command("docker", "push", rebuiltImageTag)
		if err := Run(push); err != nil {
			return microerror.Maskf(err, "could not push image")
		}
	}

	ok, err := r.registry.CheckImageTagExists(imageName, tag.Tag)
	if ok {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("retagged image %q with tag %q already exists, skipping", imageName, tag.Tag))
		return nil
	} else if err != nil {
		return microerror.Maskf(err, "could not check image %q and tag %q: %v", imageName, tag.Tag, err)
	} else {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("retagged image %q with tag %q does not exist", imageName, tag.Tag))
	}

	shaName := images.ShaName(image.Name, tag.Sha)

	r.logger.Log("level", "debug", "message", fmt.Sprintf("pulling original image"))
	pullOriginal := exec.Command("docker", "pull", shaName)
	if err := Run(pullOriginal); err != nil {
		return microerror.Maskf(err, "could not pull image")
	}

	retaggedNameWithTag, err := r.registry.Retag(imageName, shaName, tag.Tag)
	if err != nil {
		return microerror.Maskf(err, "could not retag image")
	}

	r.logger.Log("level", "debug", "message", fmt.Sprintf("pushing retagged image"))
	push := exec.Command("docker", "push", retaggedNameWithTag)
	if err := Run(push); err != nil {
		return microerror.Maskf(err, "could not push image")
	}

	return nil
}
