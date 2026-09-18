package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The rule and the tags of the alpine/k8s mirror, as in images/renamed-images.yaml.
const (
	testImage      = "alpine/k8s"
	testOverride   = "alpine-k8s"
	testConstraint = ">= 1.37.0"
	testTag        = "1.37.0"
	testPrefixed   = "v1.2.3"
	testRegistry   = "registry.example/ns"
)

func TestRenamedImageDestination(t *testing.T) {
	for name, tc := range map[string]struct {
		image    RenamedImage
		tag      string
		wantName string
		wantTag  string
	}{
		"basename": {
			image:    RenamedImage{Image: testImage, Semver: testConstraint},
			tag:      testTag,
			wantName: "k8s", wantTag: testTag,
		},
		"override": {
			image:    RenamedImage{Image: testImage, OverrideRepoName: testOverride, Semver: testConstraint},
			tag:      testTag,
			wantName: testOverride, wantTag: testTag,
		},
		"suffix and stripped prefix": {
			image:    RenamedImage{Image: "ghcr.io/example/img", Semver: ">= 1.0.0", StripSemverPrefix: true, AddTagSuffix: "giantswarm"},
			tag:      testPrefixed,
			wantName: "img", wantTag: "1.2.3-giantswarm",
		},
		"prefix kept without a semver rule": {
			image:    RenamedImage{Image: "example/img", TagOrPattern: "v.*", StripSemverPrefix: true},
			tag:      testPrefixed,
			wantName: "img", wantTag: testPrefixed,
		},
	} {
		if got := tc.image.DestinationName(); got != tc.wantName {
			t.Errorf("%s: DestinationName() = %q, want %q", name, got, tc.wantName)
		}
		if got := tc.image.DestinationTag(tc.tag); got != tc.wantTag {
			t.Errorf("%s: DestinationTag(%q) = %q, want %q", name, tc.tag, got, tc.wantTag)
		}
	}
}

func TestFindMissingTagsUsesTheDestinationTag(t *testing.T) {
	img := RenamedImage{Image: "example/img", Semver: ">= 1.0.0", StripSemverPrefix: true, AddTagSuffix: "gs"}
	got := img.FindMissingTags([]string{"v1.0.0", "v1.1.0"}, []string{"1.0.0-gs"}, []string{"1.0.0-gs", "1.1.0-gs"})
	if want := []string{"v1.1.0"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("missing = %v, want %v", got, want)
	}
}

// fakeSkopeoListTags puts a `skopeo` script on PATH whose `list-tags` answers
// with tags for every repository.
func fakeSkopeoListTags(t *testing.T, tags ...string) {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\ncase \"$*\" in *list-tags*) printf 'Tags:\\n'; for tag in $TAGS; do printf -- '- %s\\n' \"$tag\"; done;; *) exit 1;; esac\n"
	if err := os.WriteFile(filepath.Join(dir, "skopeo"), []byte(script), 0o700); err != nil { // #nosec G306 -- an executable test fixture
		t.Fatal(err)
	}
	t.Setenv("TAGS", joinSpace(tags))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func joinSpace(items []string) string {
	s := ""
	for i, item := range items {
		if i > 0 {
			s += " "
		}
		s += item
	}
	return s
}

func TestDestinationReferences(t *testing.T) {
	fakeSkopeoListTags(t, "1.36.0", testTag, "1.37.1", "latest")

	semver := RenamedImage{Image: testImage, OverrideRepoName: testOverride, Semver: testConstraint}
	got, err := semver.DestinationReferences(testRegistry)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{testRegistry + "/alpine-k8s:1.37.0", testRegistry + "/alpine-k8s:1.37.1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("semver rule: references = %v, want %v", got, want)
	}

	sha := RenamedImage{Image: "amazon/aws-cli", OverrideRepoName: "awscli", TagOrPattern: "2.7.35", SHA: "e5988c45"}
	got, err = sha.DestinationReferences(testRegistry)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{testRegistry + "/awscli:2.7.35"}; !reflect.DeepEqual(got, want) {
		t.Errorf("sha rule: references = %v, want %v (no registry listing for a pinned digest)", got, want)
	}
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestIsRenamedImagesFile(t *testing.T) {
	renamed := writeFile(t, "renamed-images.yaml", "- image: alpine/k8s\n  override_repo_name: alpine-k8s\n  semver: \">= 1.37.0\"\n")
	skopeo := writeFile(t, "skopeo-docker-io.yaml", "docker.io:\n  images:\n    kserve/agent:\n      - v0.20.0\n")
	scalar := writeFile(t, "scalar.yaml", "just a string\n")

	if got, err := isRenamedImagesFile(renamed); err != nil || !got {
		t.Errorf("renamed-images file: got %v, %v", got, err)
	}
	if got, err := isRenamedImagesFile(skopeo); err != nil || got {
		t.Errorf("skopeo file: got %v, %v", got, err)
	}
	if _, err := isRenamedImagesFile(scalar); err == nil {
		t.Error("a scalar document is neither kind of file")
	}
}

func TestGovernedReferencesShardsRenamedImagesByExecutor(t *testing.T) {
	fakeSkopeoListTags(t, "1.0.0", "2.0.0")
	file := writeFile(t, "renamed-images.yaml", ""+
		"- image: example/first\n  semver: \">= 2.0.0\"\n"+
		"- image: example/second\n  override_repo_name: second-renamed\n  semver: \">= 1.0.0\"\n"+
		"- image: example/third\n  tag_or_pattern: \"^2\"\n")
	flagRegistry = testRegistry
	flagExecutorCount, flagExecutorID = 2, 0
	t.Cleanup(func() { flagRegistry, flagExecutorCount, flagExecutorID = azureURL, 1, 0 })

	got, err := governedReferences(file)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{testRegistry + "/first:2.0.0", testRegistry + "/third:2.0.0"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("executor 0 of 2: references = %v, want %v", got, want)
	}

	flagExecutorID = 1
	got, err = governedReferences(file)
	if err != nil {
		t.Fatal(err)
	}
	want = []string{testRegistry + "/second-renamed:1.0.0", testRegistry + "/second-renamed:2.0.0"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("executor 1 of 2: references = %v, want %v", got, want)
	}
}

func TestGovernedReferencesReportsInvalidRulesAndKeepsTheRest(t *testing.T) {
	fakeSkopeoListTags(t, "1.0.0")
	file := writeFile(t, "renamed-images.yaml", ""+
		"- image: example/valid\n  semver: \">= 1.0.0\"\n"+
		"- image: example/invalid\n")
	flagRegistry = testRegistry
	t.Cleanup(func() { flagRegistry = azureURL })

	got, err := governedReferences(file)
	if err == nil {
		t.Fatal("a rule without a tag selector must be reported")
	}
	if want := []string{testRegistry + "/valid:1.0.0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("references = %v, want %v", got, want)
	}
	if n := len(multiErrors(err)); n != 1 {
		t.Errorf("multiErrors = %d, want 1", n)
	}
}
