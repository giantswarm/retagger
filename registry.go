package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/giantswarm/microerror"
)

type RegistryConfig struct {
	Client *http.Client

	Host         string
	Organisation string
	Password     string
	Token        string
	Username     string
}

type Registry struct {
	client *http.Client

	host         string
	organisation string
	password     string
	token        string
	username     string
}

func NewRegistry(cfg *RegistryConfig) (*Registry, error) {
	if cfg.Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Client must not be empty", cfg)
	}
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

	if cfg.Token == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Token must not be empty", cfg)
	}

	qr := &Registry{
		client: cfg.Client,

		host:         cfg.Host,
		organisation: cfg.Organisation,
		password:     cfg.Password,
		token:        cfg.Token,
		username:     cfg.Username,
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
	url := fmt.Sprintf("https://%s/api/v1/repository/%s/tag/%s/images", r.host, ImageName(r.organisation, image), tag)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, microerror.Mask(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.token))

	res, err := r.client.Do(req)
	if err != nil {
		return false, microerror.Mask(err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	default:
		log.Printf("could not check retag status: %v", res.StatusCode)
		return false, nil
	}
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
