package retagger

import (
	"fmt"
	"regexp"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/retagger/pkg/images"
	dockerRegistry "github.com/nokia/docker-registry-client/registry"
)

// CompilableJob represents any Job which can be Compiled.
type CompilableJob interface {
	Compile(*Retagger) ([]SingleJob, error)
}

// ExecutableJob represents any Job which can be Executed.
type ExecutableJob interface {
	Execute(r *Retagger) error
}

// Destination contains information about the target repository and tag of a job.
type Destination struct {
	Image string
	Tag   string
}

// SingleJob is a concrete job which can be executed.
type SingleJob struct {
	SourceImage string
	SourceTag   string
	SourceSha   string

	Destination Destination

	Options JobOptions
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

// SingleJobFromJobRequest converts a JobRequest into a SingleJob
func SingleJobFromJobRequest(j *JobRequest, r *Retagger) *SingleJob {
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

// JobRequest represents a single or pattern job which has been read from the input file but is yet to be compiled.
type JobRequest struct { // Definition?
	SourceImage   string
	SourceTag     string
	SourceSha     string
	SourcePattern string

	Options JobOptions
}

// Compile expands a JobRequest into one or multiple concrete jobs.
func (jr *JobRequest) Compile(r *Retagger) ([]SingleJob, error) {
	// The job can either be a SingleJob or a PatternJob
	if jr.SourceSha != "" {
		if jr.SourcePattern != "" {
			// Configuration specified a SHA and a pattern -- use SHA to be safe, but warn about misconfiguration
			r.logger.Log("level", "warn", "message", fmt.Sprintf("invalid configuration: Job %v specifies both a SHA (%v) and a Pattern (%v). Using SHA", jr.SourceImage, jr.SourceSha, jr.SourcePattern))
		}

		// This is a single job -- return it on its own
		job := *SingleJobFromJobRequest(jr, r)
		job.Destination = GetDestinationForJob(&job, r)

		return []SingleJob{job}, nil
	}

	// If no SHA is given, treat this as a pattern job
	patternJob := PatternJobFromJobRequest(jr, r)
	nestedJobs, err := patternJob.Compile(r)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return nestedJobs, nil
}

// PatternJob contains a definition for generating multiple single jobs based on a pattern.
type PatternJob struct {
	SourceImage   string
	SourcePattern string

	Destination Destination

	Options JobOptions
}

// PatternJobFromJobRequest converts a JobRequest into a PatternJob
func PatternJobFromJobRequest(j *JobRequest, r *Retagger) *PatternJob {
	job := &PatternJob{
		SourceImage:   j.SourceImage,
		SourcePattern: j.SourcePattern,

		Options: j.Options,
	}
	return job
}

// Compile expands a PatternJob into one or multiple SingleJobs using the given Retagger instance.
func (job *PatternJob) Compile(r *Retagger) ([]SingleJob, error) {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("compiling jobs for image %v using pattern %v, with options %#v", job.SourceImage, job.SourcePattern, job.Options))

	// Make sure our pattern is valid.
	pattern, err := regexp.Compile(job.SourcePattern)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Get SHA/Tag pairs from our quay registry.
	quayTagMap, err := r.registry.GetQuayTagMap(job.SourceImage)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Handle remote, Docker Hub, and Docker library image path formats.
	registryPath, err := r.registry.GuessRegistryPath(job.SourceImage)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Create a reference to the external registry.
	o := dockerRegistry.Options{
		Logf: dockerRegistry.Quiet,
		// Logf:          dockerRegistry.Log,
		DoInitialPing: false,
	}

	externalRegistry, err := dockerRegistry.NewCustom(fmt.Sprintf("https://%s", registryPath.Hostname()), o)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	fullImageName := r.registry.GetRepositoryFromPath(registryPath)

	// Get the tags for this image from the external registry.
	externalRegistryTags, err := externalRegistry.Tags(fullImageName)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Find tags matching our configured pattern.
	matches := []string{}
	for _, t := range externalRegistryTags {
		if pattern.MatchString(t) {
			matches = append(matches, t)
		}
	}

	if len(matches) == 0 {
		r.logger.Log("level", "warn", "message", fmt.Sprintf("No upstream image tags were found matching the pattern %s", job.SourcePattern))
	} else {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("Found %d upstream tags which match the pattern %s", len(matches), job.SourcePattern))
	}

	jobs := []SingleJob{}
	// Find tags which need to be re-checked and updated.
	for _, match := range matches {
		sourceSHA := ""

		_, exists := quayTagMap[match]

		if !exists {
			// Tag is new - get SHA and tag it.
			newDigest, err := externalRegistry.ManifestDigest(fullImageName, match)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			sourceSHA = newDigest.String()

		} else {
			tag := quayTagMap[match]
			if job.Options.UpdateOnChange {
				// Tag exists, but we should update the image.

				newDigest, err := externalRegistry.ManifestDigest(fullImageName, tag.Name)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				if newDigest.String() != tag.ManifestDigest {
					// Retag this image with this tag.
					r.logger.Log("level", "debug", "message",
						fmt.Sprintf("Image %s:%s will be retagged to %s from %s",
							job.SourceImage, tag.Name, newDigest, tag.ManifestDigest))

					sourceSHA = newDigest.String()
				}

			} else {
				r.logger.Log("level", "debug", "message",
					fmt.Sprintf("Ignored: image %s:%s has changed but will not be retagged",
						job.SourceImage, tag.Name))
			}
		}

		if sourceSHA != "" {
			// Create job with new SHA.
			j := SingleJob{
				SourceTag:   match,
				SourceImage: job.SourceImage,
				SourceSha:   sourceSHA,

				Options: job.Options,
			}
			j.Destination = GetDestinationForJob(&j, r)
			jobs = append(jobs, j)
		}
	}

	r.logger.Log("level", "debug", "message", fmt.Sprintf("Compiled %d jobs to process", len(jobs)))

	return jobs, nil
}

