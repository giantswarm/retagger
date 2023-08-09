package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

var (
	imageTagPattern = regexp.MustCompile(`Would have copied image.*?from="docker://(.*?):(.*?)".*`)
	flagSrc         string
)

const (
	dockerPrefix    = "xtractor"
	dockerTransport = "docker://"
	quayURL         = "quay.io/giantswarm"
	aliyunURL       = "giantswarm-registry.cn-shanghai.cr.aliyuncs.com/giantswarm"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	pflag.StringVar(&flagSrc, "src", "", "skopeo sync file to filter")
	pflag.Parse()
}

// skopeoTagList is used to unmarshal `skopeo list-tags` command output.
type skopeoTagList struct {
	Tags []string `yaml:"Tags"`
}

// skopeoRegistry is used to marshal/unmarshal yaml used by `skopeo sync` command.
// See: https://github.com/containers/skopeo/blob/main/docs/skopeo-sync.1.md#yaml-file-content-used-source-for---src-yaml
type skopeoRegistry struct {
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

// findMissingTags returns a list of items of the 'tags' slice that are missing
// from at least one of the 'present' slices.
func findMissingTags(tags []string, present ...[]string) []string {
	var filteredTags []string
	for _, tag := range tags {
		tagIsMissing := false

		for _, existingTags := range present {
			if !slices.Contains(existingTags, tag) {
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

// imageBaseName is a helper function extracting base image name.
// Example: "registry.k8s.io/kube-apiserver" -> "kube-apiserver"
func imageBaseName(name string) string {
	if !strings.ContainsRune(name, '/') {
		return name
	}
	elems := strings.Split(name, "/")
	return elems[len(elems)-1]
}

func main() {
	// TODO: filename should be command parameter
	filename := flagSrc
	c, _, stderr := command("skopeo", "sync", "--all", "--dry-run", "--src", "yaml", "--dest", "docker", filename, dockerPrefix)
	if err := c.Run(); err != nil {
		logrus.Fatalf("error listing images and tags in %q: %v\n%s", filename, err, stderr.String())
	}

	imageTagMap := map[string][]string{}

	matches := imageTagPattern.FindAllStringSubmatch(stderr.String(), -1)
	if matches == nil {
		logrus.Fatalf("found no images to check")
	}

	for _, m := range matches {
		// by index: 0 - entire line, 1 - image name, 2 - tag
		image := strings.TrimPrefix(m[1], dockerPrefix+"/")
		tag := m[2]
		imageTagMap[image] = append(imageTagMap[image], tag)
	}

	missingTagMap := map[string][]string{}

	for image, tags := range imageTagMap {
		logrus.Debugf("searching for missing tags of %q", image)
		quayTags, err := listTags(fmt.Sprintf("%s/%s", quayURL, imageBaseName(image)))
		if err != nil {
			logrus.Errorf("error listing quay tags for %q: %v", image, err)
			continue
		}

		aliyunTags, err := listTags(fmt.Sprintf("%s/%s", aliyunURL, imageBaseName(image)))
		if err != nil {
			logrus.Errorf("error listing aliyun tags for %q: %v", image, err)
			continue
		}

		missingTagMap[image] = findMissingTags(tags, quayTags, aliyunTags)
	}

	skopeoFile := map[string]skopeoRegistry{}

	b, err := os.ReadFile(filename)
	if err != nil {
		logrus.Fatalf("error reading %q: %v", filename, err)
	}
	if err := yaml.Unmarshal(b, &skopeoFile); err != nil {
		logrus.Fatalf("error unmarshaling %q: %v", filename, err)
	}
	// there should be only one registry name there
	for registry := range skopeoFile {
		for image, tags := range missingTagMap {
			if skopeoFile[registry].Images == nil {
				skopeoFile[registry] = skopeoRegistry{
					Images: map[string][]string{},
				}
			}
			skopeoFile[registry].Images[image] = tags
		}
	}

	b, err = yaml.Marshal(&skopeoFile)
	if err != nil {
		logrus.Fatalf("error marshaling %q: %v", filename, err)
	}
	err = os.WriteFile(filename, b, 0644)
	if err != nil {
		logrus.Fatalf("error writing %q: %v", filename, err)
	}

}
