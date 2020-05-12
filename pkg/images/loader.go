package images

import (
	"io/ioutil"

	"github.com/giantswarm/microerror"
	"gopkg.in/yaml.v2"
)

func FromFile(filePath string) (Images, error) {
	var err error

	var yamlFile []byte
	{
		yamlFile, err = ioutil.ReadFile(filePath)
		if err != nil {
			return nil, microerror.Maskf(executionFailedError, "failed to read file %#q with error %#q", filePath, err)
		}
	}

	var images []Image
	{
		err = yaml.Unmarshal(yamlFile, &images)
		if err != nil {
			return nil, microerror.Maskf(executionFailedError, "failed to parse YAML file %#q with error %#q", filePath, err)
		}
	}
	return images, nil
}
