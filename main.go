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
	"gopkg.in/yaml.v3"
)

var (
	temporaryWorkingDir = path.Join(os.TempDir(), "retagger")
)

const (
	defaultPlatform = "linux/amd64"
	dockerTransport = "docker://"
	quayURL         = "quay.io/giantswarm"
	aliyunURL       = "giantswarm-registry.cn-shanghai.cr.aliyuncs.com/giantswarm"
)

// CustomImage represents a set of rules used to rebuild/retag multiple tags of
// the same image that match the specified tag pattern.
type CustomImage struct {
	// Image is the full name of the image to pull.
	// Example: "alpine", "docker.io/giantswarm/app-operator", or
	// "ghcr.io/fluxcd/kustomize-controller"
	Image string `yaml:"image"`
	// TagOrPattern is used to filter image tags. All tags matching the pattern
	// will be retagged. If Semver is defined, TagOrPattern will act as a
	// filter.
	// Example: "v1.[234].*" or ".+-stable"
	TagOrPattern string `yaml:"tag_or_pattern,omitempty"`
	// SHA is used to filter image tags. If SHA is specified, it will take
	// precedence over TagOrPattern. However TagOrPattern is still required!
	// Example: 234cb88d3020898631af0ccbbcca9a66ae7306ecd30c9720690858c1b007d2a0
	SHA string `yaml:"sha,omitempty"`
	// Semver is used to filter image tags by semantic version constraints. All
	// tags satisfying the constraint will be retagged.
	Semver string `yaml:"semver,omitempty"`
	// DockerfileExtras is a list of additional Dockerfile statements you want to
	// append to the upstream Dockerfile. (optional)
	// Example: ["RUN apk add -y bash"]
	DockerfileExtras []string `yaml:"dockerfile_extras,omitempty"`
	// AddTagSuffix is an extra string to append to the tag. (optional)
	// Example: "giantswarm", the tag would become "<tag>-giantswarm"
	AddTagSuffix string `yaml:"add_tag_suffix,omitempty"`
	// OverrideRepoName allows user to rewrite the name of the image entirely.
	// (optional)
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
	if img.Semver != "" && img.SHA != "" {
		return fmt.Errorf("%q acts as a filter for %q, %q is redundant", "tag_or_pattern", "semver", "sha")
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
	destinationName := img.Image
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
	// Compile pattern first, so we can exit early if it fails.
	var pattern *regexp.Regexp
	if img.TagOrPattern != "" {
		compiled, err := regexp.Compile(img.TagOrPattern)
		if err != nil {
			return fmt.Errorf("error compiling regexp pattern %q: %w", img.TagOrPattern, err)
		}
		pattern = compiled
	}
	// Compile semver constraint
	var constraint *semver.Constraints
	if img.Semver != "" {
		compiled, err := semver.NewConstraint(img.Semver)
		if err != nil {
			return fmt.Errorf("error compiling semver constraint %q: %w", img.Semver, err)
		}
		constraint = compiled
	}

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

	errorCounter := &atomic.Int64{}
tagLoop:
	// Iterate through all found tags and retag ones matching the semver/pattern
	for _, tag := range tags {
		if constraint != nil {
			version, err := semver.NewVersion(tag)
			if err != nil {
				logrus.Debugf("image %q's tag %q is not a semantic version", img.Image, tag)
				continue
			}
			if !constraint.Check(version) {
				continue
			}
		}
		// pattern can be either a sole filter, or used in tandem with semver
		if pattern != nil && !pattern.MatchString(tag) {
			continue
		}

		// Overwrite image name if applicable
		destinationName := img.Image
		if img.OverrideRepoName != "" {
			destinationName = img.OverrideRepoName
		}
		// Add tag suffix if applicable
		destinationTag := tag
		if img.AddTagSuffix != "" {
			destinationTag = tag + "-" + img.AddTagSuffix
		}
		if constraint != nil && img.StripSemverPrefix {
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

// skopeoTagList is used to unmarshal `skopeo list-tags` command output.
type skopeoTagList struct {
	Tags []string `yaml:"Tags"`
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
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	if err := os.MkdirAll(temporaryWorkingDir, 0777); err != nil {
		logrus.Fatal(err)
	}

	// Load custom dockerfile definitions from a file
	customizedImages := []CustomImage{}
	{
		b, err := os.ReadFile("customized-images.yaml")
		if err != nil {
			logrus.Fatalf("error reading customized-images.yaml: %s", err)
		}
		if err := yaml.Unmarshal(b, &customizedImages); err != nil {
			logrus.Fatalf("error unmarshaling customized-images.yaml: %s", err)
		}
	}

	logrus.Infof("Found %d custom Images to build", len(customizedImages))

	// Iterate over every image x tag and retag/rebuild it
	errorCounter := 0
	for i, image := range customizedImages {
		if err := image.Validate(); err != nil {
			logrus.Errorf("[%d/%d] %q error: %s", i+1, len(customizedImages), image.Image, err)
			errorCounter++
			continue
		}
		logrus.Printf("[%d/%d] Retagging %q", i+1, len(customizedImages), image.Image)
		if image.SHA != "" {
			if err := image.RetagUsingSHA(); err != nil {
				logrus.Errorf("got error: %v", err)
				errorCounter++
			}
		} else {
			if err := image.RetagUsingTags(); err != nil {
				logrus.Errorf("got error: %v", err)
				errorCounter++
			}
		}
	}

	if errorCounter > 0 {
		logrus.Fatalf("Retagging ended with %d errors", errorCounter)
	}
	logrus.Infof("Done retagging %d images with no errors", len(customizedImages))
}
