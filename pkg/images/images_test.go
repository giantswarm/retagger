package images

import (
	"sort"
	"testing"
)

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

func TestImageName(t *testing.T) {
	organisation := "giantswarm"

	tests := []struct {
		image        Image
		expectedName string
	}{
		{
			image: Image{
				Name: "quay.io/coreos/hyperkube",
			},
			expectedName: "giantswarm/hyperkube",
		},
		{
			image: Image{
				Name: "prom/prometheus",
			},
			expectedName: "giantswarm/prometheus",
		},
		{
			image: Image{
				Name: "golang",
			},
			expectedName: "giantswarm/golang",
		},
	}

	for _, test := range tests {
		returnedName := ImageName(organisation, test.image.Name)
		if returnedName != test.expectedName {
			t.Fatalf("'%v' != '%v'", returnedName, test.expectedName)
		}
	}
}

func TestRetaggedName(t *testing.T) {
	registry := "quay.io"
	organisation := "giantswarm"

	tests := []struct {
		image        Image
		expectedName string
	}{
		{
			image: Image{
				Name: "quay.io/coreos/hyperkube",
			},
			expectedName: "quay.io/giantswarm/hyperkube",
		},
		{
			image: Image{
				Name: "prom/prometheus",
			},
			expectedName: "quay.io/giantswarm/prometheus",
		},
		{
			image: Image{
				Name: "golang",
			},
			expectedName: "quay.io/giantswarm/golang",
		},
	}

	for _, test := range tests {
		returnedName := RetaggedName(registry, organisation, test.image.Name)
		if returnedName != test.expectedName {
			t.Fatalf("'%v' != '%v'", returnedName, test.expectedName)
		}
	}
}

func TestShaName(t *testing.T) {
	tests := []struct {
		imageName    string
		sha          string
		expectedName string
	}{
		{
			imageName:    "quay.io/coreos/hyperkube",
			sha:          "5ff22b5c65d5b93aa948b79028dc136a22cda2f049283103f10bd45650b47312",
			expectedName: "quay.io/coreos/hyperkube@sha256:5ff22b5c65d5b93aa948b79028dc136a22cda2f049283103f10bd45650b47312",
		},
	}

	for _, test := range tests {
		returnedName := ShaName(test.imageName, test.sha)
		if returnedName != test.expectedName {
			t.Fatalf("'%v' != '%v'", returnedName, test.expectedName)
		}
	}
}

func TestImagesOrderedAlphabetically(t *testing.T) {
	conf, err := FromFile("../../images.yaml")
	if err != nil {
		t.Fatal(err)
	}

	var imageNames []string
	for _, image := range conf.Images {
		imageNames = append(imageNames, image.Name)
	}

	if !sort.StringsAreSorted(imageNames) {
		t.Logf("images are not sorted alphabetically")
		t.Logf("have \n%v", imageNames)
		sort.Strings(imageNames)
		t.Logf("want \n%v", imageNames)
		t.Fail()
	}
}
