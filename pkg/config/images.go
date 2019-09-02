package config

import (
	"fmt"
	"strings"
)

// Image defines the data we process about a docker image.
type Image struct {
	Name             string `yaml:"name"`
	Comment          string `yaml:"comment,omitempty"`
	OverrideRepoName string `yaml:"overrideRepoName,omitempty"`
	Tags             []Tag  `yaml:"tags"`
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

// Images is the data structure that will hold all image definitions.
var Images []Image

func ImageName(organisation string, image string) string {
	parts := strings.Split(image, "/")

	return fmt.Sprintf("%v/%v", organisation, parts[len(parts)-1])
}

func RetaggedName(registry, organisation string, image string) string {
	parts := strings.Split(image, "/")

	return fmt.Sprintf("%v/%v/%v", registry, organisation, parts[len(parts)-1])
}

func ImageWithTag(image, tag string) string {
	return fmt.Sprintf("%v:%v", image, tag)
}

func ShaName(imageName, sha string) string {
	return fmt.Sprintf("%v@sha256:%v", imageName, sha)
}
