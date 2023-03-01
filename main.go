package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	temporaryWorkingDir = path.Join(os.TempDir(), "retagger")

	flagLogLevel      string
	flagExecutorCount int
	flagExecutorID    int
)

const (
	customizedImagesFile = "images/customized-images.yaml"
	defaultPlatform      = "linux/amd64"
	dockerTransport      = "docker://"
	quayURL              = "quay.io/giantswarm"
	aliyunURL            = "giantswarm-registry.cn-shanghai.cr.aliyuncs.com/giantswarm"
)

// CustomImage represents a set of rules used to rebuild/retag multiple tags of
// the same image that match the specified tag pattern.
type CustomImage struct {
	// Image is the full name of the image to pull.
	// Example: "alpine", "docker.io/giantswarm/app-operator", or
	// "ghcr.io/fluxcd/kustomize-controller"
	Image string `yaml:"image"`
	// TagOrPattern is used to filter image tags. All tags matching the pattern
	// will be retagged. Required if SHA is specified.
	// Example: "v1.[234].*" or ".*-stable"
	TagOrPattern string `yaml:"tag_or_pattern,omitempty"`
	// SHA is used to filter image tags. If SHA is specified, it will take
	// precedence over TagOrPattern. However TagOrPattern is still required!
	// Example: 234cb88d3020898631af0ccbbcca9a66ae7306ecd30c9720690858c1b007d2a0
	SHA string `yaml:"sha,omitempty"`
	// Semver is used to filter image tags by semantic version constraints. All
	// tags satisfying the constraint will be retagged.
	Semver string `yaml:"semver,omitempty"`
	// Filter is a regexp pattern used to extract a part of the tag for Semver
	// comparison. First matched group will be supplied for semver comparison.
	// Example:
	//   Filter: "(.+)-alpine"  ->  Image tag: "3.12-alpine" -> Comparison: "3.12>=3.10"
	//   Semver: ">= 3.10"          Extracted group: "3.12"
	Filter string `yaml:"filter,omitempty"`
	// DockerfileExtras is a list of additional Dockerfile statements you want to
	// append to the upstream Dockerfile. (optional)
	// Example: ["RUN apk add -y bash"]
	DockerfileExtras []string `yaml:"dockerfile_extras,omitempty"`
	// AddTagSuffix is an extra string to append to the tag.
	// Example: "giantswarm", the tag would become "<tag>-giantswarm"
	AddTagSuffix string `yaml:"add_tag_suffix,omitempty"`
	// OverrideRepoName allows user to rewrite the name of the image entirely.
	// Example: "alpinegit", so "alpine" would become
	// "quay.io/giantswarm/alpinegit"
	OverrideRepoName string `yaml:"override_repo_name,omitempty"`
	// StripSemverPrefix removes the initial 'v' in 'v1.2.3' if enabled. Works
	// only when Semver is defined.
	StripSemverPrefix bool `yaml:"strip_semver_prefix,omitempty"`
}

func (img *CustomImage) Validate() error {
	if img.TagOrPattern == "" && img.SHA == "" && img.Semver == "" {
		return fmt.Errorf("neither %q, %q, nor %q specified", "tag_or_pattern", "semver", "sha")
	}
	if img.SHA != "" && img.TagOrPattern == "" {
		return fmt.Errorf("%q has to be specified when using %q", "tag_or_pattern", "sha")
	}
	if img.Semver != "" && (img.SHA != "" || img.TagOrPattern != "") {
		return fmt.Errorf("%q defined, %q and %q are redundant and will not be used", "semver", "tag_or_pattern", "sha")
	}
	if img.Filter != "" && img.Semver == "" {
		return fmt.Errorf("cannot use %q without a defined %q", "filter", "semver")
	}
	if img.Semver == "" && img.StripSemverPrefix {
		return fmt.Errorf("cannot strip semver prefix when %q is not defined", "semver")
	}
	return nil
}

