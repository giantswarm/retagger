package retagger

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/retagger/pkg/images"
)

type Job struct {
	SourceImage   string
	SourceTag     string
	SourceSha     string
	SourcePattern string

	Options JobOptions
}

type JobOptions struct {
	// DockerfileOptions - list of strings we add for Dockerfile to build custom image.
	DockerfileOptions []string

	TagSuffix string

	OverrideRepoName string

	// UpdateOnChange sets whether a pattern Job should update the destination image if a source image changes for a given tag
	UpdateOnChange bool
}

func FromImages(images images.Images) ([]Job, error) {
	var jobs []Job

	for _, i := range images {
		js, err := FromImage(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		jobs = append(jobs, js...)
	}

	return jobs, nil
}

func FromImage(image images.Image) ([]Job, error) {
	var jobs []Job

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

func fromImageTagIncludeCustom(image images.Image, tag images.Tag) ([]Job, error) {
	var jobs []Job

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

func fromImageTagPatternIncludeCustom(image images.Image, pattern images.TagPattern) ([]Job, error) {
	var jobs []Job

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

func fromImageTag(image images.Image, tag images.Tag) (Job, error) {
	j := Job{
		SourceImage: image.Name,
		SourceTag:   tag.Tag,
		SourceSha:   tag.Sha,
	}

	if image.OverrideRepoName != "" {
		j.Options.OverrideRepoName = image.OverrideRepoName
	}

	return j, nil
}

func fromImageTagPattern(image images.Image, tagPattern images.TagPattern) (Job, error) {
	j := Job{
		SourceImage:   image.Name,
		SourceTag:     tagPattern.Tag,
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
