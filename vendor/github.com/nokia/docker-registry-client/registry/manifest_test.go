package registry_test

import (
	"reflect"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	digest "github.com/opencontainers/go-digest"
)

// checkManifest compares the result of the getManifest() call with the expected values of given TestCase
// Does nothing (skips the TestCase) if its media type is not wantMediaType.
// Writes back the results into tc.Expected if --update was given as a command line argument.
func checkManifest(t *testing.T, tc *TestCase, wantMediaType string,
	getManifest func(t *testing.T) (distribution.Manifest, error)) {

	if tc.MediaType != wantMediaType {
		return
	}

	t.Run(tc.Name(), func(t *testing.T) {
		got, err := getManifest(t)
		if err != nil {
			t.Error(err)
			return
		}
		mediaType, payload, err := got.Payload()
		if err != nil {
			t.Error("Payload() error:", err)
			return
		}
		d := digest.FromBytes(payload)

		if !*_testDataUpdate {
			// do actual testing of manifest
			if mediaType == "" {
				mediaType = schema1.MediaTypeSignedManifest
			}
			if mediaType != tc.MediaType {
				t.Errorf("MediaType = %v, want %v", mediaType, tc.MediaType)
			}
			if d != tc.ManifestDigest {
				t.Errorf("digest = %s, want %s", d, tc.ManifestDigest)
			}
			if !blobSlicesAreEqual(got.References(), tc.Blobs) {
				t.Errorf("\nblobs:\n%v,\nwant:\n%v", got.References(), tc.Blobs)
			}
		} else {
			// update TestCase to reflect the result of the tested method
			tc.ManifestDigest = d
			tc.Blobs = got.References()
		}
	})
}

func TestRegistry_Manifest(t *testing.T) {
	for _, tc := range testCases(t) {
		checkManifest(t, tc, schema1.MediaTypeSignedManifest, func(t *testing.T) (distribution.Manifest, error) {
			return tc.Registry(t).Manifest(tc.Repository, tc.Reference)
		})
		checkManifest(t, tc, schema2.MediaTypeManifest, func(t *testing.T) (distribution.Manifest, error) {
			return tc.Registry(t).Manifest(tc.Repository, tc.Reference)
		})
	}
	// updateTestData skipped deliberately
}

func TestRegistry_ManifestV1(t *testing.T) {
	for _, tc := range testCases(t) {
		checkManifest(t, tc, schema1.MediaTypeSignedManifest, func(t *testing.T) (distribution.Manifest, error) {
			return tc.Registry(t).ManifestV1(tc.Repository, tc.Reference)
		})
	}
	updateTestData(t)
}

func TestRegistry_ManifestV2(t *testing.T) {
	for _, tc := range testCases(t) {
		checkManifest(t, tc, schema2.MediaTypeManifest, func(t *testing.T) (distribution.Manifest, error) {
			return tc.Registry(t).ManifestV2(tc.Repository, tc.Reference)
		})
	}
	updateTestData(t)
}

func TestRegistry_ManifestList(t *testing.T) {
	for _, tc := range testCases(t) {
		checkManifest(t, tc, manifestlist.MediaTypeManifestList, func(t *testing.T) (distribution.Manifest, error) {
			return tc.Registry(t).ManifestList(tc.Repository, tc.Reference)
		})
	}
	updateTestData(t)
}

// checkManifestDescriptor compares the result of the getManifestDescriptor() call with the expected values of the given TestCase (tc).
// Writes back the results into tc.Expected if --update was given as a command line argument.
func checkManifestDescriptor(t *testing.T, tc *TestCase,
	getManifestDescriptor func(t *testing.T) (distribution.Descriptor, error)) {

	if tc.Writeable {
		return
	}

	t.Run(tc.Name(), func(t *testing.T) {
		got, err := getManifestDescriptor(t)
		if err != nil {
			t.Error(err)
			return
		}
		if !*_testDataUpdate {

			if !reflect.DeepEqual(got, tc.ManifestDescriptor) {
				t.Errorf("mainfest descriptor = %v, want %v", got, tc.ManifestDescriptor)
			}
		} else {
			// update TestCase to reflect the result of the tested method
			tc.ManifestDescriptor = got
		}
	})
}

func TestRegistry_ManifestDescriptor(t *testing.T) {
	for _, tc := range testCases(t) {
		checkManifestDescriptor(t, tc, func(t *testing.T) (distribution.Descriptor, error) {
			return tc.Registry(t).ManifestDescriptor(tc.Repository, tc.Reference)
		})
	}
	updateTestData(t)
}

// checkManifestDigest compares the result of the getManifestDigest() call with the expected values of the given TestCase (tc).
// Does nothing (skips the TestCase) if its media type is not wantMediaType or --update was given as a command line argument..
func checkManifestDigest(t *testing.T, tc *TestCase, wantMediaType string,
	getManifestDigest func(t *testing.T) (digest.Digest, error)) {

	if *_testDataUpdate || tc.MediaType != wantMediaType {
		return
	}

	t.Run(tc.Name(), func(t *testing.T) {
		got, err := getManifestDigest(t)
		if err != nil {
			t.Error(err)
			return
		}
		if got != tc.ManifestDescriptor.Digest {
			t.Errorf("digest = %s, want %s", got, tc.ManifestDigest)
		}
	})
}

func TestRegistry_ManifestDigest(t *testing.T) {
	for _, tc := range testCases(t) {
		checkManifestDigest(t, tc, schema1.MediaTypeSignedManifest, func(t *testing.T) (digest.Digest, error) {
			return tc.Registry(t).ManifestDigest(tc.Repository, tc.Reference)
		})
	}
}

func TestRegistry_ManifestV2Digest(t *testing.T) {
	for _, tc := range testCases(t) {
		checkManifestDigest(t, tc, schema2.MediaTypeManifest, func(t *testing.T) (digest.Digest, error) {
			return tc.Registry(t).ManifestV2Digest(tc.Repository, tc.Reference)
		})
	}
}
