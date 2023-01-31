package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	nurl "net/url"
	"regexp"
	"strings"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/retagger/pkg/images"
)

// This code is largely lifted from nokia/docker-registry-client. It provides Quay-specific API behavior needed by retagger.

// RequestOptions wraps the url and optional query parameters for the HTTP request.
type RequestOptions struct {
	URL   string
	Query nurl.Values
}

// TagsResponse wraps a response from the Quay tags endpoint.
type TagsResponse struct {
	Tags []QuayTag `json:"tags"`
}

// QuayTag describes a tag object returned by the Quay API.
type QuayTag struct {
	Name           string `json:"name"`
	ImageID        string `json:"image_id"`
	ManifestDigest string `json:"manifest_digest"`
	Modified       string `json:"last_modified"`
	DockerImageID  string `json:"docker_image_id"`
	IsManifestList bool   `json:"is_manifest_list"`
	Size           int64  `json:"size"`
	Reversion      bool   `json:"reversion"`
	StartTS        int64  `json:"start_ts"`
}

func (t *QuayTag) GetName() string {
	return t.Name
}

func (t *QuayTag) GetImageID() string {
	return t.ImageID
}

func (t *QuayTag) GetDigest() string {
	return t.ManifestDigest
}

func (t *QuayTag) GetSize() int64 {
	return t.Size
}

// Matches an RFC 5988 (https://tools.ietf.org/html/rfc5988#section-5)
// Link header. For example,
//
//	<http://registry.example.com/v2/_catalog?n=5&last=tag5>; type="application/json"; rel="next"
//
// The URL is _supposed_ to be wrapped by angle brackets `< ... >`,
// but e.g., quay.io does not include them. Similarly, params like
// `rel="next"` may not have quoted values in the wild.
var nextLinkRE = regexp.MustCompile(`^ *<?([^;>]+)>? *(?:;[^;]*)*; *rel="?next"?(?:;.*)?`)

// GetQuayTagsWithDetails fetches tags for the given image including extra information defined in a QuayTag
// This uses the Quay API, so assumes a Quay host. Other hosts are likely to fail.
func (r *Registry) GetQuayTagsWithDetails(image string) (tags []QuayTag, err error) {
	if r.host == "giantswarm-registry.cn-shanghai.cr.aliyuncs.com" {
		// Get Aliyun tags instead
		return r.GetAliyunTagsWithDetails(image)
	}

	url := fmt.Sprintf("https://%s/api/v1/repository/%s/tag/", r.host, images.Name(r.organisation, image))

	// The Quay API includes deleted tags by default. Limit our request to active tags.
	q := nurl.Values{}
	q.Set("onlyActiveTags", "true")

	req := RequestOptions{
		URL:   url,
		Query: q,
	}

	var response TagsResponse
	o := func() error {
		for {
			r.logger.Log("level", "debug", "message", fmt.Sprintf("requesting registry tags from %s", url))

			url, err = r.getPaginatedJSON(req, &response)
			if err != nil {
				if IsNoMorePages(err) {
					tags = append(tags, response.Tags...)
					return nil
				}
				if strings.Contains(err.Error(), "Requires authentication") {
					msg := fmt.Sprintf("unable to list tags for %s: HTTP 401 - Requires authentication. If the repository does not exist, retagger will try to create it. If the repository exists, check that the user has access to it", image)
					r.logger.Log("level", "warn", "message", msg)
					return nil
				}

				return err
			}
			req.URL = url
			tags = append(tags, response.Tags...)
		}
	}
	b := backoff.NewExponential(500*time.Millisecond, 5*time.Second)
	err = backoff.Retry(o, b)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tags, nil
}

// GetQuayTagMap fetches the tag details for the given image, and returns a map containing
// the tags as keys and the tag details (QuayTag) as values.
func (r *Registry) GetQuayTagMap(image string) (map[string]QuayTag, error) {
	existingQuayTags, err := r.GetQuayTagsWithDetails(image)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	quayTagMap := make(map[string]QuayTag)
	for _, t := range existingQuayTags {
		quayTagMap[t.Name] = t
	}
	return quayTagMap, nil
}

// getPaginatedJSON accepts a string and a pointer, and returns the
// next page URL while updating pointed-to variable with a parsed JSON
// value. When there are no more pages it returns `ErrNoMorePages`.
func (r *Registry) getPaginatedJSON(request RequestOptions, response interface{}) (string, error) {
	req, err := http.NewRequest("GET", request.URL, nil)
	if err != nil {
		return "", err
	}

	// If caller supplied additional query parameters, add them to the request
	if request.Query != nil {
		q := req.URL.Query()
		for key, vals := range request.Query {
			for _, v := range vals {
				q.Add(key, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := r.registryClient.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(response)
	if err != nil {
		return "", err
	}

	nextURI, err := getNextLink(resp)
	if err != nil {
		return "", err
	}

	base, err := nurl.Parse(r.host)
	if err != nil {
		return "", err
	}

	u, err := nurl.Parse(nextURI)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(u).String(), nil
}

// Parses an HTTP response header looking for the "next" link in a paginated response.
func getNextLink(resp *http.Response) (string, error) {
	for _, link := range resp.Header[http.CanonicalHeaderKey("Link")] {
		parts := nextLinkRE.FindStringSubmatch(link)
		if parts != nil {
			return parts[1], nil
		}
	}
	return "", microerror.Mask(noMorePagesError)
}
