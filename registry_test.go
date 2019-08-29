package main

import "testing"

func Test_RetaggedName(t *testing.T) {
	tcs := []struct {
		host           string
		organization   string
		image          string
		expectedResult string
	}{
		{
			host:           "quay.io",
			organization:   "acme",
			image:          "gcr.io/foo/bar",
			expectedResult: "quay.io/acme/bar",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.image, func(t *testing.T) {
			result := RetaggedName(tc.host, tc.organization, tc.image)
			if result != tc.expectedResult {
				t.Errorf("Expected %s, got %s", tc.expectedResult, result)
			}
		})
	}
}
