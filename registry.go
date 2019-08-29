package main

import (
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/giantswarm/microerror"
	"github.com/nokia/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"
)

const customDockerfileTmpl = `FROM {{ .BaseImage }}:{{ .Tag }}
{{range .DockerfileOptions -}}
{{ . }}
{{ end -}}
`

type Dockerfile struct {
	BaseImage         string
	DockerfileOptions []string
	Tag               string
}

type RegistryConfig struct {
	Host         string
	Organisation string
	Password     string
	Username     string
}

type Registry struct {
	registryClient *registry.Registry

	host         string
	organisation string
	password     string
	username     string
}

func NewRegistry(cfg *RegistryConfig) (*Registry, error) {
	if cfg.Host == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Host must not be empty", cfg)
	}

	if cfg.Organisation == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Organisation must not be empty", cfg)
	}

	if cfg.Username == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Username must not be empty", cfg)
	}

	if cfg.Password == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Password must not be empty", cfg)
	}

	registryClient, err := registry.New(fmt.Sprintf("https://%s", cfg.Host), cfg.Username, cfg.Password)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	qr := &Registry{
		host:           cfg.Host,
		organisation:   cfg.Organisation,
		password:       cfg.Password,
		username:       cfg.Username,
		registryClient: registryClient,
	}

	return qr, nil
}

func (r *Registry) Login() error {
	login := exec.Command("docker", "login", "-u", r.username, "-p", r.password, r.host)
	if err := Run(login); err != nil {
		return fmt.Errorf("could not login to registry: %v", err)
	}
	return nil
}

func (r *Registry) CheckImageTagExists(image, tag string) (bool, error) {
	tags, err := r.ListImageTags(image)
	if err != nil {
		return false, microerror.Mask(err)
	}

	for _, imageTag := range tags {
		if imageTag == tag {
			return true, nil
		}
	}

	return false, nil
}

func (r *Registry) ListImageTags(image string) ([]string, error) {
	tags, err := r.registryClient.Tags(ImageName(r.organisation, image))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tags, nil
}

func (r *Registry) Retag(image, sha, tag string) (string, error) {
	retaggedName := RetaggedName(r.host, r.organisation, image)
	retaggedNameWithTag := ImageWithTag(retaggedName, tag)

	retag := exec.Command("docker", "tag", sha, retaggedNameWithTag)
	err := Run(retag)
	if err != nil {
		return "", microerror.Mask(err)
	}
	return retaggedNameWithTag, nil
}

func (r *Registry) Rebuild(image, tag string, customImage CustomImage) (string, error) {
	RetaggedName := RetaggedName(r.host, r.organisation, image)
	rebuiltImageTag := ImageWithTag(RetaggedName, fmt.Sprintf("%s-%s", tag, customImage.TagSuffix))

	dockerfile := Dockerfile{
		BaseImage:         image,
		DockerfileOptions: customImage.DockerfileOptions,
		Tag:               tag,
	}

	f, err := os.Create(fmt.Sprintf("Dockerfile-%s", customImage.TagSuffix))
	if err != nil {
		return "", microerror.Mask(err)
	}

	// render Dockerfile with overrides
	t := template.Must(template.New("").Parse(customDockerfileTmpl))
	err = t.Execute(f, dockerfile)
	if err != nil {
		return "", microerror.Mask(invalidTemplateError)
	}

	rebuild := exec.Command("docker", "build", "-t", rebuiltImageTag, "-f", fmt.Sprintf("Dockerfile-%s", customImage.TagSuffix), ".")
	err = Run(rebuild)
	if err != nil {
		return "", microerror.Mask(err)
	}
	return rebuiltImageTag, nil
}

func (r *Registry) GetDigest(image string, tag string) (digest.Digest, error) {
	digest, err := r.registryClient.ManifestV2Digest(ImageName(r.organisation, image), tag)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return digest, nil
}

func (r *Registry) DeleteImage(image string, tag string) error {
	digest, err := r.GetDigest(image, tag)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.registryClient.DeleteManifest(ImageName(r.organisation, image), digest)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
