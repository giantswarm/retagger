package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/giantswarm/microerror"
)

type RegistryConfig struct {
	Client *http.Client

	Host         string
	Organisation string
	Password     string
	Username     string
}

type Registry struct {
	client *http.Client

	host         string
	organisation string
	password     string
	username     string
}

type TagsListResponse struct {
	Tags []string `json:"tags"`
}

type TokenRequestResponse struct {
	Token string `json:"token"`
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

	qr := &Registry{
		client: cfg.Client,

		host:         cfg.Host,
		organisation: cfg.Organisation,
		password:     cfg.Password,
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
	url := fmt.Sprintf("https://%s/v2/%s/tags/list", r.host, ImageName(r.organisation, image))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, microerror.Mask(err)
	}
	token, err := r.getToken(req)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	res, err := r.client.Do(req)
	if err != nil {
		return false, microerror.Mask(err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		tagResponse := &TagsListResponse{}
		err = json.NewDecoder(res.Body).Decode(tagResponse)
		if err != nil {
			return false, microerror.Mask(err)
		}
		for _, imageTag := range tagResponse.Tags {
			if imageTag == tag {
				return true, nil
			}
		}
		return false, nil
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

func (r *Registry) getToken(req *http.Request) (string, error) {
	const authorizationHeaderKey = "Www-Authenticate"

	res, err := r.client.Do(req)
	if err != nil {
		return "", microerror.Mask(err)
	}
	defer res.Body.Close()

	authorizationHeaderValue := res.Header[authorizationHeaderKey]
	if len(authorizationHeaderValue) == 0 {
		// no need for authorization
		return "", nil
	}

	authURL, err := getAuthURL(authorizationHeaderValue[0])
	if err != nil {
		return "", microerror.Mask(err)
	}

	reqToken, err := http.NewRequest(http.MethodGet, authURL, nil)
	if err != nil {
		return "", microerror.Mask(err)
	}
	resToken, err := r.client.Do(reqToken)
	if err != nil {
		return "", microerror.Mask(err)
	}
	defer resToken.Body.Close()

	tokenResponse := &TokenRequestResponse{}
	err = json.NewDecoder(resToken.Body).Decode(tokenResponse)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return tokenResponse.Token, nil
}

func getAuthURL(authenticateChallenge string) (string, error) {
	// www-authenticate headers have this form:
	// Bearer realm="<realm>",service="<service>"[,scope="<scope>"]

	authenticateChallenge = strings.Replace(authenticateChallenge, `"`, "", -1)

	parts := strings.Fields(authenticateChallenge)
	if len(parts) < 2 {
		return "", microerror.Mask(invalidAuthenticateChallengeError)
	}
	items := strings.Split(parts[1], ",")
	if len(items) < 2 {
		return "", microerror.Mask(invalidAuthenticateChallengeError)
	}
	kv := strings.Split(items[0], "=")
	if len(kv) < 2 {
		return "", microerror.Mask(invalidAuthenticateChallengeError)
	}
	realm := kv[1]
	kv = strings.Split(items[1], "=")
	if len(kv) < 2 {
		return "", microerror.Mask(invalidAuthenticateChallengeError)
	}
	service := kv[1]
	var scope string
	if len(items) == 3 {
		kv = strings.Split(items[2], "=")
		if len(kv) < 2 {
			return "", microerror.Mask(invalidAuthenticateChallengeError)
		}
		scope = kv[1]
	}

	url := fmt.Sprintf("%s?service=%s&scope=%s", realm, service, scope)

	return url, nil
}
