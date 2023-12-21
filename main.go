package main

import (
	"bytes"
	"fmt"
	"net/url"
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
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

var (
	skopeoSyncOutputPattern = regexp.MustCompile(`Would have copied image.*?from="docker://(.*?)[@:](.*?)".*`)
	temporaryWorkingDir     = path.Join(os.TempDir(), "retagger")

	flagLogLevel         string
	flagExecutorCount    int
	flagExecutorID       int
	flagSkipExistingTags bool
)

const (
	customizedImagesFile = "images/customized-images.yaml"
	defaultPlatform      = "linux/amd64"
	dockerTransport      = "docker://"
	filteredFileSuffix   = ".filtered"

	aliyunURL        = "giantswarm-registry.cn-shanghai.cr.aliyuncs.com/giantswarm"
	azureURL         = "gsoci.azurecr.io/giantswarm"
	azureUpstreamURL = "gsoci.azurecr.io/upstream"
	quayURL          = "quay.io/giantswarm"
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

// RetagUsingSHA pulls an image matching the SHA, retags, and pushes it to
// Quay, AzureCR and Aliyun.
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
	upstreamName := removeRegistryPrefix(img.Image)

	errorCounter := &atomic.Int64{}

	// If no DockerfileExtras were defined, we can simply copy the upstream
	// image. We'll use skopeo for this, because it's awesome.
	if len(img.DockerfileExtras) == 0 {
		source := fmt.Sprintf("%s%s@sha256:%s", dockerTransport, img.Image, img.SHA)
		wg := sync.WaitGroup{}
		wg.Add(4)
		// Quay
		destination := fmt.Sprintf("%s%s/%s:%s", dockerTransport, quayURL, destinationName, destinationTag)
		go copyImage(&wg, errorCounter, source, destination)
		// AzureCR/giantswarm
		destination = fmt.Sprintf("%s%s/%s:%s", dockerTransport, azureURL, destinationName, destinationTag)
		go copyImage(&wg, errorCounter, source, destination)
		// AzureCR/upstream
		destination = fmt.Sprintf("%s%s/%s:@sha256:%s", dockerTransport, azureUpstreamURL, upstreamName, img.SHA)
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
	wg.Add(4)
	name := fmt.Sprintf("%s:%s", destinationName, destinationTag)
	quayName := fmt.Sprintf("%s/%s", quayURL, name)
	azureName := fmt.Sprintf("%s/%s", azureURL, name)
	azureUpstreamName := fmt.Sprintf("%s/%s@sha256:%s", azureUpstreamURL, upstreamName, img.SHA)
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

	// Tag the image we've just built for AzureCR...
	{
		c, _, stderr := command("docker", "tag", quayName, azureName)
		if err := c.Run(); err != nil {
			return fmt.Errorf("error tagging custom image for \"%s@sha256:%s\": %w\n%s", img.Image, img.SHA, err, stderr.String())
		}
	}
	// push to AzureCR...
	go pushImage(&wg, errorCounter, azureName)

	// Tag the image we've just built for AzureCR upstream...
	{
		c, _, stderr := command("docker", "tag", quayName, azureUpstreamName)
		if err := c.Run(); err != nil {
			return fmt.Errorf("error tagging custom image for \"%s@sha256:%s\": %w\n%s", img.Image, img.SHA, err, stderr.String())
		}
	}
	// push to AzureCR upstream...
	go pushImage(&wg, errorCounter, azureUpstreamName)

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
	tags, err := listTags(img.Image)
	if err != nil {
		return err
	}

	// Overwrite image name if applicable
	destinationName := imageBaseName(img.Image)
	if img.OverrideRepoName != "" {
		destinationName = img.OverrideRepoName
	}
	upstreamName := removeRegistryPrefix(img.Image)

	// Filter the tags using TagOrPattern or Semver+Filter.
	tags, err = img.FilterTags(tags)
	if err != nil {
		return fmt.Errorf("error filtering tags: %w", err)
	}

	// Exclude tags existing in all registries
	if flagSkipExistingTags {
		quayTags, err := listTags(fmt.Sprintf("%s/%s", quayURL, destinationName))
		if err != nil {
			logrus.Warnf("error getting Quay.io tags: %w", err)
		}
		azureTags, err := listTags(fmt.Sprintf("%s/%s", azureURL, destinationName))
		if err != nil {
			logrus.Warnf("error getting AzureCR tags: %w", err)
		}
		aliyunTags, err := listTags(fmt.Sprintf("%s/%s", aliyunURL, destinationName))
		if err != nil {
			logrus.Warnf("error getting Aliyun tags: %w", err)
		}
		tags = img.FindMissingTags(tags, quayTags, azureTags, aliyunTags)
		logrus.Infof("Found %d missing tags for image %q", len(tags), img.Image)
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
			wg.Add(4)
			// Quay
			destination := fmt.Sprintf("%s%s/%s:%s", dockerTransport, quayURL, destinationName, destinationTag)
			go copyImage(&wg, errorCounter, source, destination)
			// Azure
			destination = fmt.Sprintf("%s%s/%s:%s", dockerTransport, azureURL, destinationName, destinationTag)
			go copyImage(&wg, errorCounter, source, destination)
			// Azure upstream
			destination = fmt.Sprintf("%s%s/%s:%s", dockerTransport, azureUpstreamURL, upstreamName, tag)
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
		wg.Add(4)
		name := fmt.Sprintf("%s:%s", destinationName, destinationTag)
		quayName := fmt.Sprintf("%s/%s", quayURL, name)
		azureName := fmt.Sprintf("%s/%s", azureURL, name)
		azureUpstreamName := fmt.Sprintf("%s/%s:%s", azureUpstreamURL, upstreamName, tag)
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

		// Tag the image we've just built for AzureCR...
		{
			c, _, stderr := command("docker", "tag", quayName, azureName)
			if err := c.Run(); err != nil {
				logrus.Errorf("error tagging custom image for %s:%s: %v\n%s", img.Image, tag, err, stderr.String())
				continue tagLoop
			}
		}
		// and push to Azure
		go pushImage(&wg, errorCounter, azureName)

		// Tag the image we've just built for AzureCR upstream...
		{
			c, _, stderr := command("docker", "tag", quayName, azureUpstreamName)
			if err := c.Run(); err != nil {
				logrus.Errorf("error tagging custom image for %s:%s: %v\n%s", img.Image, tag, err, stderr.String())
				continue tagLoop
			}
		}
		// and push to AzureCR upstream
		go pushImage(&wg, errorCounter, azureUpstreamName)

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
			logrus.Tracef("image %q's tag (or its portion) %q is not a semantic version", img.Image, semverToCompare)
			continue
		}

		if constraint.Check(version) {
			filteredTags = append(filteredTags, tag)
		}
	}

	return filteredTags, nil
}

// findMissingTags returns a list of items of the 'tags' slice that are missing
// from at least one of the 'present' slices.
func (img *CustomImage) FindMissingTags(tags []string, present ...[]string) []string {
	var filteredTags []string
	for _, tag := range tags {
		tagIsMissing := false

		destinationTag := tag
		if img.AddTagSuffix != "" {
			destinationTag = tag + "-" + img.AddTagSuffix
		}
		if img.Semver != "" && img.StripSemverPrefix {
			destinationTag = strings.TrimPrefix(destinationTag, "v")
		}

		for _, existingTags := range present {
			if !slices.Contains(existingTags, destinationTag) {
				tagIsMissing = true
				break
			}
		}

		// We always want to attempt to sync these mutable tags
		if tag == "latest" || tag == "develop" || tag == "debug" {
			logrus.Tracef("image %s has a mutable tag (%s) so considering it missing", img.Image, tag)
			tagIsMissing = true
		}

		if tagIsMissing {
			filteredTags = append(filteredTags, tag)
		}
	}
	return filteredTags
}

// skopeoTagList is used to unmarshal `skopeo list-tags` command output.
type skopeoTagList struct {
	Tags []string `yaml:"Tags"`
}

// skopeoFile is used to un/marshal YAML file format used by `skopeo sync`.
// Only partial support is implemented, since we don't need the full functionality.
// docs: https://github.com/containers/skopeo/blob/main/docs/skopeo-sync.1.md#yaml-file-content-used-source-for---src-yaml
type skopeoFile map[string]skopeoFileRegistry

type skopeoFileRegistry struct {
	// Images is a map of ImageName -> []Tags
	Images map[string][]string `yaml:"images"`
}

// listTags gets a list of available tags for a given registry+image, for
// example 'quay.io/giantswarm/curl'.
func listTags(image string) ([]string, error) {
	var tags []string
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		attemptLogger := logrus.WithField("attempt", attempt+1)

		c, stdout, stderr := command("skopeo", "list-tags", dockerTransport+image)
		err = c.Run()
		if err != nil && strings.Contains(stderr.String(), "repository name not known to registry") {
			// This image has never been pushed to registry - has no synced tags.
			return tags, nil
		} else if err != nil && strings.Contains(stderr.String(), "name unknown") {
			// This image has never been pushed to registry - has no synced tags.
			return tags, nil
		} else if err != nil {
			err = fmt.Errorf("error listing tags for %q: %w\n%s", image, err, stderr.String())
			attemptLogger.Warn(err)
			continue
		}

		stl := skopeoTagList{
			Tags: []string{},
		}
		err = yaml.Unmarshal(stdout.Bytes(), &stl)
		if err != nil {
			err = fmt.Errorf("error listing tags for %q: %w\n%s", image, err, stderr.String())
			attemptLogger.Warn(err)
			continue
		}

		tags = stl.Tags
		break
	}

	if err != nil {
		return []string{}, err
	}
	return tags, nil
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

// removeRegistryPrefix is a helper function to get rid of the
// registry name if it exists.
// Examples:
// "docker.io/alpine/alpine" -> "alpine/alpine"
// "grafana" -> "grafana"
func removeRegistryPrefix(name string) string {
	if !strings.ContainsRune(name, '/') {
		return name
	}
	name = "docker://" + name
	u, err := url.Parse(name)
	if err != nil {
		logrus.Fatalf("error parsing %q as URL", name)
	}
	return strings.TrimLeft(u.RequestURI(), "/")
}

// copyImage is a helper function used to invoke `skopeo copy`. Please note the
// `--all`, which makes skopeo include ALL SHAs included in the tag's digest,
// ensuring builds for all available platforms.
func copyImage(wg *sync.WaitGroup, errorCounter *atomic.Int64, source, destination string) {
	defer wg.Done()
	c, _, stderr := command("skopeo", "copy", "--all", "--retry-times", "3", source, destination)
	logrus.Debugf("copying %q to %q", source, destination)
	if err := c.Run(); err != nil {
		logrus.Errorf("error copying %q to %q: %v\n%s", source, destination, err, stderr.String())
		return
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
		return
	}
	logrus.Debugf("pushed %q", nameAndTag)
	// try to free up docker space
	c = exec.Command("docker", "image", "rm", nameAndTag)
	if err := c.Run(); err != nil {
		logrus.Tracef("error running %q", c.String())
	}
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
	// `retagger run` flags
	flag.IntVar(&flagExecutorCount, "executor-count", 1, "Number of executors in a parallelized run. Used with 'retagger run'.")
	flag.IntVar(&flagExecutorID, "executor-id", 0, "ID of the executor in a parallelized run. Used with 'retagger run'.")
	flag.BoolVar(&flagSkipExistingTags, "skip-existing-tags", true, "Skip tags which are already present in the target container registry. Used with 'retagger run'.")
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	lvl, err := logrus.ParseLevel(flagLogLevel)
	if err != nil {
		logrus.Warnf("could not parse log level %q, defaulting to %q", flagLogLevel, logrus.DebugLevel.String())
	} else {
		logrus.SetLevel(lvl)
	}
}

// commandRun is invoked when `retagger run` is called.
func commandRun() {
	// Validate commandRun-specific flags
	if flagExecutorID < 0 || flagExecutorID >= flagExecutorCount {
		logrus.Fatalf("%q flag has to be greater than 0 and lower than %q", "executor-id", "executor-count")
	}
	if flagExecutorCount < 1 {
		logrus.Fatalf("%q cannot be lower than 1", "executor-count")
	}
	if flagExecutorCount > 10 {
		logrus.Warnf("%q is set to %d, are you sure that's on purpose?", "executor-count", flagExecutorCount)
	}

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

// commandFilter is invoked when `retagger filter` is called.
func commandFilter(filepath string) {
	if filepath == "" {
		logrus.Fatal("You need to specify filepath: 'retagger filter <path>'")
	}
	logger := logrus.WithField("file", filepath)

	logger.Infof("Listing images & tags")
	missingTagsPerImage := map[string][]string{}
	{
		filterPrefix := "auniqueprefixa"
		c, _, stderr := command("skopeo", "sync", "--all", "--dry-run", "--src", "yaml", "--dest", "docker", filepath, filterPrefix)
		if err := c.Run(); err != nil {
			logger = logger.WithField("stderr", stderr.String())
			logger.Fatalf("error running 'skopeo sync --dry-run': %v", err)
		}

		tagsPerImage := map[string][]string{}

		matches := skopeoSyncOutputPattern.FindAllStringSubmatch(stderr.String(), -1)
		if matches == nil {
			logger = logger.WithField("stderr", stderr.String())
			logger.Fatalf("found no images or tags in 'skopeo sync' output")
		}
		for _, m := range matches {
			// by index: 0 - entire line, 1 - image name, 2 - tag
			image := strings.TrimPrefix(m[1], filterPrefix+"/")
			tag := m[2]
			tagsPerImage[image] = append(tagsPerImage[image], tag)
		}

		logger.Infof("Found %d images, checking how many tags are missing", len(tagsPerImage))
		missingTagCount := 0
		for image, tags := range tagsPerImage {
			logger.WithField("image", image).Debugf("searching for missing tags")
			quayTags, err := listTags(fmt.Sprintf("%s/%s", quayURL, imageBaseName(image)))
			if err != nil {
				logger.WithField("image", image).Errorf("error listing Quay tags: %v", err)
				continue
			}
			azureTags, err := listTags(fmt.Sprintf("%s/%s", azureURL, imageBaseName(image)))
			if err != nil {
				logger.WithField("image", image).Errorf("error listing AzureCR tags: %v", err)
				continue
			}
			aliyunTags, err := listTags(fmt.Sprintf("%s/%s", aliyunURL, imageBaseName(image)))
			if err != nil {
				logger.WithField("image", image).Errorf("error listing Aliyun tags: %v", err)
				continue
			}
			i := &CustomImage{}
			missingTags := i.FindMissingTags(tags, quayTags, azureTags, aliyunTags)
			missingTagCount += len(missingTags)
			missingTagsPerImage[image] = missingTags
		}
		logger.Infof("Found %d missing tags", missingTagCount)
	}

	logger.Debugf("Saving filtered file")
	filteredFile := skopeoFile{}
	{
		b, err := os.ReadFile(filepath)
		if err != nil {
			logger.Fatalf("error reading file: %v", err)
		}
		if err := yaml.Unmarshal(b, &filteredFile); err != nil {
			logger.Fatalf("error unmarshaling file: %v", err)
		}

		// Files are split by registry, so there is exactly one registry
		// defined in each one of them.
		registryName := maps.Keys(filteredFile)[0]
		// Ensure empty images map.
		filteredFile[registryName] = skopeoFileRegistry{
			Images: make(map[string][]string),
		}
		for fullImageName, tags := range missingTagsPerImage {
			if len(tags) > 0 {
				strippedImageName := strings.TrimPrefix(fullImageName, registryName)
				strippedImageName = strings.TrimLeft(strippedImageName, "/")
				filteredFile[registryName].Images[strippedImageName] = tags
			}
		}
	}

	b, err := yaml.Marshal(&filteredFile)
	if err != nil {
		logger.Fatalf("error marshaling file: %v", err)
	}
	err = os.WriteFile(filepath+filteredFileSuffix, b, 0644)
	if err != nil {
		logrus.Fatalf("error writing file: %v", err)
	}
	logger.Infof("Saved filtered file with missing tags")
}

func main() {
	if len(flag.Args()) == 0 {
		fmt.Println("retagger run             Retag custom images\nretagger filter <path>   Filter missing tags for skopeo YAML file\n\n")
		flag.Usage()
		os.Exit(0)
	}

	switch flag.Arg(0) {
	case "run":
		commandRun()
	case "filter":
		commandFilter(flag.Arg(1))
	default:
		logrus.Fatalf("unknown command: %v", flag.Args())
	}
}
