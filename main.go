// Package main is the retagger program.
//
// The program provides two commands:
//   - `retagger run` - Performs retagging / renaming of the images defined in images/renamed-images.yaml.
//   - `retagger filter <path>` - Processes skopeo YAML files in images/skopeo-* and creates a
//     list of image syncing tasks to be performed. This is simple copyingf of images from one
//     repository to another.
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
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

var (
	skopeoSyncOutputPattern = regexp.MustCompile(`Would have copied image.*?from="docker://(.*?)[@:](.*?)".*`)
	temporaryWorkingDir     = path.Join(os.TempDir(), "retagger")

	flagFile                  string
	destinationRegistriesFile string
	flagLogLevel              string
	flagExecutorCount         int
	flagExecutorID            int
	flagSkipExistingTags      bool

	logStdOut = logrus.New()
	logStdErr = logrus.New()
)

const (
	renamedImagesFile  = "images/renamed-images.yaml"
	defaultPlatform    = "linux/amd64"
	dockerTransport    = "docker://"
	filteredFileSuffix = ".filtered"

	aliyunURL = "giantswarm-registry.cn-shanghai.cr.aliyuncs.com/giantswarm"
	azureURL  = "gsoci.azurecr.io/giantswarm"
	quayURL   = "quay.io/giantswarm"
)

