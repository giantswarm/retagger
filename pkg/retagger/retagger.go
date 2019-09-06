package retagger

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/retagger/pkg/config"
)

type Config struct {
	Logger micrologger.Logger
}

type Retagger struct {
	logger micrologger.Logger
}

func New(config Config) (*Retagger, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Retagger{
		logger: config.Logger,
	}
	return r, nil
}

func (r *Retagger) RetagImages(images []config.Image) error {
	return nil
}

func (r *Retagger) RetagImage(image config.Image) error {
	return nil
}
