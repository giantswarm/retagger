package retagger

import (
	"fmt"
	"regexp"

	dockerRegistry "github.com/nokia/docker-registry-client/registry"

	"github.com/giantswarm/microerror"
)

// ExecuteJobs runs the jobs associated with this Retagger
func (r *Retagger) ExecuteJobs() error {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("start executing %d jobs", len(r.jobs)))

	if r.whatif {
		r.logger.Log("level", "info", "message", "Retagger is in --whatif mode. Listing jobs, but not running.")
	}

	for _, j := range r.jobs {
		err := r.executeJob(j)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	r.logger.Log("level", "debug", "message", fmt.Sprintf("successfully finished executing %d jobs", len(r.jobs)))

	return nil
}

// executeJob accepts a job definition and runs it as either a pattern job or a single job
func (r *Retagger) executeJob(job Job) error {
	if job.SourceSha != "" {
		if job.SourcePattern != "" {
			// Configuration specified a SHA and a pattern -- use SHA to be safe, but warn about misconfiguration
			r.logger.Log("level", "warn", "message", fmt.Sprintf("invalid configuration: Job %v specifies both a SHA (%v) and a Pattern (%v). Using SHA", job.SourceImage, job.SourceSha, job.SourcePattern))
		}
		return r.executeSingleJob(job, true) // Skip existing tags for this job
	}
	return r.executePatternJob(job)
}

// executePatternJob accepts a job definition containing a tag pattern
// and creates and runs single jobs for each tag matching the pattern.
func (r *Retagger) executePatternJob(job Job) error {
	jobs, err := r.compilePatternJobs(job)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		r.executeSingleJob(job, false) // Do not skip existing tags for this job
	}

	return nil
}

func (r *Retagger) compilePatternJobs(job Job) ([]Job, error) {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("compiling jobs for image %v using pattern %v, with options %#v", job.SourceImage, job.SourcePattern, job.Options))

	// Populate tags to check
	//  - get tag list from source repo
	//  - run list against pattern
	//  - take matches from ^ and extract those not in quay as A, those which are as B
	// Check for SHAs
	//  - Get upstream SHAs for items in B IF their update option is set
	//  - Add items with new SHAs to A
	// Run single jobs for each SHA in A

	// Make sure our pattern is valid
	pattern, err := regexp.Compile(job.SourcePattern)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Get SHA/Tag pairs from our quay registry
	quayTagMap, err := r.registry.GetQuayTagMap(job.SourceImage)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Handle remote, Docker Hub, and Docker library image path formats
	registryPath, err := r.registry.GuessRegistryPath(job.SourceImage)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Create a reference to the external registry
	o := dockerRegistry.Options{
		Logf:          dockerRegistry.Quiet,
		DoInitialPing: false,
	}
	externalRegistry, err := dockerRegistry.NewCustom(fmt.Sprintf("https://%s", registryPath.Hostname()), o)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	fullImageName, err := r.registry.GetRepositoryFromPath(registryPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Get the tags for this image from the external registry
	externalRegistryTags, err := externalRegistry.Tags(fullImageName)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Find tags matching our configured pattern
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

	jobs := []Job{}
	// Find tags which need to be re-checked and updated
	for _, match := range matches {
		sourceSHA := ""

		_, exists := quayTagMap[match]

		if !exists {
			// Tag is new - get SHA and tag it
			newDigest, err := externalRegistry.ManifestDigest(fullImageName, match)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			sourceSHA = newDigest.String()

		} else {
			tag := quayTagMap[match]
			if job.Options.UpdateOnChange {
				// Tag exists, but we should update the image

				newDigest, err := externalRegistry.ManifestDigest(fullImageName, tag.Name)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				if newDigest.String() != tag.ManifestDigest {
					// Retag this image with this tag
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
			// Create job with new SHA
			j := Job{
				SourceTag:   match,
				SourceImage: job.SourceImage,
				SourceSha:   sourceSHA,
				Options:     job.Options,
			}
			jobs = append(jobs, j)
		}
	}

	r.logger.Log("level", "debug", "message", fmt.Sprintf("Compiled %d jobs to process", len(jobs)))

	return jobs, nil
}

// executeSingleJob runs one job definition, optionally skipping jobs with tags which already exist
func (r *Retagger) executeSingleJob(job Job, skipExisting bool) error {
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

	var err error
	if skipExisting {
		exists, err := r.registry.CheckImageTagExists(destinationImage, destinationTag)
		if err != nil {
			return microerror.Mask(err)
		}
		if exists {
			r.logger.Log("level", "debug", "message", fmt.Sprintf("image %s:%s already exists, skipping it now", destinationImage, destinationTag))
			return nil
		}
	}

	if r.whatif {
		r.logger.Log("level", "info", "message", fmt.Sprintf("WHATIF: %s:%s will be tagged as %s:%s with digest %s",
			job.SourceImage, job.SourceTag, destinationImage, destinationTag, job.SourceSha))
	} else {
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
	}

	return nil
}
