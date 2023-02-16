package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	platformListPattern = regexp.MustCompile(`\w+\s+\w+\s+\w+\s+(.*)`) // name/node driver/endpoint status (platforms)
	temporaryWorkingDir = path.Join(os.TempDir(), "retagger")
)

const (
	defaultPlatform = "linux/amd64"
	dockerTransport = "docker://"
	quayURL         = "quay.io/giantswarm"
	aliyunURL       = "giantswarm-registry.cn-shanghai.cr.aliyuncs.com/giantswarm"
)

type Retagger struct {
	customDockerfiles []Dockerfile
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
tagLoop:
	for _, tag := range tags {
		if !pattern.MatchString(tag) {
			continue
		}

		destinationName := d.Image
		if d.OverrideRepoName != "" {
			destinationName = d.OverrideRepoName
		}
		destinationTag := tag
		if d.AddTagSuffix != "" {
			destinationTag = tag + "-" + d.AddTagSuffix
		}

		// (code) Prepare temporary dockerfile by generating 'FROM X:Y, buildplatform' and appending dockerfile_path
		// (docker buildx binary) Rebuild image with temporary dockerfile for each tag
		// (skopeo binary) Push the images to QUAY and ALIYUN
		if len(d.DockerfileExtras) == 0 {
			source := fmt.Sprintf("%s%s:%s", dockerTransport, d.Image, tag)
			wg := sync.WaitGroup{}
			wg.Add(2)
			// Quay
			destination := fmt.Sprintf("%s%s/%s:%s", dockerTransport, quayURL, destinationName, destinationTag)
			go copyImage(&wg, source, destination)
			// Aliyun
			destination = fmt.Sprintf("%s%s/%s:%s", dockerTransport, aliyunURL, destinationName, destinationTag)
			go copyImage(&wg, source, destination)
			wg.Wait()
			continue
		}

		// build temporary dockerfile
		var dockerfile string
		{
			// generate the Dockerfile
			tmp, err := os.CreateTemp(temporaryWorkingDir, "Dockerfile.*")
			if err != nil {
				logrus.Errorf("error creating temporary Dockerfile: %v", err)
				continue tagLoop
			}
			dockerfile = tmp.Name()
			defer func() {
				os.Remove(dockerfile)
			}()
			fmt.Fprintf(tmp, "FROM --platform=%s %s:%s\n", defaultPlatform, d.Image, tag)
			for _, line := range d.DockerfileExtras {
				_, err := tmp.WriteString(line + "\n")
				if err != nil {
					logrus.Errorf("error writing temporary Dockerfile: %v", err)
					continue tagLoop
				}
			}
		}

		name := fmt.Sprintf("%s:%s", destinationName, destinationTag)
		quayName := fmt.Sprintf("%s/%s", quayURL, name)
		aliyunName := fmt.Sprintf("%s/%s", aliyunURL, name)
		// build Docker image from the Dockerfile
		{
			c, stdout, stderr := command("docker", "build", "-t", quayName, "-f", dockerfile, temporaryWorkingDir)
			if err := c.Run(); err != nil {
				logrus.Errorf("error building custom image for %s:%s: %v\n%s", d.Image, tag, err, stderr.String())
				continue tagLoop
			}
			logrus.Debugf(stdout.String())
		}
		// retag for Aliyun
		{
			c, _, stderr := command("docker", "tag", quayName, aliyunName)
			if err := c.Run(); err != nil {
				logrus.Errorf("error tagging custom image for %s:%s: %v\n%s", d.Image, tag, err, stderr.String())
				continue tagLoop
			}
		}

		wg := sync.WaitGroup{}
		wg.Add(2)
		go pushImage(&wg, quayName)
		go pushImage(&wg, aliyunName)
		wg.Wait()
	}

	return nil
}

func copyImage(wg *sync.WaitGroup, source, destination string) {
	defer wg.Done()
	c, _, stderr := command("skopeo", "copy", "--all", source, destination)
	logrus.Debugf("copying %q to %q", source, destination)
	if err := c.Run(); err != nil {
		logrus.Errorf("error copying %q to %q: %v\n%s", source, destination, err, stderr.String())
	}
	logrus.Debugf("copied %q to %q", source, destination)
}

func pushImage(wg *sync.WaitGroup, nameAndTag string) {
	defer wg.Done()
	c, _, stderr := command("docker", "push", nameAndTag)
	logrus.Debugf("pushing %q", nameAndTag)
	if err := c.Run(); err != nil {
		logrus.Errorf("error pushing %q: %v\n%s", nameAndTag, err, stderr.String())
	}
	logrus.Debugf("pushed %q", nameAndTag)
}

func (r *Retagger) BuildAndTagAll() {
	errors := 0
	for i, job := range r.customDockerfiles {
		logrus.Printf("[%d/%d] Retagging %s", i+1, len(r.customDockerfiles), job.Image)
		if err := job.BuildAndTag(); err != nil {
			logrus.Errorf("got error: %v", err)
			errors++
		}
	}
	if errors > 0 {
		logrus.Fatal("Retagging ended with %d errors", errors)
	}
}

func NewRetagger() (*Retagger, error) {
	r := &Retagger{
		customDockerfiles: []Dockerfile{},
	}
	// load custom dockerfile specs
	b, err := os.ReadFile("customized-dockerfiles.yaml")
	if err != nil {
		return nil, fmt.Errorf("error reading customized-dockerfiles.yaml: %w", err)
	}
	if err := yaml.Unmarshal(b, &r.customDockerfiles); err != nil {
		return nil, fmt.Errorf("error unmarshaling customized-dockerfiles.yaml: %w", err)
	}

	return r, nil
}

func command(name string, args ...string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c := exec.Command(name, args...)
	c.Stdout = stdout
	c.Stderr = stderr
	return c, stdout, stderr
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	if err := os.MkdirAll(temporaryWorkingDir, 0777); err != nil {
		logrus.Fatal(err)
	}
	r, err := NewRetagger()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("Found %d custom Dockerfiles to build", len(r.customDockerfiles))
	r.BuildAndTagAll()
}
