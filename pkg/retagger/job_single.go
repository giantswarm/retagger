package retagger

import (
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
)

// SingleJob is a concrete job which can be executed.
type SingleJob struct {
	Source Source

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
		job.Source.Image, job.Source.Tag, job.Destination.Image, job.Destination.Tag, job.Source.SHA)
}

// Execute runs the job using the given Retagger instance
func (job *SingleJob) Execute(r *Retagger) error {
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
func SingleJobFromJobDefinition(jobDef *JobDefinition, r *Retagger) *SingleJob {
	job := &SingleJob{
		Source: GetSourceForJob(jobDef, r),

		Options: jobDef.Options,
	}
	job.Destination = GetDestinationForJob(job, r)
	return job
}

func getRepoHostForJob(j *JobDefinition, r *Retagger) string {
	// Handle remote, Docker Hub, and Docker library image path formats.
	registryPath, err := r.registry.GuessRegistryPath(j.SourceImage)
	if err != nil {
		return j.SourceImage // Fall back to trying to use given image name
	}
	return registryPath.Hostname()
}

func getFullImageNameForJob(j *JobDefinition, r *Retagger) string {
	registryPath, err := r.registry.GuessRegistryPath(j.SourceImage)
	if err != nil {
		return j.SourceImage // Fall back to trying to use given image name
	}
	return strings.Trim(registryPath.Path, "/") // Remove leading slash
}

// GetSourceForJob populates a Source object based on the given JobDefinition.
func GetSourceForJob(jobDef *JobDefinition, r *Retagger) Source {
	return Source{
		Image:         jobDef.SourceImage,
		SHA:           jobDef.SourceSha,
		Tag:           jobDef.SourceTag,
		RepoPath:      getRepoHostForJob(jobDef, r),
		FullImageName: getFullImageNameForJob(jobDef, r),
	}
}

// GetDestinationForJob populates a job's Destination information based on the job's Options.
func GetDestinationForJob(j *SingleJob, r *Retagger) Destination {
	var destinationImage string
	{
		if j.Options.OverrideRepoName == "" {
			destinationImage = r.registry.RetaggedName(j.Source.Image)
		} else {
			destinationImage = r.registry.RetaggedName(j.Options.OverrideRepoName)
		}
	}

	var destinationTag string
	{
		if j.Options.TagSuffix == "" {
			destinationTag = j.Source.Tag
		} else {
			destinationTag = fmt.Sprintf("%s-%s", j.Source.Tag, j.Options.TagSuffix)
		}
	}
	return Destination{
		Image: destinationImage,
		Tag:   destinationTag,
	}
}
