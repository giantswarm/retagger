package main

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

const tagLatest = "latest"

// TestFilterTags pins the tag selection of the entries in images/: a release-tag
// filter with a prerelease floor keeps the -gs.N line and nothing else, a plain
// semver floor keeps releases and drops release candidates and non-semver tags.
func TestFilterTags(t *testing.T) {
	tests := []struct {
		name string
		img  RenamedImage
		tags []string
		want []string
	}{
		{
			name: "release tags of a -gs.N line from a prerelease floor",
			img: RenamedImage{
				Filter: `^([0-9]+\.[0-9]+\.[0-9]+-gs\.[0-9]+)$`,
				Semver: ">= 0.11.0-gs.16",
			},
			tags: []string{
				"0.11.0-gs.15",
				"0.11.0-gs.16",
				"0.11.0-gs.20",
				"0.11.1-gs.1",
				"0.12.0-gs.3",
				"0.11.0-dev.poc.2026-09-10.21-15-57.hcc77fbb",
				"0.11.1-dev.sync-20260914-main.h7b3d881",
				"0.11.0",
				"sha256-fc177422d843219b017cce398443b2458337ca5084abecfc0b7e4a622d7e6843.sbom",
				tagLatest,
			},
			want: []string{"0.11.0-gs.16", "0.11.0-gs.20", "0.11.1-gs.1", "0.12.0-gs.3"},
		},
		{
			name: "release tags of a -gs.N line, floor at a patch release",
			img: RenamedImage{
				Filter: `^([0-9]+\.[0-9]+\.[0-9]+-gs\.[0-9]+)$`,
				Semver: ">= 0.0.30-gs.4",
			},
			tags: []string{"0.0.27-gs.9", "0.0.30-gs.3", "0.0.30-gs.4", "0.0.31-gs.1"},
			want: []string{"0.0.30-gs.4", "0.0.31-gs.1"},
		},
		{
			name: "v-prefixed releases from a floor, no release candidates or non-semver tags",
			img:  RenamedImage{Semver: ">= v1.9.0"},
			tags: []string{"v1.8.5", "v1.9.0", "v1.9.5", "v1.10.0-rc.1", "v2.0.0", "v2beta2-dd92f466", tagLatest},
			want: []string{"v1.9.0", "v1.9.5", "v2.0.0"},
		},
		{
			name: "filter extracts the version to compare",
			img: RenamedImage{
				Filter: `^(\d+\.\d+\.\d+)-kafka-.*`,
				Semver: ">= 0.51.0",
			},
			tags: []string{"0.50.0-kafka-3.9.0", "0.51.0-kafka-4.1.0", "0.51.0", tagLatest},
			want: []string{"0.51.0-kafka-4.1.0"},
		},
		{
			name: "tag pattern",
			img:  RenamedImage{TagOrPattern: `^[0-9]+\.[0-9]+\.[0-9]+-full$`},
			tags: []string{"0.10.0", "0.10.0-full", "0.10.0-full-debug"},
			want: []string{"0.10.0-full"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.img.Validate(); err != nil {
				t.Fatalf("Validate: %v", err)
			}
			got, err := tt.img.FilterTags(tt.tags)
			if err != nil {
				t.Fatalf("FilterTags: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterTags = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRenamedImagesFiles parses every renamed-images file the way `retagger run`
// does and validates each entry, its tag pattern, semver constraint and filter
// included, so a broken entry fails here and not in the nightly run.
func TestRenamedImagesFiles(t *testing.T) {
	files, err := filepath.Glob("images/renamed-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no images/renamed-*.yaml files found")
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			images, err := loadRenamedImages(file)
			if err != nil {
				t.Fatal(err)
			}
			if len(images) == 0 {
				t.Fatal("no entries")
			}
			for _, img := range images {
				if img.Image == "" {
					t.Errorf("entry without image: %+v", img)
					continue
				}
				if err := img.Validate(); err != nil {
					t.Errorf("%s: %v", img.Image, err)
					continue
				}
				if _, err := img.FilterTags(nil); err != nil {
					t.Errorf("%s: %v", img.Image, err)
				}
			}
		})
	}
}

// recordingSigner records the references it was asked to sign.
type recordingSigner struct {
	mu    sync.Mutex
	calls []string
	seen  chan string
}

func (r *recordingSigner) Sign(image string) signOutcome {
	r.mu.Lock()
	r.calls = append(r.calls, image)
	r.mu.Unlock()
	r.seen <- image
	return outcomeSigned
}

// TestCopyTagSignsAzureBeforeAliyunReturns pins the ordering copyTag has to
// keep: the AzureCR copy is signed as soon as it lands, while the Aliyun copy is
// still running (a stalled Aliyun push must never hold back the signature), the
// Aliyun copy runs under aliyunCopyTimeout and the AzureCR copy unbounded, and a
// failed AzureCR copy is not signed.
func TestCopyTagSignsAzureBeforeAliyunReturns(t *testing.T) {
	defer func(orig func(string, string, time.Duration) error) { copyFn = orig }(copyFn)

	release := make(chan struct{})
	var mu sync.Mutex
	timeouts := map[string]time.Duration{}
	copyFn = func(source, destination string, timeout time.Duration) error {
		mu.Lock()
		timeouts[destination] = timeout
		mu.Unlock()
		if strings.Contains(destination, aliyunURL) {
			<-release // the Aliyun push stalls until the test lets it go
			return errors.New("stalled")
		}
		return nil
	}
	img := &RenamedImage{Image: "ghcr.io/example/app", OverrideRepoName: "example/app"}
	signer := &recordingSigner{seen: make(chan string, 1)}

	done := make(chan struct{})
	go func() {
		img.copyTag("docker://ghcr.io/example/app:1.2.3", "1.2.3", signer)
		close(done)
	}()

	want := azureURL + "/example/app:1.2.3"
	select {
	case got := <-signer.seen:
		if got != want {
			t.Fatalf("signed %q, want %q", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the AzureCR copy was not signed while the Aliyun copy was still running")
	}
	select {
	case <-done:
		t.Fatal("copyTag returned before the Aliyun copy did")
	default:
	}
	close(release)
	<-done

	mu.Lock()
	defer mu.Unlock()
	if got := timeouts[dockerTransport+want]; got != 0 {
		t.Errorf("AzureCR copy timeout = %v, want unbounded", got)
	}
	if got := timeouts[dockerTransport+aliyunURL+"/example/app:1.2.3"]; got != aliyunCopyTimeout {
		t.Errorf("Aliyun copy timeout = %v, want %v", got, aliyunCopyTimeout)
	}
	if len(signer.calls) != 1 {
		t.Errorf("signed %d times, want once: %v", len(signer.calls), signer.calls)
	}
}

// TestCopyTagDoesNotSignAFailedAzureCopy: a copy that did not land is not signed.
func TestCopyTagDoesNotSignAFailedAzureCopy(t *testing.T) {
	defer func(orig func(string, string, time.Duration) error) { copyFn = orig }(copyFn)
	copyFn = func(source, destination string, timeout time.Duration) error { return errors.New("push failed") }
	img := &RenamedImage{Image: "ghcr.io/example/app", OverrideRepoName: "example/app"}
	signer := &recordingSigner{seen: make(chan string, 1)}
	img.copyTag("docker://ghcr.io/example/app:1.2.3", "1.2.3", signer)
	if len(signer.calls) != 0 {
		t.Errorf("signed a failed copy: %v", signer.calls)
	}
}
