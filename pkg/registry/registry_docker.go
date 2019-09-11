package registry

import (
	"fmt"
	"os/exec"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/retagger/pkg/images"
)

func (r *Registry) PullImage(image string, sha string) error {
	if sha == "" {
		return microerror.Maskf(invalidArgumentError, "%s SHA should not be empty", image)
	}

	shaName := images.ShaName(image, sha)

	r.logger.Log("level", "debug", "message", fmt.Sprintf("executing: docker pull %s", shaName))
	pullOriginal := exec.Command("docker", "pull", shaName)
	if err := Run(pullOriginal); err != nil {
		return microerror.Maskf(err, "could not pull image")
	}

	return nil
}

func (r *Registry) TagSha(sourceImage, sha, destinationImage, destinationTag string) (string, error) {
	imageSha := images.ShaName(sourceImage, sha)
	retaggedNameWithTag := fmt.Sprintf("%s:%s", destinationImage, destinationTag)

	r.logger.Log("level", "debug", "message", fmt.Sprintf("executing: docker tag %s %s", sha, retaggedNameWithTag))
	retag := exec.Command("docker", "tag", imageSha, retaggedNameWithTag)
	err := Run(retag)
	if err != nil {
		return "", microerror.Mask(err)
	}
	return retaggedNameWithTag, nil
}

func (r *Registry) PushImage(destinationImage, destinationTag string) error {
	push := exec.Command("docker", "push", fmt.Sprintf("%s:%s", destinationImage, destinationTag))
	if err := Run(push); err != nil {
		return microerror.Maskf(err, "could not push image")
	}

	return nil
}
