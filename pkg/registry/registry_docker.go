package registry

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/retagger/pkg/images"
)

func (r *Registry) PullImage(image string, sha string) error {
	if sha == "" {
		return microerror.Maskf(invalidArgumentError, "%s SHA should not be empty", image)
	}

	shaName := images.ShaName(image, sha)

	r.logger.Log("level", "debug", "message", fmt.Sprintf("executing: docker pull %s", shaName))

	res, err := r.docker.ImagePull(context.Background(), shaName, types.ImagePullOptions{})
	if err != nil {
		return microerror.Maskf(err, "could not pull image")
	}
	defer res.Close()
	err = r.logDocker(res)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) TagSha(sourceImage, sha, destinationImage, destinationTag string) (string, error) {
	imageSha := images.ShaName(sourceImage, sha)
	retaggedNameWithTag := fmt.Sprintf("%s:%s", destinationImage, destinationTag)

	r.logger.Log("level", "debug", "message", fmt.Sprintf("executing: docker tag %s %s", imageSha, retaggedNameWithTag))

	err := r.docker.ImageTag(context.Background(), imageSha, retaggedNameWithTag)
	if err != nil {
		return "", microerror.Maskf(err, "could not tag image")
	}

	return retaggedNameWithTag, nil
}

func (r *Registry) PushImage(destinationImage, destinationTag string) error {
	imageTag := fmt.Sprintf("%s:%s", destinationImage, destinationTag)

	r.logger.Log("level", "debug", "message", fmt.Sprintf("executing: docker push %s", imageTag))

	opts := types.ImagePushOptions{
		All:          true,
		RegistryAuth: r.getAuthBase64(),
	}
	res, err := r.docker.ImagePush(context.Background(), imageTag, opts)
	if err != nil {
		return microerror.Mask(err)
	}
	defer res.Close()
	err = r.logDocker(res)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) RebuildImage(sourceImage, sha, destinationImage, destinationTag string, dockerfileOptions []string) error {
	//dockerfile := Dockerfile{
	//	BaseImage:         sourceImage,
	//	DockerfileOptions: dockerfileOptions,
	//	Tag:               destinationTag,
	//}

	return nil
}

func (r *Registry) logDocker(reader io.Reader) error {
	s := bufio.NewScanner(reader)

	for s.Scan() {
		logMsg := string(s.Bytes())

		if strings.Contains(logMsg, "error") {
			return microerror.Maskf(dockerError, "docker task failed: %s", logMsg)
		}

		// TODO inline json.
		r.logger.Log("level", "debug", "message", fmt.Sprintf("docker status"), "docker", string(s.Bytes()))
	}
	if err := s.Err(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) getAuthBase64() string {
	auth := fmt.Sprintf(`{"username": "%s", "password": "%s", "serveraddress": "https://%s"}`, r.username, r.password, r.host)

	return base64.StdEncoding.EncodeToString([]byte(auth))
}
