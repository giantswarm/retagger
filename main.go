package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	platformListPattern = regexp.MustCompile(`\w+\s+\w+\s+\w+\s+(.*)`) // name/node driver/endpoint status (platforms)
)

const (
	defaultPlatform = "linux/amd64"
	dockerTransport = "docker://"
)

type DockerBuildx struct {
	logger             *logrus.Logger
	supportedPlatforms map[string]Platform
	customDockerfiles  []Dockerfile
}

type Platform struct {
	System       string
	Architecture string
	Variant      string
}

func (p Platform) String() string {
	if p.Variant == "" {
		return fmt.Sprintf("%s/%s", p.System, p.Architecture)
	}
	return fmt.Sprintf("%s/%s/%s", p.System, p.Architecture, p.Variant)
}

func (p Platform) ArchitectureAndVariant() string {
	if p.Variant == "" {
		return p.Architecture
	}
	return fmt.Sprintf("%s/%s", p.Architecture, p.Variant)
}

func NewPlatform(name string) Platform {
	p := Platform{}
	elems := strings.Split(name, "/")
	p.System = elems[0]
	p.Architecture = elems[1]
	if len(elems) == 3 {
		p.Variant = elems[2]
	}
	return p
}

type Dockerfile struct {
	Image            string   `yaml:"image"`
	TagPattern       string   `yaml:"tag_pattern"`
	DockerfileExtras []string `yaml:"dockerfile_extras,omitempty"`
	AddTagSuffix     string   `yaml:"add_tag_suffix,omitempty"`
	OverrideRepoName string   `yaml:"override_repo_name,omitempty"`
}

type skopeoTagList struct {
	Tags []string `yaml:"Tags"`
}

func (d *Dockerfile) BuildAndTag() error {
	// (code) List tags and find the ones that match
	var tags []string
	{
		c, stdout, stderr := command("skopeo", "list-tags", dockerTransport+d.Image)
		if err := c.Run(); err != nil {
			return fmt.Errorf("error listing tags for %q: %w\n%s", d.Image, err, stderr.String())
		}
		stl := skopeoTagList{
			Tags: []string{},
		}
		if err := yaml.Unmarshal(stdout.Bytes(), &stl); err != nil {
			return fmt.Errorf("error unmarshaling tags: %w", err)
		}
		tags = stl.Tags
	}

	pattern, err := regexp.Compile(d.TagPattern)
	if err != nil {
		return fmt.Errorf("error compiling regexp pattern %q: %w", d.TagPattern, err)
	}
	for _, tag := range tags {
		if !pattern.MatchString(tag) {
			continue
		}
		// (code) Prepare temporary dockerfile by generating 'FROM X:Y, buildplatform' and appending dockerfile_path
		// (docker buildx binary) Rebuild image with temporary dockerfile for each tag
		// (skopeo binary) Push the images to QUAY and ALIYUN
		if len(d.DockerfileExtras) == 0 {
			// no customization, just sync
			c, _, stderr := command("skopeo", "list-tags", dockerTransport+d.Image)
			if err := c.Run(); err != nil {
				return fmt.Errorf("error listing tags for %q: %w\n%s", d.Image, err, stderr.String())
			}
		}
	}

	return nil
}

func NewDockerBuildx(logger *logrus.Logger) (*DockerBuildx, error) {
	c, stdout, stderr := command("docker", "buildx", "ls")
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("error running %q: %w\nstderr:\n%s", c.String(), err, stderr.String())
	}
	dbx := &DockerBuildx{
		logger:             logger,
		supportedPlatforms: map[string]Platform{},
		customDockerfiles:  []Dockerfile{},
	}
	// extract supported platforms
	{
		// Example:
		// NAME/NODE DRIVER/ENDPOINT STATUS  PLATFORMS
		// default * docker
		//   default default         running linux/amd64, linux/386, linux/arm64, linux/riscv64, linux/ppc64le, linux/s390x, linux/arm/v7, linux/arm/v6
		matches := platformListPattern.FindAllStringSubmatch(stdout.String(), -1)
		if matches == nil || len(matches) < 2 {
			return nil, fmt.Errorf("could not extract supported platforms using 'docker buildx ls'")
		}
		platformStrings := strings.Split(matches[1][1], ", ")
		for _, platformString := range platformStrings {
			dbx.supportedPlatforms[platformString] = NewPlatform(platformString)
		}
	}
	// load custom dockerfile specs
	{
		b, err := os.ReadFile("customized-dockerfiles.yaml")
		if err != nil {
			return nil, fmt.Errorf("error reading customized-dockerfiles.yaml: %w", err)
		}
		if err := yaml.Unmarshal(b, &dbx.customDockerfiles); err != nil {
			return nil, fmt.Errorf("error unmarshaling customized-dockerfiles.yaml: %w", err)
		}
	}

	return dbx, nil
}

func command(name string, args ...string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c := exec.Command(name, args...)
	c.Stdout = stdout
	c.Stderr = stderr
	return c, stdout, stderr
}

func main() {
	l := logrus.New()
	l.Warnf("hello")
	dbx, err := NewDockerBuildx(l)
	if err != nil {
		l.Fatal(err)
	}
	l.Warnf("%+v", dbx.supportedPlatforms)
	l.Warnf("%+v", dbx.customDockerfiles)
}
