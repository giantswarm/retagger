package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMissingCopies(t *testing.T) {
	dir := t.TempDir()
	filtered := filepath.Join(dir, "images.yaml.filtered")
	copied := filepath.Join(dir, "images.yaml.filtered.copied")
	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filtered, `docker.io:
  images:
    giantswarm/bats:
      - "0.4.0"
    giantswarm/localstack:
      - v2.2.0-gsdev2
    library/alpine:
      - "3.18"
      - "3.19"
`)

	// The localstack and alpine:3.18 copies landed; bats and alpine:3.19 did not.
	write(copied, "sha256:aaa docker://gsoci.azurecr.io/giantswarm/localstack:v2.2.0-gsdev2\n"+
		"\n"+
		"sha256:bbb docker://gsoci.azurecr.io/giantswarm/alpine:3.18\n")
	got, err := missingCopies(filtered, copied)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"alpine:3.19", "bats:0.4.0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("missing = %v, want %v", got, want)
	}

	// An empty record, the one the job creates up front, misses everything.
	write(copied, "")
	got, err = missingCopies(filtered, copied)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Errorf("empty record: missing = %v, want all 4 tags", got)
	}

	// Everything recorded: nothing missing.
	write(copied, "sha256:1 docker://r/giantswarm/bats:0.4.0\nsha256:2 docker://r/giantswarm/localstack:v2.2.0-gsdev2\n"+
		"sha256:3 docker://r/giantswarm/alpine:3.18\nsha256:4 docker://r/giantswarm/alpine:3.19\n")
	got, err = missingCopies(filtered, copied)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("missing = %v, want none", got)
	}
}

func TestMissingCopiesErrorsOnUnreadableInput(t *testing.T) {
	if _, err := missingCopies("/nonexistent/filtered", "/nonexistent/copied"); err == nil {
		t.Error("expected an error for a missing images file")
	}
}

func TestMissingCopiesSkipsDigestEntries(t *testing.T) {
	dir := t.TempDir()
	filtered := filepath.Join(dir, "f.yaml.filtered")
	copied := filepath.Join(dir, "f.yaml.filtered.copied")
	if err := os.WriteFile(filtered, []byte("docker.io:\n  images:\n    a/foo:\n      - sha256:abc\n      - \"1\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(copied, []byte("docker://r/giantswarm/foo:1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := missingCopies(filtered, copied)
	if err != nil || len(got) != 0 {
		t.Errorf("got %v, %v; want nothing missing", got, err)
	}
}
