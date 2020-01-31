package registry

import (
	"context"
	"fmt"
	"net/http"
	nurl "net/url"
	"time"

	dockerRef "github.com/docker/distribution/reference"
	dockerclient "github.com/docker/docker/client"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/nokia/docker-registry-client/registry"
	godigest "github.com/opencontainers/go-digest"

	"github.com/giantswarm/retagger/pkg/images"
)

type Tag interface {
	GetName() string
	GetImageID() string
	GetDigest() string
	GetSize() int64
}

type Config struct {
	AccessKey    string
	AccessSecret string
	AliyunRegion string
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
	accessKey    string
	accessSecret string
	aliyunRegion string
}

type BackoffTransport struct {
	Transport http.RoundTripper
	Logger micrologger.Logger
}

func (t *BackoffTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var resp *http.Response
	o := func () error {
		var respErr error
		resp, respErr = t.Transport.RoundTrip(request)
		if respErr != nil {
			return respErr
		}
		return nil
	}
	b := backoff.NewExponential(30*time.Second, 5*time.Second)
	n := backoff.NewNotifier(t.Logger, context.Background())

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return resp, nil
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
		url := fmt.Sprintf("https://%s", config.Host)
		logf := registry.Log
		if config.LogFunc != nil {
			logf = config.LogFunc
		}
		registryClient = &registry.Registry{
			URL: url,
			Client: &http.Client{
				Transport: &registry.ErrorTransport{
					Transport: &BackoffTransport{
						Transport: &registry.BasicTransport{
							Transport: &registry.TokenTransport{
								Transport: http.DefaultTransport,
								Username:  config.Username,
								Password:  config.Password,
							},
							URL:      url,
							Username: config.Username,
							Password: config.Password,
						},
					},
				},
			},
			Logf: logf,
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
		accessKey:    config.AccessKey,
		accessSecret: config.AccessSecret,
		aliyunRegion: config.AliyunRegion,
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

func (r *Registry) GetDigest(image string, tag string) (godigest.Digest, error) {
	digest, err := r.registryClient.ManifestV2Digest(images.Name(r.organisation, image), tag)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return digest, nil
}

// GuessRegistryPath examines the given image string, determines whether it describes a full
// image path, is hosted on Docker hub, or belongs to the Docker standard library, and returns
//  a URL representing the full path.
func (r *Registry) GuessRegistryPath(image string) (*nurl.URL, error) {

	dockerName, err := dockerRef.ParseNormalizedNamed(image)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	url, err := nurl.Parse(fmt.Sprintf("https://%s", dockerName.String()))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// The normalized docker reference uses docker.io as the default domain, but this does not redirect API paths,
	// so we override this host here to point to the API endpoint instead of the default docker backend
	// https://docker.io -- > registry.hub.docker.com
	if url.Hostname() == "docker.io" {
		url.Host = "registry.hub.docker.com"
	}

	return url, nil
}

// GetRepositoryFromPathString guesses the full path of an image and returns the organization/image for the image.
func (r *Registry) GetRepositoryFromPathString(path string) (string, error) {
	name, err := r.getDockerName(path)
	if err != nil {
		return "", err
	}
	return dockerRef.FamiliarString(name), nil
}

func (r *Registry) ListImageTags(image string) ([]string, error) {
	var tags []string
	o := func() error {
		imageTags, err := r.registryClient.Tags(images.Name(r.organisation, image))
		if IsRepositoryNotFound(err) {
			_ = r.logger.Log("level", "warning", "message", fmt.Sprintf("repository %s was not found in registry, retagger will try create the repository", image))
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

func (r *Registry) RetaggedName(image string) string {
	return images.RetaggedName(r.host, r.organisation, image)
}

func (r *Registry) getDockerName(image string) (dockerRef.Named, error) {
	return dockerRef.ParseNormalizedNamed(image)
}
