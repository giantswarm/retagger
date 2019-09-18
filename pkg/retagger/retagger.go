package retagger

import (
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
	jobs, err := FromImages(images)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	r.jobs = append(r.jobs, jobs...)

	return len(jobs), nil
}
