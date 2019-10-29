package registry

import (
	"fmt"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/nokia/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"

	"github.com/giantswarm/retagger/pkg/images"
)

type Config struct {
	Host         string
	Organisation string
	Password     string
	Username     string
	LogFunc      func(format string, args ...interface{})
	Logger       micrologger.Logger
}

type Registry struct {
	registryClient *registry.Registry
	logger         micrologger.Logger
	docker         *dockerclient.Client

	host         string
	organisation string
	password     string
	username     string
}

func New(config Config) (*Registry, error) {
	if config.Host == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Host must not be empty", config)
	}
	if config.Organisation == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Organisation must not be empty", config)
	}
	if config.Username == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Username must not be empty", config)
	}
	if config.Password == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Password must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be nil", config)
	}

	var err error

	var registryClient *registry.Registry
	{
		o := registry.Options{
			Username: config.Username,
			Password: config.Password,
		}

		if config.LogFunc != nil {
			o.Logf = config.LogFunc
		}

		registryClient, err = registry.NewCustom(fmt.Sprintf("https://%s", config.Host), o)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var dockerClient *dockerclient.Client
	{
		dockerClient, err = dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithVersion("1.38"))
		if err != nil {
			return nil, microerror.Maskf(err, "dockerclient.New")
		}
	}

	qr := &Registry{
		host:         config.Host,
		organisation: config.Organisation,
		password:     config.Password,
		username:     config.Username,
		logger:       config.Logger,

		registryClient: registryClient,
		docker:         dockerClient,
	}

	return qr, nil
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
	var tags []string
	o := func() error {
		imageTags, err := r.registryClient.Tags(images.Name(r.organisation, image))
		if IsRepositoryNotFound(err) {
			r.logger.Log("level", "warning", "message", fmt.Sprintf("repository %s was not found in registry, retagger will try create the repository", image))
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		tags = imageTags
		return nil
	}
	b := backoff.NewExponential(500*time.Millisecond, 5*time.Second)
	err := backoff.Retry(o, b)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tags, nil
}

func (r *Registry) GetDigest(image string, tag string) (digest.Digest, error) {
	digest, err := r.registryClient.ManifestV2Digest(images.Name(r.organisation, image), tag)
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

	err = r.registryClient.DeleteManifest(images.Name(r.organisation, image), digest)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) RetaggedName(image string) string {
	return images.RetaggedName(r.host, r.organisation, image)
}
