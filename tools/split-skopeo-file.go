package main

import (
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// Source:
// https://github.com/kubasobon/skopeo/blob/semver/cmd/skopeo/sync.go#L72-L79
type skopeoStruct struct {
	Images           map[string][]string `yaml:"images"`
	ImagesByTagRegex map[string]string   `yaml:"images-by-tag-regex"`
	ImagesBySemver   map[string][]string `yaml:"images-by-semver"`
}

var (
	flagParts uint
	flagSrc   string
	flagDst   string
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	pflag.UintVar(&flagParts, "parts", 1, "Number of slices to produce out of source file")
	pflag.StringVar(&flagSrc, "src", "", "skopeo sync file to split")
	pflag.StringVar(&flagDst, "dest", "", "Directory to write file parts to")
	pflag.Parse()

	if flagParts == 0 {
		logrus.Fatal("Number of parts has to be positive")
	}
	if flagParts > 10 {
		logrus.Warnf("Given %d parts, are you certain?", flagParts)
	}
	if flagSrc == "" {
		logrus.Fatalf("%q flag has to be set", "src")
	}
	if flagDst == "" {
		logrus.Fatalf("%q flag has to be set", "dest")
	}
}

func main() {
	b, err := os.ReadFile(flagSrc)
	if err != nil {
		logrus.Fatalf("error reading %q: %s", flagSrc, err)
	}

	var registryMap map[string]skopeoStruct
	if err := yaml.Unmarshal(b, &registryMap); err != nil {
		logrus.Fatalf("error reading %q: %s", flagSrc, err)
	}
	logrus.Infof("Splitting %q into %d parts", flagSrc, flagParts)

	parts := make([]map[string]skopeoStruct, flagParts)
	for i := uint(0); i < flagParts; i++ {
		parts[i] = map[string]skopeoStruct{}
	}

	for registryName, skopeo := range registryMap {
		for i := uint(0); i < flagParts; i++ {
			parts[i][registryName] = skopeoStruct{
				Images:           map[string][]string{},
				ImagesByTagRegex: map[string]string{},
				ImagesBySemver:   map[string][]string{},
			}
		}

		i := 0
		for k := range skopeo.Images {
			index := i % int(flagParts)
			parts[index][registryName].Images[k] = skopeo.Images[k]
			i++
		}

		i = 0
		for k := range skopeo.ImagesByTagRegex {
			index := i % int(flagParts)
			parts[index][registryName].ImagesByTagRegex[k] = skopeo.ImagesByTagRegex[k]
			i++
		}

		i = 0
		for k := range skopeo.ImagesBySemver {
			index := i % int(flagParts)
			parts[index][registryName].ImagesBySemver[k] = skopeo.ImagesBySemver[k]
			i++
		}
	}

	if err := os.MkdirAll(flagDst, 0777); err != nil {
		logrus.Fatalf("error creating directory %q: %s", flagDst, err)
	}

	for i, part := range parts {
		data, err := yaml.Marshal(part)
		if err != nil {
			logrus.Fatalf("error marshaling part %d: %s", i+1, err)
		}
		err = os.WriteFile(
			path.Join(flagDst, fmt.Sprintf("part-%d.yaml", i)),
			data, 0644,
		)
		if err != nil {
			logrus.Fatalf("error writing part %d: %s", i+1, err)
		}
		logrus.Printf("Wrote part %d", i+1)
	}

}
