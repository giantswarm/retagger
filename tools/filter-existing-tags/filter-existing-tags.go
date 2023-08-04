package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

const (
	dockerTransport = "docker://"
	quayURL         = "quay.io/giantswarm"
	aliyunURL       = "giantswarm-registry.cn-shanghai.cr.aliyuncs.com/giantswarm"
)

var (
	flagSrc string
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	pflag.StringVar(&flagSrc, "src", "", "skopeo sync file to split")
	pflag.Parse()

	if flagSrc == "" {
		logrus.Fatalf("%q flag has to be set", "src")
	}
}

// skopeoTagList is used to unmarshal `skopeo list-tags` command output.
type skopeoTagList struct {
	Tags []string `yaml:"Tags"`
}

// skopeoImagesFile is used to unmarshal `images/skopeo-*` files.
type skopeoRegistry struct {
	Images           map[string][]string `yaml:"images"`
	ImagesByTagRegex map[string]string   `yaml:"images-by-tag-regex"`
	ImagesBySemver   map[string][]string `yaml:"images-by-semver"`
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
		if err != nil {
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

func makeImageNames(registry, image string) (sourceName, quayName, aliyunName string) {
	sourceName = fmt.Sprintf("%s/%s", registry, image)
	quayName = fmt.Sprintf("%s/%s", quayURL, image)
	aliyunName = fmt.Sprintf("%s/%s", aliyunURL, image)
	return sourceName, quayName, aliyunName
}

// findMissingTags returns a list of items of the 'tags' slice that are missing
// from at least one of the 'present' slices.
func findMissingTags(tags []string, present ...[]string) []string {
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

		if tagIsMissing {
			filteredTags = append(filteredTags, tag)
		}
	}
	return filteredTags
}

func main() {
	b, err := os.ReadFile(flagSrc)
	if err != nil {
		logrus.Fatalf("error reading %q: %s", flagSrc, err)
	}

	var registryMap map[string]skopeoRegistry
	if err := yaml.Unmarshal(b, &registryMap); err != nil {
		logrus.Fatalf("error reading %q: %s", flagSrc, err)
	}

	logrus.Infof("Filtering tags in %q", flagSrc)
	newRegistryMap := map[string]skopeoRegistry{}

	for registry := range registryMap {
		newRegistryMap[registry] = skopeoRegistry{
			Images:           map[string][]string{},
			ImagesByTagRegex: map[string]string{},
			ImagesBySemver:   map[string][]string{},
		}

		for image, tags := range registryMap[registry].Images {
			_, quayName, aliyunName := makeImageNames(registry, image)
			quayTags, err := listTags(quayName)
			if err != nil {
				logrus.Errorf("error listing %q tags: %v", quayName, err)
				continue
			}
			aliyunTags, err := listTags(aliyunName)
			if err != nil {
				logrus.Errorf("error listing %q tags: %v", aliyunName, err)
				continue
			}
			missingTags := findMissingTags(tags, quayTags, aliyunTags)
			if len(missingTags) != 0 {
				newRegistryMap[registry].Images[image] = missingTags
			}
		}

		// by tag regex
		for image, tags := range registryMap[registry].ImagesByTagRegex {
		}

		// by semver

	}

	// Marshal newRegistryMap & save to flagSrc filepath
}
