package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
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
	url := fmt.Sprintf("https://%s/v2/%s/tags/list", r.host, ImageName(r.organisation, image))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	token, err := r.getToken(req)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	var tags []string
	o := func() error {
		res, err := r.client.Do(req)
		if err != nil {
			return microerror.Mask(err)
		}
		defer res.Body.Close()

		switch res.StatusCode {
		case http.StatusOK:
			tagResponse := &TagsListResponse{}
			err = json.NewDecoder(res.Body).Decode(tagResponse)
			if err != nil {
				return microerror.Mask(err)
			}

			tags = tagResponse.Tags
			return nil
		default:
			return microerror.Maskf(invalidStatusCodeError, "could not check retag status: %d", res.StatusCode)
		}
	}
	b := backoff.NewExponential(10*time.Second, 1*time.Second)
	err = backoff.Retry(o, b)
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

func (r *Registry) getToken(req *http.Request) (string, error) {
	const authenticationHeaderKey = "www-authenticate"

	res, err := r.client.Do(req)
	if err != nil {
		return "", microerror.Mask(err)
	}
	defer res.Body.Close()

	var authenticationHeaderValue []string
	// the authentication header can be found as www-authenticate, Www-Authenticate
	// or WWW-Authenticate.
	for k, v := range res.Header {
		if strings.ToLower(k) == authenticationHeaderKey {
			authenticationHeaderValue = v
			break
		}
	}
	if len(authenticationHeaderValue) == 0 {
		// no need for authentication
		return "", nil
	}

	authURL, err := getAuthURL(authenticationHeaderValue[0])
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

func (r *Registry) GetDigest(image string, tag string) (string, error) {
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", r.host, ImageName(r.organisation, image), tag)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", microerror.Mask(err)
	}
	token, err := r.getToken(req)
	if err != nil {
		return "", microerror.Mask(err)
	}
	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	var digest string
	o := func() error {
		res, err := r.client.Do(req)
		if err != nil {
			return microerror.Mask(err)
		}
		defer res.Body.Close()

		switch res.StatusCode {
		case http.StatusOK:
			digest = res.Header.Get("docker-content-digest")
			if digest == "" {
				return microerror.Maskf(invalidStatusCodeError, "remote didn't return docker-content-digest header")
			}
			return nil
		default:
			return microerror.Maskf(invalidStatusCodeError, "could not get manifest: %d", res.StatusCode)
		}
	}
	b := backoff.NewExponential(10*time.Second, 1*time.Second)
	err = backoff.Retry(o, b)
	if err != nil {
		return "", microerror.Mask(err)
	}

	digest = strings.TrimPrefix(digest, "sha256:")

	return digest, nil
}

func (r *Registry) DeleteImage(image string, tag string) error {
	digest, err := r.GetDigest(image, tag)
	if err != nil {
		return microerror.Mask(err)
	}

	const delete = "/v2/%s/manifests/%s"
	// returns 202

	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", r.host, ImageName(r.organisation, image), digest)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return microerror.Mask(err)
	}
	token, err := r.getToken(req)
	if err != nil {
		return microerror.Mask(err)
	}
	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	o := func() error {
		res, err := r.client.Do(req)
		if err != nil {
			return microerror.Mask(err)
		}
		defer res.Body.Close()

		switch res.StatusCode {
		case http.StatusAccepted:
			return nil
		default:
			return microerror.Maskf(invalidStatusCodeError, "could not get manifest: %d", res.StatusCode)
		}
	}
	b := backoff.NewExponential(10*time.Second, 1*time.Second)
	err = backoff.Retry(o, b)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
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
