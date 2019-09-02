package config

import (
	"io/ioutil"

	"github.com/giantswarm/microerror"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Images []Image
}

func FromFile(filePath string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, microerror.Maskf(err, "could not read file %s: #%v ", filePath)
	}

	var images []Image
	err = yaml.Unmarshal(yamlFile, &images)
	if err != nil {
		return nil, microerror.Maskf(err, "could not parse YAML file %s: %v", filePath)
	}

	c := &Config{
		Images: images,
	}
	return c, nil
}
