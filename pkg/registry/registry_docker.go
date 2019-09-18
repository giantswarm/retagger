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

	// We don't care for the AuthenticateOKBody.IdentityToken, as it's always nil.
	err := IsDockerLoginFailed(r.docker.RegistryLogin(context.Background(), authConfig))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) PullImage(image string, sha string) error {
	if image == "" {
		return microerror.Maskf(invalidArgumentError, "image should not be empty")
	}
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
	if sourceImage == "" {
		return "", microerror.Maskf(invalidArgumentError, "sourceImage should not be empty")
	}
	if sha == "" {
		return "", microerror.Maskf(invalidArgumentError, "%s SHA should not be empty", sourceImage)
	}
	if destinationImage == "" {
		return "", microerror.Maskf(invalidArgumentError, "destinationImage should not be empty")
	}
	if destinationTag == "" {
		return "", microerror.Maskf(invalidArgumentError, "destinationTag should not be empty")
	}

	imageSha := images.ShaName(sourceImage, sha)
	retaggedNameWithTag := images.ImageWithTag(destinationImage, destinationTag)

	r.logger.Log("level", "debug", "message", fmt.Sprintf("executing: docker tag %s %s", imageSha, retaggedNameWithTag))

	err := r.docker.ImageTag(context.Background(), imageSha, retaggedNameWithTag)
	if err != nil {
		return "", microerror.Maskf(err, "could not tag image")
	}

	return retaggedNameWithTag, nil
}

func (r *Registry) PushImage(destinationImage, destinationTag string) error {
	if destinationImage == "" {
		return microerror.Maskf(invalidArgumentError, "destinationImage should not be empty")
	}
	if destinationTag == "" {
		return microerror.Maskf(invalidArgumentError, "destinationTag should not be empty")
	}

	imageTag := images.ImageWithTag(destinationImage, destinationTag)

	r.logger.Log("level", "debug", "message", fmt.Sprintf("executing: docker push %s", imageTag))

	opts := types.ImagePushOptions{
		All:          true,
		RegistryAuth: r.getRegistryAuthBase64(),
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

func (r *Registry) RebuildImage(sourceImage, sha, destinationImage, destinationTag string, dockerfileOptions []string) (string, error) {
	if sourceImage == "" {
		return "", microerror.Maskf(invalidArgumentError, "sourceImage should not be empty")
	}
	if sha == "" {
		return "", microerror.Maskf(invalidArgumentError, "%s SHA should not be empty", sourceImage)
	}
	if destinationImage == "" {
		return "", microerror.Maskf(invalidArgumentError, "destinationImage should not be empty")
	}
	if destinationTag == "" {
		return "", microerror.Maskf(invalidArgumentError, "destinationTag should not be empty")
	}

	retaggedNameWithTag := images.ImageWithTag(destinationImage, destinationTag)

	dockerfile := Dockerfile{
		BaseImage:         sourceImage,
		Sha:               sha,
		DockerfileOptions: dockerfileOptions,
		Tag:               destinationTag,
	}

	f, err := os.Create(TempDockerfileName(destinationTag))
	if err != nil {
		return "", microerror.Mask(err)
	}

	// render Dockerfile with overrides
	t := template.Must(template.New("").Parse(customDockerfileTmpl))
	err = t.Execute(f, dockerfile)
	if err != nil {
		return "", microerror.Mask(invalidTemplateError)
	}

	rebuild := exec.Command("docker", "build", "-t", retaggedNameWithTag, "-f", TempDockerfileName(destinationTag), ".")
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

		r.logger.Log("level", "debug", "message", fmt.Sprintf("docker status"), "docker", string(s.Bytes()))
	}
	if err := s.Err(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) getRegistryAuthBase64() string {
	auth := fmt.Sprintf(`{"username": "%s", "password": "%s", "serveraddress": "https://%s"}`, r.username, r.password, r.host)

	return base64.StdEncoding.EncodeToString([]byte(auth))
}
