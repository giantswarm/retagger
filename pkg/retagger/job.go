package retagger

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/retagger/pkg/images"
)

// CompilableJob represents any Job which can be Compiled.
type CompilableJob interface {
	Compile(*Retagger) ([]SingleJob, error)
}

// Destination contains information about the target repository and tag of a job.
type Destination struct {
	Image string
	Tag   string
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

	if tagPattern.UpdateOnChange {
		j.Options.UpdateOnChange = true
	}

	return j, nil
}

// ExecutableJob represents any Job which can be Executed.
// type ExecutableJob interface {
// 	Execute(r *Retagger) error
// }

// JobCompiler contains a Job which can be Compiled.
// type JobCompiler struct {
// 	Job CompilableJob
// }

// Compile takes a CompilableJob and a Retagger and Compiles the job.
// func (jc *JobCompiler) Compile(job CompilableJob, r *Retagger) ([]SingleJob, error) {
// 	return job.Compile(r)
// }
