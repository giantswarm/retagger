package retagger

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/retagger/pkg/images"
)

// JobDefinition represents a single or pattern job which is yet to be compiled.
type JobDefinition struct {
	SourceImage   string
	SourceTag     string
	SourceSha     string
	SourcePattern string

	Options JobOptions
}

// Compile expands a JobDefinition into one or multiple concrete jobs.
func (jr *JobDefinition) Compile(r *Retagger) ([]SingleJob, error) {
	job, err := jr.toSingleOrPatternJob(r)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return job.Compile(r)
}

func (jr *JobDefinition) toSingleOrPatternJob(r *Retagger) (CompilableJob, error) {
	// The job can either be a SingleJob or a PatternJob.
	if jr.SourceSha != "" {
		if jr.SourcePattern != "" {
			// Configuration specified a SHA and a pattern -- use SHA to be safe, but warn about misconfiguration.
			r.logger.Log("level", "warn", "message", fmt.Sprintf("invalid configuration: Job %v specifies both a SHA (%v) and a Pattern (%v). Using SHA", jr.SourceImage, jr.SourceSha, jr.SourcePattern))
		}

		// This is a single job -- return it on its own.
		return SingleJobFromJobDefinition(jr, r), nil
	}

	// If no SHA is given, treat this as a pattern job.
	return PatternJobFromJobDefinition(jr, r), nil
}

// FromImages receives a list of Images and converts them into JobDefinitions.
func FromImages(images images.Images) ([]JobDefinition, error) {
	var jobs []JobDefinition

	for _, i := range images {
		js, err := FromImage(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		jobs = append(jobs, js...)
	}

	return jobs, nil
}

// FromImage takes an Image and converts it into a JobDefinition.
func FromImage(image images.Image) ([]JobDefinition, error) {
	var jobs []JobDefinition

	for _, t := range image.Tags {
		j, err := fromImageTagIncludeCustom(image, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		jobs = append(jobs, j...)
	}

	for _, p := range image.Patterns {
		j, err := fromImageTagPatternIncludeCustom(image, p)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		jobs = append(jobs, j...)
	}

	return jobs, nil
}

func fromImageTagIncludeCustom(image images.Image, tag images.Tag) ([]JobDefinition, error) {
	var jobs []JobDefinition

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

func fromImageTagPatternIncludeCustom(image images.Image, pattern images.TagPattern) ([]JobDefinition, error) {
	var jobs []JobDefinition

	j, err := fromImageTagPattern(image, pattern)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(pattern.CustomImages) == 0 {
		jobs = append(jobs, j)
	} else {
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
	}

	return jobs, nil
}

func fromImageTag(image images.Image, tag images.Tag) (JobDefinition, error) {
	j := JobDefinition{
		SourceImage: image.Name,
		SourceTag:   tag.Tag,
		SourceSha:   tag.Sha,
	}

	if image.OverrideRepoName != "" {
		j.Options.OverrideRepoName = image.OverrideRepoName
	}

	return j, nil
}

func fromImageTagPattern(image images.Image, tagPattern images.TagPattern) (JobDefinition, error) {
	j := JobDefinition{
		SourceImage:   image.Name,
		SourcePattern: tagPattern.Pattern,
	}

	if image.OverrideRepoName != "" {
		j.Options.OverrideRepoName = image.OverrideRepoName
	}

	return j, nil
}
