package main

import (
	"flag"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/giantswarm/microerror"
)

var testE2e = flag.Bool("e2e", false, "if set run a e2e test")

const e2eRepository = "retagger-e2e"

func init() {
	flag.Parse()
}

func TestE2e(t *testing.T) {
	if !*testE2e {
		t.Skip("Use -e2e flag to execute this test.")
	}

	c := &RegistryConfig{
		Client: &http.Client{},

		Host:         os.Getenv("REGISTRY"),
		Organisation: os.Getenv("REGISTRY_ORGANISATION"),
		Password:     os.Getenv("REGISTRY_PASSWORD"),
		Username:     os.Getenv("REGISTRY_USERNAME"),
	}
	r, err := NewRegistry(c)
	if err != nil {
		t.Fatal(err)
	}

	defer SetUpE2eTest(t, r)()

	retagger := exec.Command("./retagger -f images-e2e.yaml")
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
	ok, err = r.CheckImageTagExists("retagger-e2e", "3.9")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Image retagger-e2e:3.9 does not exist.")
	}
}

func SetUpE2eTest(t *testing.T, r *Registry) func() {
	if r == nil {
		t.Fatal("registry must not be nil")
	}

	t.Log("setup")

	return func() {
		t.Log("teardown")
		err := cleanupRegistry(t, r)
		if err != nil {
			t.Fatalf("ERROR: Could not clean-up registry: %v", err)
		}
	}
}

func cleanupRegistry(t *testing.T, r *Registry) error {
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

	t.Log("cleaned up")
	return nil
}
