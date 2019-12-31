package retagger

import (
	"fmt"

	"github.com/giantswarm/microerror"
)

// CompileJobs compiles all jobs in this Retagger's list of jobs into a list of concrete SingleJobs.
func (r *Retagger) CompileJobs() error {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("compiling %d job definitions", len(r.jobs)))

	// TODO: Gate for not performing API calls during compile?
	compiledJobs := []SingleJob{}
	for _, j := range r.jobs {
		compiled, err := j.Compile(r)
		if err != nil {
			return microerror.Mask(err)
		}
		compiledJobs = append(compiledJobs, compiled...)
	}

	r.compiledJobs = append(r.compiledJobs, compiledJobs...)

	r.logger.Log("level", "debug", "message", fmt.Sprintf("compiled %d jobs", len(r.compiledJobs)))

	return nil
}

// ExecuteJobs runs the jobs associated with this Retagger
func (r *Retagger) ExecuteJobs() error {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("start executing %d jobs", len(r.compiledJobs)))

	if r.dryrun {
		r.logger.Log("level", "info", "message", "Retagger is in --dry-run mode. Listing jobs, but not running them.")
	}

	for _, j := range r.compiledJobs {
		if r.dryrun {
			r.logger.Log("level", "info", "message", fmt.Sprintf("Dry-Run: %s", j.Describe()))
		} else {
			err := j.Execute(r)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	r.logger.Log("level", "debug", "message", fmt.Sprintf("successfully finished executing %d jobs", len(r.compiledJobs)))

	return nil
}

// executeSingleJob runs one job definition, optionally skipping jobs with tags which already exist
func (r *Retagger) executeSingleJob(job SingleJob) error {

	shouldTag, err := job.ShouldRetag(r)
	if err != nil {
		return microerror.Mask(err)
	}

	if shouldTag {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("pulling original image"))

		err = r.registry.PullImage(job.SourceImage, job.SourceSha)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Log("level", "debug", "message", fmt.Sprintf("pulled original image"))

		if job.Options.DockerfileOptions != nil && len(job.Options.DockerfileOptions) > 0 {
			_, err = r.registry.RebuildImage(job.SourceImage, job.SourceSha, job.Destination.Image, job.Destination.Tag, job.Options.DockerfileOptions)
			if err != nil {
				return microerror.Mask(err)
			}
		} else {
			_, err = r.registry.TagSha(job.SourceImage, job.SourceSha, job.Destination.Image, job.Destination.Tag)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		err = r.registry.PushImage(job.Destination.Image, job.Destination.Tag)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
