package retagger

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/retagger/pkg/images"
	"github.com/giantswarm/retagger/pkg/registry"
)

// Config contains configuration values for Retagger itself
type Config struct {
	Logger   micrologger.Logger
	Registry *registry.Registry
	DryRun   bool
}

// Retagger allows retagging external Docker images into the specified internal registry.
type Retagger struct {
	logger   micrologger.Logger
	registry *registry.Registry
	dryrun   bool

	jobs         []JobRequest
	compiledJobs []SingleJob
}

// New creates a new instance of Retagger based on the given Config
func New(config Config) (*Retagger, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Registry == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Registry must not be empty", config)
	}

	r := &Retagger{
		logger:       config.Logger,
		registry:     config.Registry,
		dryrun:       config.DryRun,
		jobs:         []JobRequest{},
		compiledJobs: []SingleJob{},
	}
	return r, nil
}

// LoadImages populates Retagger's job list with jobs defined in the given image list
func (r *Retagger) LoadImages(images images.Images) (int, error) {
	jobs, err := FromImages(images)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	r.jobs = append(r.jobs, jobs...)

	return len(jobs), nil
}
