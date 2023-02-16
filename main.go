package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sync"

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
	// TagPattern is used to filter image tags. All tags matching the pattern
	// will be retagged.
	// Example: "v1.[234].*" or ".+-stable"
	TagPattern string `yaml:"tag_pattern"`
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
}

// skopeoTagList is used to unmarshal `skopeo list-tags` command output.
type skopeoTagList struct {
	Tags []string `yaml:"Tags"`
}

// BuildAndTag find all tags matching the img.TagPattern, retags, and pushes
// them to Quay and Aliyun container registries. Any optional parameters
// configured will be applied as well, e.g. tag suffix.
func (img *CustomImage) BuildAndTag() error {
	// Compile pattern first, so we can exit early if it fails.
	pattern, err := regexp.Compile(img.TagPattern)
	if err != nil {
		return fmt.Errorf("error compiling regexp pattern %q: %w", img.TagPattern, err)
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

tagLoop:
	// Iterate through all found tags and retag ones matching the pattern
	for _, tag := range tags {
		if !pattern.MatchString(tag) {
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

		// If no DockerfileExtras were defined, we can simply copy the upstream
		// image. We'll use skopeo for this, because it's awesome.
		if len(img.DockerfileExtras) == 0 {
			source := fmt.Sprintf("%s%s:%s", dockerTransport, img.Image, tag)
			wg := sync.WaitGroup{}
			wg.Add(2)
			// Quay
			destination := fmt.Sprintf("%s%s/%s:%s", dockerTransport, quayURL, destinationName, destinationTag)
			go copyImage(&wg, source, destination)
			// Aliyun
			destination = fmt.Sprintf("%s%s/%s:%s", dockerTransport, aliyunURL, destinationName, destinationTag)
			go copyImage(&wg, source, destination)
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
				continue tagLoop
			}
			logrus.Tracef(stdout.String())
		}

		// Start pushing to Quay immediately
		go pushImage(&wg, quayName)

		// Tag the image we've just built for Aliyun as well...
		{
			c, _, stderr := command("docker", "tag", quayName, aliyunName)
			if err := c.Run(); err != nil {
				logrus.Errorf("error tagging custom image for %s:%s: %v\n%s", img.Image, tag, err, stderr.String())
				continue tagLoop
			}
		}
		// and push to Aliyun
		go pushImage(&wg, aliyunName)

		wg.Wait()
	}

	return nil
}

// copyImage is a helper function used to invoke `skopeo copy`. Please note the
// `--all`, which makes skopeo include ALL SHAs included in the tag's digest,
// ensuring builds for all available platforms.
func copyImage(wg *sync.WaitGroup, source, destination string) {
	defer wg.Done()
	c, _, stderr := command("skopeo", "copy", "--all", source, destination)
	logrus.Debugf("copying %q to %q", source, destination)
	if err := c.Run(); err != nil {
		logrus.Errorf("error copying %q to %q: %v\n%s", source, destination, err, stderr.String())
	}
	logrus.Debugf("copied %q to %q", source, destination)
}

// pushImage is a helper function used to invoke `docker push`.
func pushImage(wg *sync.WaitGroup, nameAndTag string) {
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
	errors := 0
	for i, image := range customizedImages {
		logrus.Printf("[%d/%d] Retagging %s", i+1, len(customizedImages), image.Image)
		if err := image.BuildAndTag(); err != nil {
			logrus.Errorf("got error: %v", err)
			errors++
		}
	}

	if errors > 0 {
		logrus.Fatal("Retagging ended with %d errors", errors)
	}
	logrus.Infof("Done retagging %d images with no errors", len(customizedImages))
}
