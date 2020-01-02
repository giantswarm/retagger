package retagger

import (
	"fmt"

	"github.com/giantswarm/microerror"
)

// SingleJob is a concrete job which can be executed.
type SingleJob struct {
	SourceImage string
	SourceTag   string
	SourceSha   string

	Destination Destination

	Options JobOptions
}

// Compile wraps this job in an array to keep consistency with the CompilableJob interface
func (job *SingleJob) Compile(r *Retagger) ([]SingleJob, error) {
	return []SingleJob{*job}, nil
}

// Describe returns a string containing basic information about the job.
func (job *SingleJob) Describe() string {
	return fmt.Sprintf("%s:%s will be tagged as %s:%s with digest %s",
		job.SourceImage, job.SourceTag, job.Destination.Image, job.Destination.Tag, job.SourceSha)
}

// Execute runs the job using the given Retagger instance
func (job *SingleJob) Execute(r *Retagger) error {
	// r.logger.Log("level", "debug", "message", fmt.Sprintf("Executing %v#", job))
	return r.executeSingleJob(*job)
}

// ShouldRetag examines the state of the job and the given Retagger's registry and returns whether the job should be run.
func (job *SingleJob) ShouldRetag(r *Retagger) (bool, error) {

	tagExists, err := r.registry.CheckImageTagExists(job.Destination.Image, job.Destination.Tag)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return (!tagExists || job.Options.UpdateOnChange), nil
}

// SingleJobFromJobDefinition converts a JobDefinition into a SingleJob
func SingleJobFromJobDefinition(j *JobDefinition, r *Retagger) *SingleJob {
	job := &SingleJob{
		SourceImage: j.SourceImage,
		SourceTag:   j.SourceTag,
		SourceSha:   j.SourceSha,

		Options: j.Options,
	}
	job.Destination = GetDestinationForJob(job, r)
	return job
}

// GetDestinationForJob populates a job's Destination information based on the job's Options.
func GetDestinationForJob(j *SingleJob, r *Retagger) Destination {
	var destinationImage string
	{
		if j.Options.OverrideRepoName == "" {
			destinationImage = r.registry.RetaggedName(j.SourceImage)
		} else {
			destinationImage = r.registry.RetaggedName(j.Options.OverrideRepoName)
		}
	}

	var destinationTag string
	{
		if j.Options.TagSuffix == "" {
			destinationTag = j.SourceTag
		} else {
			destinationTag = fmt.Sprintf("%s-%s", j.SourceTag, j.Options.TagSuffix)
		}
	}
	return Destination{
		Image: destinationImage,
		Tag:   destinationTag,
	}
}
