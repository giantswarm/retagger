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

	r.logger.Log("level", "debug", "message", fmt.Sprintf("successfully finished executing %d jobs", len(r.jobs)))

	return nil
}

func (r *Retagger) executeJob(job Job) error {
	var destinationImage string
	{
		if job.Options.OverrideRepoName == "" {
			destinationImage = r.registry.RetaggedName(job.SourceImage)
		} else {
			destinationImage = r.registry.RetaggedName(job.Options.OverrideRepoName)
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

	r.logger.Log("level", "debug", "message", fmt.Sprintf("executing: %v, %v with options %#v", job.SourceImage, job.SourceTag, job.Options))

	exists, err := r.registry.CheckImageTagExists(destinationImage, destinationTag)
	if err != nil {
		return microerror.Mask(err)
	}
	if exists {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("image %s:%s already exists, skipping it now", destinationImage, destinationTag))
		return nil
	}

	r.logger.Log("level", "debug", "message", fmt.Sprintf("pulling original image"))

	err = r.registry.PullImage(job.SourceImage, job.SourceSha)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("level", "debug", "message", fmt.Sprintf("pulled original image"))

	if job.Options.DockerfileOptions != nil && len(job.Options.DockerfileOptions) > 0 {
		_, err = r.registry.RebuildImage(job.SourceImage, job.SourceSha, destinationImage, destinationTag, job.Options.DockerfileOptions)
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		_, err = r.registry.TagSha(job.SourceImage, job.SourceSha, destinationImage, destinationTag)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = r.registry.PushImage(destinationImage, destinationTag)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
