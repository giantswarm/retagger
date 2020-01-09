package registry_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/docker/distribution"
	"github.com/nokia/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"
)

// Expected stores the expected results of various tests
type Expected struct {
	MediaType          string                    `json:"media_type,omitempty"`
	ManifestDigest     digest.Digest             `json:"manifest_digest,omitempty"`
	Blobs              []distribution.Descriptor `json:"blobs,omitempty"`
	ManifestDescriptor distribution.Descriptor   `json:"manifest_descriptor,omitempty"`
}

// TestCase represents a test case (normally read from a test data file).
type TestCase struct {
	Url        string `json:"url"`
	Repository string `json:"repository"`
	Reference  string `json:"reference"`
	Writeable  bool   `json:"writeable,omitempty"`

	registry.Options
	registry *registry.Registry

	Expected `json:"expected"`
	Origin   string `json:"-"` // name of the test data file that this TestCase was read from
}

func (tc TestCase) Name() string {
	return fmt.Sprintf("%s/%s@%s,%v", tc.Url, tc.Repository, tc.Reference, tc.Writeable)
}

func (tc *TestCase) Registry(t *testing.T) *registry.Registry {
	if tc.registry == nil {
		var err error
		tc.registry, err = registry.NewCustom(tc.Url, tc.Options)
		if err != nil {
			t.Fatal("failed to create registry client:", err)
		}
	}
	return tc.registry
}

const testDataFilePattern = "testdata/*registry_tests.json"

var _testDataUpdate = flag.Bool("update", false, "update testdata files")
var _testCases []*TestCase

// testCases loads all test data files and returns with the union of all testcases read
func testCases(t *testing.T) []*TestCase {
	// cache results in _testCases
	if _testCases != nil {
		return _testCases
	}

	// load testcases from files matching testDataFilePattern
	tdFilenames, err := filepath.Glob(testDataFilePattern)
	if err != nil {
		t.Fatal("failed to list test data files:", testDataFilePattern)
	}
	for _, tdFilename := range tdFilenames {
		tdFile, err := os.Open(tdFilename)
		if err != nil {
			t.Fatal("failed to open test data file:", tdFilename)
		}

		var tcs []*TestCase
		err = json.NewDecoder(tdFile).Decode(&tcs)
		if err != nil {
			t.Fatalf("failed to load test data from %s: %s", tdFilename, err)
		}
		for i := range tcs {
			tcs[i].Origin = tdFilename
		}
		_testCases = append(_testCases, tcs...)
	}

	// add testcase for DockerHub if credentials are given in environment variables
	username := os.Getenv("DRC_TEST_DOCKERHUB_USERNAME")
	password := os.Getenv("DRC_TEST_DOCKERHUB_PASSWORD")
	if username != "" && password != "" {
		tc := &TestCase{
			Url:        "https://registry-1.docker.io",
			Repository: username + "/docker-registry-client-test",
			Reference:  "latest",
			Writeable:  true,
			Options: registry.Options{
				Username: username,
				Password: password,
			},
		}
		_testCases = append(_testCases, tc)
	}
	return _testCases
}

// updateTestData writes the actual results back to the test data files
// if the --update flag was given to the test
func updateTestData(t *testing.T) {
	if !*_testDataUpdate {
		return
	}

	tdFiles := make(map[string][]*TestCase)
	for _, tc := range testCases(t) {
		if tc.Origin != "" {
			tdFiles[tc.Origin] = append(tdFiles[tc.Origin], tc)
		}
	}
	for tdFilename, tcs := range tdFiles {
		t.Log("Updating", tdFilename)
		tdFile, err := os.Create(tdFilename)
		if err != nil {
			t.Fatal(err)
		}
		enc := json.NewEncoder(tdFile)
		enc.SetIndent("", "  ")
		err = enc.Encode(tcs)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func foreachWritableTestcase(t *testing.T, testFunction func(*testing.T, *TestCase)) {
	for _, tc := range testCases(t) {
		if tc.Writeable {
			t.Run(tc.Name(), func(t *testing.T) {
				testFunction(t, tc)
			})
		}
	}

}

// blobSlicesAreEqual checks if the two given slices are equal
// WARNING: this will modify (i.e. sort) both a and b
func blobSlicesAreEqual(a, b []distribution.Descriptor) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Slice(a, func(i, j int) bool { return a[i].Digest.String() < a[j].Digest.String() })
	sort.Slice(b, func(i, j int) bool { return b[i].Digest.String() < b[j].Digest.String() })
	return reflect.DeepEqual(a, b)
}
