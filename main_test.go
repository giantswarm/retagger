package main

import (
	"path/filepath"
	"reflect"
	"testing"
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
