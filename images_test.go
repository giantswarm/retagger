package main

import "testing"

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
		returnedName := ImageName(organisation, test.image)
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
		returnedName := RetaggedName(registry, organisation, test.image)
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
