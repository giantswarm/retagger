package retagger

import (
	"fmt"
	"regexp"

	"github.com/giantswarm/microerror"
	dockerRegistry "github.com/nokia/docker-registry-client/registry"
)

// PatternJob contains a definition for generating multiple single jobs based on a pattern.
type PatternJob struct {
	SourceImage   string
	SourcePattern string

	Destination Destination

	Options JobOptions
}

// PatternJobFromJobDefinition converts a JobDefinition into a PatternJob
func PatternJobFromJobDefinition(j *JobDefinition, r *Retagger) *PatternJob {
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
	quayTagMap, err := r.GetTagDetails(job.SourceImage)
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
