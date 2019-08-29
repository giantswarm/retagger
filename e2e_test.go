package main

import (
	"flag"
	"net/http"
	"os"
	"testing"
)

var testE2e = flag.Bool("e2e", false, "if set run a e2e test")

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
	_, err := NewRegistry(c)
	if err != nil {
		t.Fatal(err)
	}

	defer SetUpE2eTest(t)()
}

func SetUpE2eTest(t *testing.T) func() {
	t.Log("setup")

	return func() {
		t.Log("teardown")
		cleanupRegistry(t)
	}
}

func cleanupRegistry(t *testing.T) {
	t.Log("cleaned up")
}
