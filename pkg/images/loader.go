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
			return nil, microerror.Mask(err)
		}
	}

	var images []Image
	{
		err = yaml.Unmarshal(yamlFile, &images)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	return images, nil
}
