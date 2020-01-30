package retagger

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/giantswarm/microerror"
	dockerRegistry "github.com/nokia/docker-registry-client/registry"
)

// PatternJob contains a definition for generating multiple single jobs based on a pattern.
type PatternJob struct {
	SourcePattern string
	Source        Source

	Destination Destination

	Options JobOptions
}

// Compile expands a PatternJob into one or multiple SingleJobs using the given Retagger instance.
func (job *PatternJob) Compile(r *Retagger) ([]SingleJob, error) {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("compiling jobs for image %v using pattern %v, with options %#v", job.Source.Image, job.SourcePattern, job.Options))

	// Create a reference to the external registry.
	o := dockerRegistry.Options{
		Logf:          dockerRegistry.Quiet,
		DoInitialPing: false,
	}
	externalRegistry, err := dockerRegistry.NewCustom(fmt.Sprintf("https://%s", job.Source.RepoPath), o)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Find tags which match the pattern.
	matches, err := getExternalTagMatches(externalRegistry, job.Source.FullImageName, job.SourcePattern)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(matches) == 0 {
		r.logger.Log("level", "warn", "message", fmt.Sprintf("No upstream image tags were found matching the pattern %s", job.SourcePattern))
	} else {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("Found %d upstream tags which match the pattern %s", len(matches), job.SourcePattern))
	}

	// Get SHA/Tag pairs from our quay registry.
	existingTagMap, err := r.getTagDetails(job.Destination.Image)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var jobs []SingleJob

	for _, match := range matches {
		_, exists := existingTagMap[match]
		if !exists {
			// Tag is new - get SHA and tag it.
			newDigest, err := externalRegistry.ManifestV2Digest(job.Source.FullImageName, match)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			sourceSHA := strings.TrimPrefix(newDigest.String(), "sha256:")
			// Create job with new SHA.
			j := SingleJob{

				Source: job.Source,

				Options: job.Options,
			}
			// Override Source options from our pattern.
			j.Source.Tag = match
			j.Source.SHA = sourceSHA
			j.Destination = GetDestinationForJob(&j, r)
			jobs = append(jobs, j)
		}
	}

	r.logger.Log("level", "debug", "message", fmt.Sprintf("Compiled %d jobs to process", len(jobs)))

	return jobs, nil
}

func (job *PatternJob) GetOptions() JobOptions {
	return job.Options
}

func (job *PatternJob) GetSource() Source {
	return job.Source
}

// PatternJobFromJobDefinition converts a JobDefinition into a PatternJob.
func PatternJobFromJobDefinition(jobDef *JobDefinition, r *Retagger) *PatternJob {
	job := &PatternJob{
		SourcePattern: jobDef.SourcePattern,
		Source:        GetSourceForJob(jobDef, r),

		Options: jobDef.Options,
	}
	job.Destination = GetDestinationForJob(job, r)
	return job
}

// getExternalTagMatches searches the given docker registry for tags matching the given pattern.
func getExternalTagMatches(r *dockerRegistry.Registry, image string, pattern string) ([]string, error) {
	// Make sure our constraint is valid.
	c, err := semver.NewConstraint(pattern)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Get the tags for this image from the external registry.
	externalRegistryTags, err := r.Tags(image)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Find tags matching our configured pattern.
	var matches []string
	for _, t := range externalRegistryTags {
		v, err := semver.NewVersion(t)
		if err != nil { // We do not care if the version is not semver.
			continue
		}
		m, _ := c.Validate(v) // We do not care why the validation might fail.
		if m {
			matches = append(matches, t)
		}
	}

	return matches, nil
}
