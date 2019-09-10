package retagger

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/retagger/pkg/config"
)

type Job struct {
	SourceImage string
	SourceTag   string
	SourceSha   string

	Options JobOptions
}

type JobOptions struct {
	// DockerfileOptions - list of strings we add for Dockerfile to build custom image.
	DockerfileOptions []string

	TagSuffix string

	OverrideRepoName string
}

func FromConfig(conf config.Config) ([]Job, error) {
	var jobs []Job

	for _, i := range conf.Images {
		js, err := FromImage(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		jobs = append(jobs, js...)
	}

	return jobs, nil
}

func FromImage(image config.Image) ([]Job, error) {
	var jobs []Job

	for _, t := range image.Tags {
		j, err := fromImageTagIncludeCustom(image, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		jobs = append(jobs, j...)
	}

	return jobs, nil
}

func fromImageTagIncludeCustom(image config.Image, tag config.Tag) ([]Job, error) {
	var jobs []Job

	j, err := fromImageTag(image, tag)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	jobs = append(jobs, j)

	for _, c := range tag.CustomImages {
		j, err = fromImageTagCustomImage(image, tag, c)
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

func fromImageTag(image config.Image, tag config.Tag) (Job, error) {
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
