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

	pullOriginal := exec.Command("docker", "pull", shaName)
	if err := Run(pullOriginal); err != nil {
		return microerror.Maskf(err, "could not pull image")
	}

	return nil
}
