package retagger

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/giantswarm/micrologger"
)

func Test_filterAndExtract(t *testing.T) {
	tcs := []struct {
		filter         string
		tag            string
		expectedResult string
	}{
		{
			filter:         "(?P<version>.*)",
			tag:            "1.2.3",
			expectedResult: "1.2.3",
		},
		{
			filter:         "(?P<version>.*)-alpine",
			tag:            "1.2.3-alpine",
			expectedResult: "1.2.3",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.tag, func(t *testing.T) {
			m, _ := regexp.Compile(tc.filter)
			result := filterAndExtract(tc.tag, m)
			if result != tc.expectedResult {
				t.Errorf("Expected %s, got %s", tc.expectedResult, result)
			}
		})
	}
}

func Test_findTags(t *testing.T) {

	tags := []string{
		"0.9",
		"0.9-alpine",
		"1.0.0",
		"1.0.0-alpine",
		"1.1.1",
		"1.1.1-alpine",
		"1.2.3",
		"1.2.3-alpine",
		"1.2.3-debian",
	}

	logConfig := micrologger.Config{}
	logger, _ := micrologger.New(logConfig)

	tcs := []struct {
		job            PatternJob
		expectedResult []string
	}{
		{
			job: PatternJob{
				logger:        logger,
				SourcePattern: "> 1.0.0",
				SourceFilter:  "(?P<version>.*)",
				Source: Source{
					Image: "nginx",
				},
				Options:     JobOptions{},
				Destination: Destination{},
			},
			expectedResult: []string{"1.1.1", "1.2.3"},
		},
		{
			job: PatternJob{
				logger:        logger,
				SourcePattern: "> 1.0.0",
				SourceFilter:  "(?P<version>.*)-alpine",
				Source: Source{
					Image: "nginx",
				},
				Options:     JobOptions{},
				Destination: Destination{},
			},
			expectedResult: []string{"1.1.1-alpine", "1.2.3-alpine"},
		},
	}

	for _, tc := range tcs {
		testName := fmt.Sprintf("%s/%s/%s", tc.job.Source.Image, tc.job.SourcePattern, tc.job.SourceFilter)
		t.Run(testName, func(t *testing.T) {
			m, _ := regexp.Compile(tc.job.SourceFilter)
			c, _ := semver.NewConstraint(tc.job.SourcePattern)

			result := tc.job.findTags(tags, c, m)
			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Errorf("Expected %s, got %s", tc.expectedResult, result)
			}
		})
	}
}

func Test_checkConflicts(t *testing.T) {
	tcs := []struct {
		description  string
		hasConflicts bool
		jobs         []SingleJob
	}{
		{
			description:  "same digest, different destination tag suffix",
			hasConflicts: false,
			jobs: []SingleJob{
				{
					Source: Source{
						SHA: "1111111111111111111111111111111111111111111111111111111111111111",
					},
					Destination: Destination{
						Image: "mytool",
						Tag:   "v1.0.0",
					},
				},
				{
					Source: Source{
						SHA: "1111111111111111111111111111111111111111111111111111111111111111",
					},
					Destination: Destination{
						Image: "mytool",
						Tag:   "v1.0.0-mysuffix",
					},
				},
			},
		},

		{
			description:  "different digest, different destination tag",
			hasConflicts: false,
			jobs: []SingleJob{
				{
					Source: Source{
						SHA: "1111111111111111111111111111111111111111111111111111111111111111",
					},
					Destination: Destination{
						Image: "mytool",
						Tag:   "v1.0.0",
					},
				},
				{
					Source: Source{
						SHA: "2222222222222222222222222222222222222222222222222222222222222222",
					},
					Destination: Destination{
						Image: "mytool",
						Tag:   "v1.0.0-0",
					},
				},
			},
		},

		{
			description:  "real use case (positive)",
			hasConflicts: false,
			jobs: []SingleJob{
				{
					Source: Source{
						// coreos image...
						SHA: "1111111111111111111111111111111111111111111111111111111111111111",
					},
					Destination: Destination{
						Image: "quay.io/giantswarm/etcd",
						Tag:   "v3.5.6",
					},
				},
				{
					Source: Source{
						// ...differs from Kubernetes image of etcd
						SHA: "2222222222222222222222222222222222222222222222222222222222222222",
					},
					Destination: Destination{
						Image: "quay.io/giantswarm/etcd",
						Tag:   "3.5.6-0",
					},
				},
				{
					Source: Source{
						SHA: "2222222222222222222222222222222222222222222222222222222222222222",
					},
					Destination: Destination{
						Image: "quay.io/giantswarm/etcd",
						Tag:   "3.5.6-0-k8s",
					},
				},
			},
		},

		{
			description:  "different digest, same destination tag",
			hasConflicts: true,
			jobs: []SingleJob{
				{
					Source: Source{
						SHA: "1111111111111111111111111111111111111111111111111111111111111111",
					},
					Destination: Destination{
						Image: "mytool",
						Tag:   "v1.0.0",
					},
				},
				{
					Source: Source{
						SHA: "2222222222222222222222222222222222222222222222222222222222222222",
					},
					Destination: Destination{
						Image: "mytool",
						Tag:   "v1.0.0",
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		err := checkConflicts(tc.jobs)
		if err != nil && !tc.hasConflicts {
			t.Errorf("Case %s: Expected case to show no conflicts, but got an error: %s", tc.description, err)
		} else if err == nil && tc.hasConflicts {
			t.Errorf("Case %s: Expected case to show conflicts, but got no error", tc.description)
		}
	}
}
