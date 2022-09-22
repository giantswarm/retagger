package retagger

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	dockerRegistry "github.com/nokia/docker-registry-client/registry"
)

// PatternJob contains a definition for generating multiple single jobs based on a pattern.
type PatternJob struct {
	logger micrologger.Logger

	SourceFilter  string
	SourcePattern string
	Source        Source

	Destination Destination

	Options JobOptions
}

// Compile expands a PatternJob into one or multiple SingleJobs using the given Retagger instance.
func (job *PatternJob) Compile(r *Retagger) ([]SingleJob, error) {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("compiling jobs for image %v / %v using pattern %v, with filter %s, with options %#v", job.Source.RepoPath, job.Source.Image, job.SourcePattern, job.SourceFilter, job.Options))

	// Create a reference to the external registry.
	url := fmt.Sprintf("https://%s", job.Source.RepoPath)
	var transport http.RoundTripper = http.DefaultTransport
	transport = wrapTransport(transport, url, job.logger)
	externalRegistry := &dockerRegistry.Registry{
		Client: &http.Client{Transport: transport},
		URL:    url,
		Logf:   dockerRegistry.Quiet, // Ignore logs
	}

	// Find tags which match the pattern.
	matches, err := job.getExternalTagMatches(externalRegistry, job.Source.FullImageName, job.SourcePattern, job.SourceFilter)
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
		j := SingleJob{
			Source:  job.Source,
			Options: job.Options,
		}

		// Override Source options from our pattern.
		j.Source.Tag = match
		j.Destination = GetDestinationForJob(&j, r)

		_, exists := existingTagMap[j.Destination.Tag]
		if !exists {
			// Tag is new - get SHA and tag it.
			newDigest, err := externalRegistry.ManifestV2Digest(job.Source.FullImageName, match)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			j.Source.SHA = strings.TrimPrefix(newDigest.String(), "sha256:")
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

// getExternalTagMatches searches the given docker registry for tags matching the given pattern.
func (job *PatternJob) getExternalTagMatches(r *dockerRegistry.Registry, image string, pattern string, filter string) ([]string, error) {
	// Make sure our constraint is valid.
	c, err := semver.NewConstraint(pattern)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Get the tags for this image from the external registry.
	var externalRegistryTags []string
	{
		o := func() error {
			externalRegistryTags, err = r.Tags(image)
			if err != nil && IsTrailerEOF(err) {
				return microerror.Mask(err)
			} else if err != nil {
				return backoff.Permanent(err)
			}
			return nil
		}
		b := backoff.NewExponential(time.Minute, 10*time.Second)
		n := backoff.NewNotifier(job.logger, context.Background())
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	if filter == "" {
		filter = "(?P<version>.*)"
	}
	m, err := regexp.Compile(filter)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	matches := job.findTags(externalRegistryTags, c, m)
	return matches, nil
}

func (job *PatternJob) findTags(tags []string, c *semver.Constraints, m *regexp.Regexp) (matches []string) {
	// Find tags matching our configured pattern.
	for _, t := range tags {
		//job.logger.Log("level", "debug", "message", fmt.Sprintf("Checking external tag: %s ", t))

		ts := filterAndExtract(t, m)
		if ts == "" {
			//job.logger.Log("level", "debug", "message", fmt.Sprintf("Image %s:%s does not match the filter %s", job.Source.Image, t, job.SourceFilter))
			continue
		}

		v, err := semver.NewVersion(ts)
		if err != nil { // We do not care if the version is not semver.
			continue
		}

		m, _ := c.Validate(v)
		//for _, e := range errs {
		//	job.logger.Log("level", "debug", "message", fmt.Sprintf("Image %s:%s does not fulfill constraint %s because %s", job.Source.Image, t, job.SourcePattern, e.Error()))
		//}

		if m {
			matches = append(matches, t)
		}
	}

	return matches
}

func filterAndExtract(t string, m *regexp.Regexp) string {
	if submatches := m.FindStringSubmatchIndex(t); len(submatches) > 0 {
		result := []byte{}
		result = m.ExpandString(result, "$version", t, submatches)
		ts := string(result)
		return ts
	}
	return ""
}

// PatternJobFromJobDefinition converts a JobDefinition into a PatternJob.
func PatternJobFromJobDefinition(jobDef *JobDefinition, r *Retagger) *PatternJob {
	job := &PatternJob{
		logger: r.logger,

		SourcePattern: jobDef.SourcePattern,
		SourceFilter:  jobDef.SourceFilter,
		Source:        GetSourceForJob(jobDef, r),

		Options: jobDef.Options,
	}
	job.Destination = GetDestinationForJob(job, r)
	return job
}
