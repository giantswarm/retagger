package images

import (
	"fmt"
	"strings"
)

type Images []Image

// Image defines the data we process about a docker image.
type Image struct {
	Name                 string       `yaml:"name"`
	Comment              string       `yaml:"comment,omitempty"`
	TagTrimVersionPrefix bool         `yaml:"tagTrimVersionPrefix,omitempty"`
	OverrideRepoName     string       `yaml:"overrideRepoName,omitempty"`
	Patterns             []TagPattern `yaml:"patterns"`
	Tags                 []Tag        `yaml:"tags"`
}

// TagPattern represents a group of tags defined by a regular expression,
// and stores some configuration for how to handle this group.
type TagPattern struct {
	Pattern      string        `yaml:"pattern"`
	Tag          string        `yaml:"tag"`
	CustomImages []CustomImage `yaml:"customImages"`
}

// Tag represents a specific version of a docker image, represented by a tag
// and verified through the SHA checksum.
type Tag struct {
	// Sha is the image SHA to pull from the original source.
	Sha string `yaml:"sha"`
	// Tag is the tag we apply to the pulled image.
	Tag string `yaml:"tag"`
	// CustomImages is the list of custom images we build from original image base.
	CustomImages []CustomImage `yaml:"customImages"`
}

type CustomImage struct {
	// TagSuffix is a string suffix we add to the original image tag.
	TagSuffix string `yaml:"tagSuffix"`
	// DockerfileOptions - list of strings we add for Dockerfile to build custom image
	DockerfileOptions []string `yaml:"dockerfileOptions"`
}

func Name(organisation string, image string) string {
	parts := strings.Split(image, "/")

	return fmt.Sprintf("%s/%s", organisation, parts[len(parts)-1])
}

func RetaggedName(registry, organisation string, image string) string {
	parts := strings.Split(image, "/")

	return fmt.Sprintf("%s/%s/%s", registry, organisation, parts[len(parts)-1])
}

func NameWithTag(image, tag string) string {
	return fmt.Sprintf("%s:%s", image, tag)
}

func ShaName(imageName, sha string) string {
	return fmt.Sprintf("%s@sha256:%s", imageName, sha)
}
