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

	r.logger.Log("level", "info", "message", "ci-cleaner tags:")
	hwTags, _ := r.registry.ListImageTags("ci-cleaner")
	r.logger.Log("level", "info", "message", fmt.Sprintf("found %d tags", len(hwTags)))
	count := 0
	for _, t := range hwTags {
		if count < 100 {
			r.logger.Log("level", "info", "message", t)
			// r.registry.GetDigest("ci-cleaner", t)
			count = count + 1
		}
	}

	r.logger.Log("level", "debug", "message", fmt.Sprintf("successfully finished executing %d jobs", len(r.jobs)))

	return nil
}

func (r *Retagger) executeJob(job Job) error {
	if job.SourceSha != "" {
		if job.SourcePattern != "" {
			// Configuration specified a SHA and a pattern -- use SHA to be safe, but warn about misconfiguration
			r.logger.Log("level", "warn", "message", fmt.Sprintf("invalid configuration: Job %v specifies both a SHA (%v) and a Pattern (%v). Using SHA", job.SourceImage, job.SourceSha, job.SourcePattern))
		}
		return r.executeSingleJob(job)
	}
	return r.executePatternJob(job)
}

func (r *Retagger) executePatternJob(job Job) error {
	r.logger.Log("level", "info", "message", fmt.Sprintf("running pattern job: %v, %v with options %#v", job.SourceImage, job.SourcePattern, job.Options))
	// Todo
	// Get SHA/Tag pairs from our quay
	// taggedImageTags := "" // in docker reg
	// Populate tags to check
	//  - get tag list from source repo
	//  - run list against pattern
	//  - take matches from ^ and extract those not in quay as A, those which are as B
	// Check for SHAs
	//  - Get upstream SHAs for items in B IF their update option is set
	//  - Add items with new SHAs to A
	// Run single jobs for each SHA in A
	return nil
}

func (r *Retagger) executeSingleJob(job Job) error {
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
