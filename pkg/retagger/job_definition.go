package retagger

import (
	"fmt"

	"github.com/giantswarm/microerror"
)

// JobDefinition represents a single or pattern job which has been read from the input file but is yet to be compiled.
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
