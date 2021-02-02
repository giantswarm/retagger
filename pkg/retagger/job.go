package retagger

// CompilableJob represents any Job which can be Compiled.
type CompilableJob interface {
	Compile(*Retagger) ([]SingleJob, error)
	GetOptions() JobOptions
	GetSource() Source
}

// Destination contains information about the target repository of a job.
type Destination struct {
	Image string
	Tag   string
}

// JobOptions specifies optional features for modifying the behavior of the job during tagging.
type JobOptions struct {
	// DockerfileOptions - list of strings we add for Dockerfile to build custom image.
	DockerfileOptions []string

	TagSuffix string

	TagTrimVersionPrefix bool

	OverrideRepoName string
}

// Source contains information about the source (upstream) of a job.
type Source struct {
	Image         string
	Tag           string
	SHA           string
	RepoPath      string
	FullImageName string
}
