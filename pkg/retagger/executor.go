package retagger

import (
	"fmt"

	"github.com/giantswarm/microerror"
)

func (r *Retagger) ExecuteJobs() error {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("start executing %d jobs", len(r.jobs)))

	for _, j := range r.jobs {
		err := r.executeJob(j)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Retagger) executeJob(job Job) error {
	var destinationImage string
	{
		if job.Options.OverrideRepoName == "" {
			destinationImage = r.destinationRegistry.RetaggedName(job.SourceImage)
		} else {
			destinationImage = r.destinationRegistry.RetaggedName(job.Options.OverrideRepoName)
		}
	}

	var destinationTag string
	{
		if job.Options.TagSuffix == "" {
			destinationTag = job.SourceTag
		} else {
			destinationTag = fmt.Sprintf("%s-%s", job.SourceTag, job.Options.TagSuffix)
		}
	}

	ok, err := r.destinationRegistry.CheckImageTagExists(destinationImage, destinationTag)
	if err != nil {
		return microerror.Mask(err)
	}
	if ok {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("image %s:%s already exists, skipping it now", destinationImage, destinationTag))
		return nil
	}

	return nil
}
