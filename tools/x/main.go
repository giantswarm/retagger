package main

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	imageTagPattern = regexp.MustCompile(`Would have copied image.*?to="docker://(.*?):(.*?)"`)
)

const (
	dockerPrefix = "xtractor"
)

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

func main() {
	// TODO: filename should be command parameter
	filename := "images/skopeo-docker-io.yaml"
	c, _, stderr := command("skopeo", "sync", "--all", "--dry-run", "--src", "yaml", "--dest", "docker", filename, dockerPrefix)
	if err := c.Run(); err != nil {
		logrus.Fatalf("error listing images and tags in %q: %v\n%s", filename, err, stderr.String())
	}

	matches := imageTagPattern.FindAllStringSubmatch(stderr.String(), -1)
	if matches == nil {
		logrus.Fatalf("found no images to check")
	}

	for _, m := range matches {
		// by index: 0 - entire line, 1 - image name, 2 - tag
		image := strings.TrimPrefix(m[1], dockerPrefix+"/")
		tag := m[2]
		logrus.Infof("%s:%s", image, tag)
	}

}