// JobCompiler contains a Job which can be Compiled.
type JobCompiler struct {
	Job CompilableJob
}

// Compile takes a CompilableJob and a Retagger and Compiles the job.
func (jc *JobCompiler) Compile(job CompilableJob, r *Retagger) ([]SingleJob, error) {
	return job.Compile(r)
}

// JobOptions specifies optional features for modifying the behavior of the job during tagging.
type JobOptions struct {
	// DockerfileOptions - list of strings we add for Dockerfile to build custom image.
	DockerfileOptions []string

	TagSuffix string

	OverrideRepoName string

	// UpdateOnChange sets whether a pattern Job should update the destination image if a source image changes for a given tag
	UpdateOnChange bool
}

// FromImages receives a list of Images and converts them into JobRequests.
func FromImages(images images.Images) ([]JobRequest, error) {
	var jobs []JobRequest

	for _, i := range images {
		js, err := FromImage(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		jobs = append(jobs, js...)
	}

	return jobs, nil
}

// FromImage takes an Image and converts it into a JobRequest.
func FromImage(image images.Image) ([]JobRequest, error) {
	var jobs []JobRequest

	for _, t := range image.Tags {
		j, err := fromImageTagIncludeCustom(image, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		jobs = append(jobs, j...)
	}

	// TODO: Combine patterns into single job -- pull/check tag list only once
	for _, p := range image.Patterns {
		j, err := fromImageTagPatternIncludeCustom(image, p)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		jobs = append(jobs, j...)
	}

	return jobs, nil
}

func fromImageTagIncludeCustom(image images.Image, tag images.Tag) ([]JobRequest, error) {
	var jobs []JobRequest

	j, err := fromImageTag(image, tag)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	jobs = append(jobs, j)

	for _, c := range tag.CustomImages {
		j, err = fromImageTag(image, tag)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if c.TagSuffix != "" {
			j.Options.TagSuffix = c.TagSuffix
		}

		if c.DockerfileOptions != nil && len(c.DockerfileOptions) > 0 {
			j.Options.DockerfileOptions = c.DockerfileOptions
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}

func fromImageTagPatternIncludeCustom(image images.Image, pattern images.TagPattern) ([]JobRequest, error) {
	var jobs []JobRequest

	j, err := fromImageTagPattern(image, pattern)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	jobs = append(jobs, j)

	for _, c := range pattern.CustomImages {
		j, err = fromImageTagPattern(image, pattern)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if c.TagSuffix != "" {
			j.Options.TagSuffix = c.TagSuffix
		}

		if c.DockerfileOptions != nil && len(c.DockerfileOptions) > 0 {
			j.Options.DockerfileOptions = c.DockerfileOptions
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}

func fromImageTag(image images.Image, tag images.Tag) (JobRequest, error) {
	j := JobRequest{
		SourceImage: image.Name,
		SourceTag:   tag.Tag,
		SourceSha:   tag.Sha,
	}

	if image.OverrideRepoName != "" {
		j.Options.OverrideRepoName = image.OverrideRepoName
	}

	return j, nil
}

func fromImageTagPattern(image images.Image, tagPattern images.TagPattern) (JobRequest, error) {
	j := JobRequest{
		SourceImage:   image.Name,
		SourcePattern: tagPattern.Pattern,
	}

	if image.OverrideRepoName != "" {
		j.Options.OverrideRepoName = image.OverrideRepoName
	}

	if tagPattern.UpdateOnChange {
		j.Options.UpdateOnChange = true
	}

	return j, nil
}
