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
	Image            string `yaml:"image"`
	TagPattern       string `yaml:"tag_pattern"`
	SHA              string `yaml:"sha,omitempty"`
	DockerfileExtras string `yaml:"dockerfile_extras"`
	AddTagSuffix     string `yaml:"add_tag_suffix,omitempty"`
}

func (d *Dockerfile) BuildAndTag() error {
	// (code) Check if dockerfile_path exists
	// (code) List tags and find the ones that match
	// (code) Prepare temporary dockerfile by generating 'FROM X:Y, buildplatform' and appending dockerfile_path
	// (docker buildx binary) Rebuild image with temporary dockerfile for each tag
	// (skopeo binary) Push the images to QUAY and ALIYUN
	return nil
}

func NewDockerBuildx() (*DockerBuildx, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c := exec.Command("docker", "buildx", "ls")
	c.Stdout = stdout
	c.Stderr = stderr
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("error running %q: %w\nstderr:\n%s", c.String(), err, stderr.String())
	}
	dbx := &DockerBuildx{
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

func main() {
	l := logrus.New()
	l.Warnf("hello")
	dbx, err := NewDockerBuildx()
	if err != nil {
		l.Fatal(err)
	}
	l.Warnf("%+v", dbx.supportedPlatforms)
	l.Warnf("%+v", dbx.customDockerfiles)
}
