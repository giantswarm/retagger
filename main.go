package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	platformListPattern = regexp.MustCompile(`\w+\s+\w+\s+\w+\s+(.*)`) // name/node driver/endpoint status (platforms)
)

type DockerBuildx struct {
	supportedPlatforms map[string]Platform
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

func NewDockerBuildx() (*DockerBuildx, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c := exec.Command("docker", "buildx", "ls")
	c.Stdout = stdout
	c.Stderr = stderr
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("error running %q: %w\nstderr:\n%s", c.String(), err, stderr.String())
	}
	// extract supported platforms
	// Example:
	// NAME/NODE DRIVER/ENDPOINT STATUS  PLATFORMS
	// default * docker
	//   default default         running linux/amd64, linux/386, linux/arm64, linux/riscv64, linux/ppc64le, linux/s390x, linux/arm/v7, linux/arm/v6
	matches := platformListPattern.FindAllStringSubmatch(stdout.String(), -1)
	if matches == nil || len(matches) < 2 {
		return nil, fmt.Errorf("could not extract supported platforms using 'docker buildx ls'")
	}
	dbx := &DockerBuildx{
		supportedPlatforms: map[string]Platform{},
	}
	platformStrings := strings.Split(matches[1][1], ", ")
	for _, platformString := range platformStrings {
		dbx.supportedPlatforms[platformString] = NewPlatform(platformString)
	}
	fmt.Printf("%+v", dbx.supportedPlatforms)
	return dbx, nil
}

func main() {
	l := logrus.New()
	l.Warnf("hello")
	_, err := NewDockerBuildx()
	if err != nil {
		l.Fatal(err)
	}
}
