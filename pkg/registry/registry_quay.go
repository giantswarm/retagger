package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	nurl "net/url"
	"regexp"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/retagger/pkg/images"
)

// This code is largely lifted from nokia/docker-registry-client. It provides Quay-specific API behavior needed by retagger.

// TagsResponse wraps a response from the Quay tags endpoint
type TagsResponse struct {
	Tags []QuayTag `json:"tags"`
}

// QuayTag describes a tag object returned by the Quay API
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

// Matches an RFC 5988 (https://tools.ietf.org/html/rfc5988#section-5)
// Link header. For example,
//
//    <http://registry.example.com/v2/_catalog?n=5&last=tag5>; type="application/json"; rel="next"
//
// The URL is _supposed_ to be wrapped by angle brackets `< ... >`,
// but e.g., quay.io does not include them. Similarly, params like
// `rel="next"` may not have quoted values in the wild.
var nextLinkRE = regexp.MustCompile(`^ *<?([^;>]+)>? *(?:;[^;]*)*; *rel="?next"?(?:;.*)?`)

// GetQuayTagsWithDetails fetches tags for the given image including extra information defined in a QuayTag
// This uses the Quay API, so assumes a Quay host. Other hosts are likely to fail.
func (r *Registry) GetQuayTagsWithDetails(image string) (tags []QuayTag, err error) {
	url := fmt.Sprintf("https://%s/api/v1/repository/%s/tag/", r.host, images.Name(r.organisation, image))

	var response TagsResponse
	for {

		r.logger.Log("level", "debug", "message", fmt.Sprintf("requesting registry tags from %s", url))

		url, err = r.getPaginatedJSON(url, &response)
		if err != nil {
			if IsNoMorePages(err) {
				tags = append(tags, response.Tags...)
				return tags, nil
			}

			return nil, err
		}
		tags = append(tags, response.Tags...)
	}
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
func (r *Registry) getPaginatedJSON(url string, response interface{}) (string, error) {
	resp, err := r.registryClient.Client.Get(url)
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

// Parses an HTTP response header looking for the "next" link in a paginated response
func getNextLink(resp *http.Response) (string, error) {
	for _, link := range resp.Header[http.CanonicalHeaderKey("Link")] {
		parts := nextLinkRE.FindStringSubmatch(link)
		if parts != nil {
			return parts[1], nil
		}
	}
	return "", microerror.Mask(noMorePagesError)
}
