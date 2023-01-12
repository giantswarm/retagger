package retagger

import (
	"fmt"

	"github.com/giantswarm/microerror"
)

// CompileJobs compiles all jobs in this Retagger's list of jobs into a list of concrete SingleJobs.
func (r *Retagger) CompileJobs() error {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("compiling %d job definitions", len(r.jobs)))

	var compiledJobs []SingleJob
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

// ExecuteJobs runs the jobs associated with this Retagger.
func (r *Retagger) ExecuteJobs() error {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("start executing %d jobs", len(r.compiledJobs)))

	if r.dryrun {
		r.logger.Log("level", "info", "message", "Retagger is in --dry-run mode. Listing jobs, but not running them.")
	}

	if conflictErrors := checkConflicts(r.compiledJobs); len(conflictErrors) > 0 {
		// These are warnings since we already had lots of conflicts on purpose in our configuration.
		// For example, if we choose a fixed SHA for `alpine:3.14` and the upstream tag `3.14` gets updated to a
		// newer image, it's a conflict and we don't want to fail for such normal use cases.
		for _, conflictError := range conflictErrors {
			r.logger.Log("level", "warn", "message", fmt.Sprintf("Found conflict: %s", conflictError))

		}
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

// executeSingleJob runs one job definition.
func (r *Retagger) executeSingleJob(job SingleJob) error {

	shouldTag, err := job.ShouldRetag(r)
	if err != nil {
		return microerror.Mask(err)
	}

	if shouldTag {
		r.logger.Log("level", "debug", "message", "pulling original image")

		err = r.registry.PullImage(job.Source.Image, job.Source.SHA)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Log("level", "debug", "message", "pulled original image")

		if job.Options.DockerfileOptions != nil && len(job.Options.DockerfileOptions) > 0 {
			_, err = r.registry.RebuildImage(job.Source.Image, job.Source.SHA, job.Destination.Image, job.Destination.Tag, job.Options.DockerfileOptions)
			if err != nil {
				return microerror.Mask(err)
			}
		} else {
			_, err = r.registry.TagSha(job.Source.Image, job.Source.SHA, job.Destination.Image, job.Destination.Tag)
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