// RenamedImage represents a set of rules used to rebuild/retag multiple tags of
// the same image that match the specified tag pattern.
type RenamedImage struct {
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

func (img *RenamedImage) Validate() error {
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
func (img *RenamedImage) RetagUsingSHA() error {
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

	// We'll use skopeo copy for this, because it's awesome.
	source := fmt.Sprintf("%s%s@sha256:%s", dockerTransport, img.Image, img.SHA)
	wg := sync.WaitGroup{}
	wg.Add(3)
	// Quay
	destination := fmt.Sprintf("%s%s/%s:%s", dockerTransport, quayURL, destinationName, destinationTag)
	go copyImage(&wg, errorCounter, source, destination)
	// AzureCR
	destination = fmt.Sprintf("%s%s/%s:%s", dockerTransport, azureURL, destinationName, destinationTag)
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

// RetagUsingTags finds all tags matching the img.TagOrPattern or
// img.Semver, retags, and pushes them to Quay and Aliyun container registries.
// Any optional parameters configured will be applied as well, e.g. tag suffix.
func (img *RenamedImage) RetagUsingTags() error {
	// Parse destination registries
	destinationRegistries, err := ParseDestinationRegistries(destinationRegistriesFile)
	if err != nil {
		logrus.Fatalf("failed to parse destination registries file %q: %v", destinationRegistriesFile, err)
		os.Exit(1)
	}

	logrus.WithField("destinationRegistries", destinationRegistries).Println("successfully parsed destination registries")

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

	// Filter the tags using TagOrPattern or Semver+Filter.
	tags, err = img.FilterTags(tags)
	if err != nil {
		return fmt.Errorf("error filtering tags: %w", err)
	}

	// Exclude tags existing in all registries
	if flagSkipExistingTags {
		tagsFromTargets := CollectTagsFromTargetRegistries(destinationRegistries, destinationName)

		tags = img.FindMissingTags(tags, tagsFromTargets...)
		logrus.Infof("Found %d missing tags for image %q", len(tags), img.Image)
	}

	errorCounter := &atomic.Int64{}

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

		// We'll use skopeo copy for this, because it's awesome.
		source := fmt.Sprintf("%s%s:%s", dockerTransport, img.Image, tag)
		wg := sync.WaitGroup{}
		wg.Add(3)
		// Quay
		destination := fmt.Sprintf("%s%s/%s:%s", dockerTransport, quayURL, destinationName, destinationTag)
		go copyImage(&wg, errorCounter, source, destination)
		// Azure
		destination = fmt.Sprintf("%s%s/%s:%s", dockerTransport, azureURL, destinationName, destinationTag)
		go copyImage(&wg, errorCounter, source, destination)
		// Aliyun
		destination = fmt.Sprintf("%s%s/%s:%s", dockerTransport, aliyunURL, destinationName, destinationTag)
		go copyImage(&wg, errorCounter, source, destination)
		wg.Wait()

		// We'll skip to the next tag
		continue
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
func (img *RenamedImage) FilterTags(tags []string) ([]string, error) {
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

type DestinationRegistries struct {
	Destinations []DestinationRegistry `yaml:"destinations"`
}

type DestinationRegistry struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
}

func ParseDestinationRegistries(filePath string) (DestinationRegistries, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return DestinationRegistries{Destinations: []DestinationRegistry{}}, err
	}

	var destinationRegistries DestinationRegistries
	err = yaml.Unmarshal(content, &destinationRegistries)
	if err != nil {
		return DestinationRegistries{Destinations: []DestinationRegistry{}}, err
	}

	return destinationRegistries, nil
}

func CollectTagsFromTargetRegistries(destinations DestinationRegistries, imageName string) [][]string {
	var tagsPerRegistry [][]string

	for _, destination := range destinations.Destinations {
		tags, err := listTags(fmt.Sprintf("%s/%s", destination.Url, imageName))
		if err != nil {
			logStdErr.WithField("image", imageName).Errorf("error listing %s tags: %v", destination.Name, err)
			continue
		}
		tagsPerRegistry = append(tagsPerRegistry, tags)
	}

	return tagsPerRegistry
}

// findMissingTags returns a list of items of the 'tags' slice that are missing
// from at least one of the 'present' slices.
func (img *RenamedImage) FindMissingTags(tags []string, present ...[]string) []string {
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
		attemptLogger := logStdOut.WithField("attempt", attempt+1)

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
	flag.StringVar(&flagFile, "filename", renamedImagesFile, "Sets the file to use for renaming")
	flag.StringVar(&flagLogLevel, "log-level", "debug", "Sets log level")
	// destination registries file
	flag.StringVar(&destinationRegistriesFile, "destination-registries", "destinations/empty.yaml", "Sets the file to use for destination registries")
	// `retagger run` flags
	flag.IntVar(&flagExecutorCount, "executor-count", 1, "Number of executors in a parallelized run. Used with 'retagger run'.")
	flag.IntVar(&flagExecutorID, "executor-id", 0, "ID of the executor in a parallelized run. Used with 'retagger run'.")
	flag.BoolVar(&flagSkipExistingTags, "skip-existing-tags", true, "Skip tags which are already present in the target container registry. Used with 'retagger run'.")
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	logStdOut.SetFormatter(&logrus.TextFormatter{})
	logStdOut.SetLevel(logrus.DebugLevel)

	lvl, err := logrus.ParseLevel(flagLogLevel)
	if err != nil {
		logrus.Warnf("could not parse log level %q, defaulting to %q", flagLogLevel, logrus.DebugLevel.String())
	} else {
		logrus.SetLevel(lvl)
	}

	logStdOut.SetLevel(lvl)
	logStdOut.Out = os.Stdout
	logStdErr.Out = os.Stderr
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

	logger.Infof("Using file %q", flagFile)

	// Load renamed image definitions from a file
	var renamedImages []RenamedImage
	{
		b, err := os.ReadFile(flagFile)
		if err != nil {
			logger.Fatalf("error reading %q: %s", flagFile, err)
		}
		if err := yaml.Unmarshal(b, &renamedImages); err != nil {
			logger.Fatalf("error unmarshaling %q: %s", flagFile, err)
		}
	}

	logger.Infof("Found %d images to rename and copy", len(renamedImages))

	// Iterate over every image x tag and retag/rebuild it
	errorCounter := 0
	for i, image := range renamedImages {
		// Skip images meant for other executors
		if i%flagExecutorCount != flagExecutorID {
			continue
		}
		if err := image.Validate(); err != nil {
			logger.Errorf("[%d/%d] %q error: %s", i+1, len(renamedImages), image.Image, err)
			errorCounter++
			continue
		}
		logger.Printf("[%d/%d] Retagging %q", i+1, len(renamedImages), image.Image)
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
	logger.Infof("Done retagging %d images with no errors", len(renamedImages))
}

// commandFilter is invoked when `retagger filter` is called.
//
// The function reads a skopeo configuration file and runs `skopeo sync --dry-run`
// with it. The output is processed for image tags to be synced. Each tag checked
// against Quay, AzureCR, and Aliyun. If a tag is missing in any of the registries,
// it is added to the list of tags to be synced. The list is stored in a file next
// to the input file, with the name suffixed with `.filtered`.
func commandFilter(filepath string) {
	destinationRegistries, err := ParseDestinationRegistries(destinationRegistriesFile)
	if err != nil {
		logrus.Fatalf("failed to parse destination registries file %q: %v", destinationRegistriesFile, err)
		os.Exit(1)
	}

	logrus.WithField("destinationRegistries", destinationRegistries).Println("successfully parsed destination registries")

	if filepath == "" {
		logrus.Fatal("You need to specify filepath: 'retagger filter <path>'")
	}
	logStdOut.WithField("file", filepath)
	logStdErr.WithField("file", filepath)

	logStdOut.Infof("Listing images & tags")
	missingTagsPerImage := map[string][]string{}
	{
		filterPrefix := "auniqueprefixa"
		c, _, stderr := command("skopeo", "sync", "--all", "--dry-run", "--src", "yaml", "--dest", "docker", filepath, filterPrefix)
		if err := c.Run(); err != nil {
			logStdErr.WithField("stderr", stderr.String())
			logStdErr.Fatalf("error running 'skopeo sync --dry-run': %v", err)
		}

		tagsPerImage := map[string][]string{}

		matches := skopeoSyncOutputPattern.FindAllStringSubmatch(stderr.String(), -1)
		if matches == nil {
			logStdErr.WithField("stderr", stderr.String())
			logStdErr.Fatalf("found no images or tags in 'skopeo sync' output")
		}
		for _, m := range matches {
			// by index: 0 - entire line, 1 - image name, 2 - tag
			image := strings.TrimPrefix(m[1], filterPrefix+"/")
			tag := m[2]
			tagsPerImage[image] = append(tagsPerImage[image], tag)
		}

		logStdOut.Infof("Found %d images, checking how many tags are missing", len(tagsPerImage))
		missingTagCount := 0
		for image, tags := range tagsPerImage {
			logStdOut.WithField("image", image).Debugf("searching for missing tags")

			tagsFromTargets := CollectTagsFromTargetRegistries(destinationRegistries, imageBaseName(image))
			/*
				Context: https://github.com/giantswarm/giantswarm/issues/31283

				I am not sure what to think about the above code. The comment to the `commandFilter`
				function states: `If a tag is missing in any of the registries, it is added to the
				list of tags to be synced.`, what means in order to proceed with a given image we need
				to establish the truth of:

				ImgTagsMissingIn(Quay) ∨ ImgTagsMissingIn(Azure) ∨ ImgTagsMissingIn(Aliyun) ≡ T

				In case any of the disjuncts is undefined, when error is returned, then obviously the
				predicate is undefined and the image is skipped, so the code meets specification.

				On the other hand, maybe the undefined state could be treated as false, what would change
				the postcondition to:

				ImgTagsMissingIn(Quay) = T ∨ ImgTagsMissingIn(Azure) = T ∨ ImgTagsMissingIn(Aliyun) = T ≡ T

				The question of whether that's possible or not boils down to the question: does it pose
				any danger to include an image if we do not know its state in the final registry? The
				repository for it may not exist at all in the registry, it may be already present there,
				with all the tags, or registry may not be accessible at all at the moment, etc.

				Maybe there is no danger to it, what's implied by how the retagger gets executed. Assuming
				tags are missing in one of the three registries, the retagger is anyway executed against
				all three of them in the CircleCI, meaning it must have a way, which is the `skopeo`,
				I believe, to account for tags already present in the destination registry. If so, the state
				of registries, that already have tags in question, do not really matter. Tags state should
				get correctly handled by the `skopeo` later, and in case there is a problem with the registry
				or repository, we should get an error.

				On the other hand, it wouldn't make any difference for the original issue, linked on top, for
				in its case the repository was missing in all three registries, so the result would still be
				the same.
			*/
			i := &RenamedImage{}
			missingTags := i.FindMissingTags(tags, tagsFromTargets...)
			missingTagCount += len(missingTags)
			missingTagsPerImage[image] = missingTags
		}
		logStdOut.Infof("Found %d missing tags", missingTagCount)
	}

	logStdOut.Debugf("Saving filtered file")
	filteredFile := skopeoFile{}
	{
		b, err := os.ReadFile(filepath)
		if err != nil {
			logStdErr.Fatalf("error reading file: %v", err)
		}
		if err := yaml.Unmarshal(b, &filteredFile); err != nil {
			logStdErr.Fatalf("error unmarshaling file: %v", err)
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
		logStdErr.Fatalf("error marshaling file: %v", err)
	}
	err = os.WriteFile(filepath+filteredFileSuffix, b, 0644)
	if err != nil {
		logStdErr.Fatalf("error writing file: %v", err)
	}
	logStdOut.Infof("Saved filtered file with missing tags")
}

func main() {
	if len(flag.Args()) == 0 {
		fmt.Println("retagger run             Retag images\nretagger filter <path>   Filter missing tags for skopeo YAML file\n\n")
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
