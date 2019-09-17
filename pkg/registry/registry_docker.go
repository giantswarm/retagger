package registry

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/retagger/pkg/images"
)

func (r *Registry) Login() error {
	r.logger.Log("level", "debug", "message", fmt.Sprintf("logging in to %s registry", r.host))

	authConfig := types.AuthConfig{
		Username:      r.username,
		Password:      r.password,
		ServerAddress: r.host,
	}
	res, err := r.docker.RegistryLogin(context.Background(), authConfig)
	if err != nil {
		return microerror.Maskf(err, "could not login to registry")
	} else if res.Status != "Login Succeeded" {
		return microerror.Mask(dockerError)
	}

	r.registryAuth = res.IdentityToken

	return nil
}

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

func (r *Registry) Rebuild(sourceImage, sha, destinationImage, destinationTag string, dockerfileOptions []string) (string, error) {
	retaggedNameWithTag := fmt.Sprintf("%s:%s", destinationImage, destinationTag)

	dockerfile := Dockerfile{
		BaseImage:         sourceImage,
		Sha:               sha,
		DockerfileOptions: dockerfileOptions,
		Tag:               destinationTag,
	}

	f, err := os.Create(fmt.Sprintf("Dockerfile-%s", destinationTag))
	if err != nil {
		return "", microerror.Mask(err)
	}

	// render Dockerfile with overrides
	t := template.Must(template.New("").Parse(customDockerfileTmpl))
	err = t.Execute(f, dockerfile)
	if err != nil {
		return "", microerror.Mask(invalidTemplateError)
	}

	rebuild := exec.Command("docker", "build", "-t", retaggedNameWithTag, "-f", fmt.Sprintf("Dockerfile-%s", destinationTag), ".")
	err = Run(rebuild)
	if err != nil {
		return "", microerror.Mask(err)
	}
	return retaggedNameWithTag, nil
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
