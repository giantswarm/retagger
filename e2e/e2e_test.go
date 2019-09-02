// +build e2e

package e2e

import (
	"os"
	"os/exec"
	"testing"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/retagger/pkg/registry"
)

const e2eRepository = "retagger-e2e"

func TestE2e(t *testing.T) {
	c := registry.Config{
		Host:         os.Getenv("REGISTRY"),
		Organisation: os.Getenv("REGISTRY_ORGANISATION"),
		Password:     os.Getenv("REGISTRY_PASSWORD"),
		Username:     os.Getenv("REGISTRY_USERNAME"),
		LogFunc: func(f string, args ...interface{}) {
			t.Logf(f, args...)
		},
	}
	r, err := registry.New(c)
	if err != nil {
		t.Fatal(err)
	}
	defer SetUpE2eTest(t, r)()

	retagger := exec.Command(
		"./retagger",
		"-f", "./images-e2e.yaml",
		"-r", c.Host,
		"-o", c.Organisation,
		"-u", c.Username,
		"-p", c.Password)
	err = Run(retagger)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := r.CheckImageTagExists("retagger-e2e", "3.7")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Image retagger-e2e:3.7 does not exist.")
	}
}

func SetUpE2eTest(t *testing.T, r *registry.Registry) func() {
	if r == nil {
		t.Fatal("registry must not be nil")
	}

	return func() {
		t.Log("teardown")
		err := cleanupRegistry(t, r)
		if err != nil {
			t.Fatalf("ERROR: Could not clean-up registry: %v", err)
		}
		t.Log("teardown completed successfully")
	}
}

func cleanupRegistry(t *testing.T, r *registry.Registry) error {
	tags, err := r.ListImageTags(e2eRepository)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, tag := range tags {
		err = r.DeleteImage(e2eRepository, tag)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func Run(c *exec.Cmd) error {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