// RetagUsingSHA pulls an image matching the SHA, retags, and pushes it to Quay and Aliyun.
// Any optional parameters configured will be applied as well, e.g. tag suffix.
// The pushed image will be tagged with the value of image.TagOrPattern.
func (img *CustomImage) RetagUsingSHA() error {
	// Overwrite image name if applicable
	destinationName := imageBaseName(img.Image)
	if img.OverrideRepoName != "" {
		destinationName = img.OverrideRepoName
	}
	// Add tag suffix if applicable
	destinationTag := img.TagOrPattern
	if img.AddTagSuffix != "" {
		destinationTag = img.TagOrPattern + "-" + img.AddTagSuffix
	}

	errorCounter := &atomic.Int64{}

	// If no DockerfileExtras were defined, we can simply copy the upstream
	// image. We'll use skopeo for this, because it's awesome.
	if len(img.DockerfileExtras) == 0 {
		source := fmt.Sprintf("%s%s@sha256:%s", dockerTransport, img.Image, img.SHA)
		wg := sync.WaitGroup{}
		wg.Add(2)
		// Quay
		destination := fmt.Sprintf("%s%s/%s:%s", dockerTransport, quayURL, destinationName, destinationTag)
		go copyImage(&wg, errorCounter, source, destination)
		// Aliyun
		destination = fmt.Sprintf("%s%s/%s:%s", dockerTransport, aliyunURL, destinationName, destinationTag)
		go copyImage(&wg, errorCounter, source, destination)
		wg.Wait()

		if errorCount := errorCounter.Load(); errorCount > 0 {
			return fmt.Errorf("finished %q with %d errors", img.Image, errorCount)
		}
		return nil
	}

	// DockerfileExtras were defined. We'll create a Dockefile which
	// references the upstream image and write our changes to it.
	var dockerfile string
	{
		tmp, err := os.CreateTemp(temporaryWorkingDir, "Dockerfile.*")
		if err != nil {
			return fmt.Errorf("error creating temporary Image: %w", err)
		}
		dockerfile = tmp.Name()
		defer func() {
			os.Remove(dockerfile)
		}()
		fmt.Fprintf(tmp, "FROM %s@sha256:%s\n", img.Image, img.SHA)
		for _, line := range img.DockerfileExtras {
			_, err := tmp.WriteString(line + "\n")
			if err != nil {
				return fmt.Errorf("error writing temporary Image: %w", err)
			}
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	name := fmt.Sprintf("%s:%s", destinationName, destinationTag)
	quayName := fmt.Sprintf("%s/%s", quayURL, name)
	aliyunName := fmt.Sprintf("%s/%s", aliyunURL, name)
	// Build the generated Dockerfile, tagging it for Quay
	{
		c, stdout, stderr := command("docker", "build", "-t", quayName, "-f", dockerfile, temporaryWorkingDir)
		if err := c.Run(); err != nil {
			return fmt.Errorf("error building custom image for \"%s@sha256:%s\": %w\n%s", img.Image, img.SHA, err, stderr.String())
		}
		logrus.Tracef(stdout.String())
	}

	// Start pushing to Quay immediately
	go pushImage(&wg, errorCounter, quayName)

	// Tag the image we've just built for Aliyun as well...
	{
		c, _, stderr := command("docker", "tag", quayName, aliyunName)
		if err := c.Run(); err != nil {
			return fmt.Errorf("error tagging custom image for \"%s@sha256:%s\": %w\n%s", img.Image, img.SHA, err, stderr.String())
		}
	}
	// and push to Aliyun
	go pushImage(&wg, errorCounter, aliyunName)

	wg.Wait()

	if errorCount := errorCounter.Load(); errorCount > 0 {
		return fmt.Errorf("finished %q with %d errors", img.Image, errorCount)
	}
	return nil
}

// RetagUsingTags finds all tags matching the img.TagOrPattern or
// img.Semver, retags, and pushes them to Quay and Aliyun container registries.
// Any optional parameters configured will be applied as well, e.g. tag suffix.
func (img *CustomImage) RetagUsingTags() error {
	// List available image tags
	var tags []string
	{
		c, stdout, stderr := command("skopeo", "list-tags", dockerTransport+img.Image)
		if err := c.Run(); err != nil {
			return fmt.Errorf("error listing tags for %q: %w\n%s", img.Image, err, stderr.String())
		}
		stl := skopeoTagList{
			Tags: []string{},
		}
		if err := yaml.Unmarshal(stdout.Bytes(), &stl); err != nil {
			return fmt.Errorf("error unmarshaling tags: %w", err)
		}
		tags = stl.Tags
	}

	// Filter the tags using TagOrPattern or Semver+Filter.
	tags, err := img.FilterTags(tags)
	if err != nil {
		return fmt.Errorf("error filtering tags: %w", err)
	}

	// Overwrite image name if applicable
	destinationName := imageBaseName(img.Image)
	if img.OverrideRepoName != "" {
		destinationName = img.OverrideRepoName
	}

	errorCounter := &atomic.Int64{}
tagLoop:
	// Iterate through all found tags and retag ones matching the semver/pattern
	for _, tag := range tags {
		// Add tag suffix if applicable
		destinationTag := tag
		if img.AddTagSuffix != "" {
			destinationTag = tag + "-" + img.AddTagSuffix
		}
		if img.Semver != "" && img.StripSemverPrefix {
			destinationTag = strings.TrimPrefix(destinationTag, "v")
		}

		// If no DockerfileExtras were defined, we can simply copy the upstream
		// image. We'll use skopeo for this, because it's awesome.
		if len(img.DockerfileExtras) == 0 {
			source := fmt.Sprintf("%s%s:%s", dockerTransport, img.Image, tag)
			wg := sync.WaitGroup{}
			wg.Add(2)
			// Quay
			destination := fmt.Sprintf("%s%s/%s:%s", dockerTransport, quayURL, destinationName, destinationTag)
			go copyImage(&wg, errorCounter, source, destination)
			// Aliyun
			destination = fmt.Sprintf("%s%s/%s:%s", dockerTransport, aliyunURL, destinationName, destinationTag)
			go copyImage(&wg, errorCounter, source, destination)
			wg.Wait()

			// We'll skip to the next tag
			continue
		}

		// DockerfileExtras were defined. We'll create a Dockefile which
		// references the upstream image and write our changes to it.
		var dockerfile string
		{
			tmp, err := os.CreateTemp(temporaryWorkingDir, "Dockerfile.*")
			if err != nil {
				logrus.Errorf("error creating temporary Image: %v", err)
				errorCounter.Add(1)
				continue tagLoop
			}
			dockerfile = tmp.Name()
			defer func() {
				os.Remove(dockerfile)
			}()
			fmt.Fprintf(tmp, "FROM --platform=%s %s:%s\n", defaultPlatform, img.Image, tag)
			for _, line := range img.DockerfileExtras {
				_, err := tmp.WriteString(line + "\n")
				if err != nil {
					logrus.Errorf("error writing temporary Image: %v", err)
					errorCounter.Add(1)
					continue tagLoop
				}
			}
		}

		wg := sync.WaitGroup{}
		wg.Add(2)
		name := fmt.Sprintf("%s:%s", destinationName, destinationTag)
		quayName := fmt.Sprintf("%s/%s", quayURL, name)
		aliyunName := fmt.Sprintf("%s/%s", aliyunURL, name)
		// Build the generated Dockerfile, tagging it for Quay
		{
			c, stdout, stderr := command("docker", "build", "-t", quayName, "-f", dockerfile, temporaryWorkingDir)
			if err := c.Run(); err != nil {
				logrus.Errorf("error building custom image for %s:%s: %v\n%s", img.Image, tag, err, stderr.String())
				errorCounter.Add(1)
				continue tagLoop
			}
			logrus.Tracef(stdout.String())
		}

		// Start pushing to Quay immediately
		go pushImage(&wg, errorCounter, quayName)

		// Tag the image we've just built for Aliyun as well...
		{
			c, _, stderr := command("docker", "tag", quayName, aliyunName)
			if err := c.Run(); err != nil {
				logrus.Errorf("error tagging custom image for %s:%s: %v\n%s", img.Image, tag, err, stderr.String())
				continue tagLoop
			}
		}
		// and push to Aliyun
		go pushImage(&wg, errorCounter, aliyunName)

		wg.Wait()
	}

	if errorCount := errorCounter.Load(); errorCount > 0 {
		return fmt.Errorf("finished %q with %d errors", img.Image, errorCount)
	}
	return nil
}

// FilterTags returns a trimmed down list of tags, based on defined rules. It
// uses either a TagOrPattern, or Semver+Filter fields, whichever are defined.
// Validate() function guarantees that only one method can be available at any
// given time.
func (img *CustomImage) FilterTags(tags []string) ([]string, error) {
	var filteredTags []string

	// Filter by TagOrPattern...
	if img.TagOrPattern != "" {
		pattern, err := regexp.Compile(img.TagOrPattern)
		if err != nil {
			return filteredTags, fmt.Errorf("error compiling regexp pattern %q: %w", img.TagOrPattern, err)
		}

		for _, tag := range tags {
			if pattern.MatchString(tag) {
				filteredTags = append(filteredTags, tag)
			}
		}

		return filteredTags, nil
	}

	// or by Semver (with Filter, if defined)
	constraint, err := semver.NewConstraint(img.Semver)
	if err != nil {
		return filteredTags, fmt.Errorf("error compiling semver constraint %q: %w", img.Semver, err)
	}
	var filter *regexp.Regexp
	if img.Filter != "" {
		f, err := regexp.Compile(img.Filter)
		if err != nil {
			return filteredTags, fmt.Errorf("error compiling semver filter %q: %w", img.Filter, err)
		}
		filter = f
	}

	for _, tag := range tags {
		semverToCompare := tag
		if filter != nil {
			matches := filter.FindAllStringSubmatch(tag, 1)
			// Tag does not match filter at all. This happens in repos/images
			// with multiple tagging formats.
			if len(matches) == 0 {
				continue
			}
			// Version subgroup not found. This may be concerning if the filter
			// is defined, but without any subgroups, hence the error message.
			if len(matches[0]) < 2 {
				logrus.Warnf("tag %q matched the pattern %q, but no groups were found", tag, filter.String())
				continue
			}
			// Always select the first subgroup
			semverToCompare = matches[0][1]
		}

		version, err := semver.NewVersion(semverToCompare)
		if err != nil {
			logrus.Debugf("image %q's tag (or its portion) %q is not a semantic version", img.Image, semverToCompare)
			continue
		}

		if constraint.Check(version) {
			filteredTags = append(filteredTags, tag)
		}
	}

	return filteredTags, nil
}

// skopeoTagList is used to unmarshal `skopeo list-tags` command output.
type skopeoTagList struct {
	Tags []string `yaml:"Tags"`
}

// imageBaseName is a helper function extracting base image name.
// Example: "registry.k8s.io/kube-apiserver" -> "kube-apiserver"
func imageBaseName(name string) string {
	if !strings.ContainsRune(name, '/') {
		return name
	}
	elems := strings.Split(name, "/")
	return elems[len(elems)-1]
}

// copyImage is a helper function used to invoke `skopeo copy`. Please note the
// `--all`, which makes skopeo include ALL SHAs included in the tag's digest,
// ensuring builds for all available platforms.
func copyImage(wg *sync.WaitGroup, errorCounter *atomic.Int64, source, destination string) {
	defer wg.Done()
	c, _, stderr := command("skopeo", "copy", "--all", source, destination)
	logrus.Debugf("copying %q to %q", source, destination)
	if err := c.Run(); err != nil {
		logrus.Errorf("error copying %q to %q: %v\n%s", source, destination, err, stderr.String())
	}
	logrus.Debugf("copied %q to %q", source, destination)
}

// pushImage is a helper function used to invoke `docker push`.
func pushImage(wg *sync.WaitGroup, errorCounter *atomic.Int64, nameAndTag string) {
	defer wg.Done()
	c, _, stderr := command("docker", "push", nameAndTag)
	logrus.Debugf("pushing %q", nameAndTag)
	if err := c.Run(); err != nil {
		logrus.Errorf("error pushing %q: %v\n%s", nameAndTag, err, stderr.String())
	}
	logrus.Debugf("pushed %q", nameAndTag)
}

// command is a helper function so I don't have to manually plug bytes.Buffer
// into command streams every time ;_;
func command(name string, args ...string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c := exec.Command(name, args...)
	c.Stdout = stdout
	c.Stderr = stderr
	return c, stdout, stderr
}

func init() {
	flag.StringVar(&flagLogLevel, "log-level", "debug", "Sets log level")
	flag.IntVar(&flagExecutorCount, "executor-count", 1, "Number of executors in a parallelized run")
	flag.IntVar(&flagExecutorID, "executor-id", 0, "ID of the executor in a parallelized run")
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	lvl, err := logrus.ParseLevel(flagLogLevel)
	if err != nil {
		logrus.Warnf("could not parse log level %q, defaulting to %q", flagLogLevel, logrus.DebugLevel.String())
	} else {
		logrus.SetLevel(lvl)
	}

	if flagExecutorID < 0 || flagExecutorID >= flagExecutorCount {
		logrus.Fatalf("%q flag has to be greater than 0 and lower than %q", "executor-id", "executor-count")
	}
	if flagExecutorCount < 1 {
		logrus.Fatal("%q cannot be lower than 1", "executor-count")
	}
	if flagExecutorCount > 10 {
		logrus.Warnf("%q is set to %d, are you sure that's on purpose?", "executor-count", flagExecutorCount)
	}
}

func main() {
	if err := os.MkdirAll(temporaryWorkingDir, 0777); err != nil {
		logrus.Fatal(err)
	}

	logger := logrus.WithField("executor", flagExecutorID)

	// Load custom dockerfile definitions from a file
	customizedImages := []CustomImage{}
	{
		b, err := os.ReadFile(customizedImagesFile)
		if err != nil {
			logger.Fatalf("error reading %q: %s", customizedImagesFile, err)
		}
		if err := yaml.Unmarshal(b, &customizedImages); err != nil {
			logger.Fatalf("error unmarshaling %q: %s", customizedImagesFile, err)
		}
	}

	logger.Infof("Found %d custom Images to build", len(customizedImages))

	// Iterate over every image x tag and retag/rebuild it
	errorCounter := 0
	for i, image := range customizedImages {
		// Skip images meant for other executors
		if i%flagExecutorCount != flagExecutorID {
			continue
		}
		if err := image.Validate(); err != nil {
			logger.Errorf("[%d/%d] %q error: %s", i+1, len(customizedImages), image.Image, err)
			errorCounter++
			continue
		}
		logger.Printf("[%d/%d] Retagging %q", i+1, len(customizedImages), image.Image)
		if image.SHA != "" {
			if err := image.RetagUsingSHA(); err != nil {
				logger.Errorf("got error: %v", err)
				errorCounter++
			}
		} else {
			if err := image.RetagUsingTags(); err != nil {
				logger.Errorf("got error: %v", err)
				errorCounter++
			}
		}
	}

	if errorCounter > 0 {
		logger.Fatalf("Retagging ended with %d errors", errorCounter)
	}
	logger.Infof("Done retagging %d images with no errors", len(customizedImages))
}
