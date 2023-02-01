package registry

import (
	"context"
	"crypto/tls"
	"net/http"
	"strconv"
	"strings"

	harbor "github.com/x893675/go-harbor"
	"github.com/x893675/go-harbor/schema"
)

func (r *Registry) GetHarborTagsWithDetails(image string) (tags []QuayTag, err error) {
	parts := strings.Split(image, "/")
	imageName := parts[len(parts)-1]

	registryHost := r.host
	if !strings.HasPrefix(r.host, "https://") {
		registryHost = "https://" + r.host
	}

	ctx := context.Background()

	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	harborClient, err := harbor.NewClientWithOpts(harbor.WithHost(registryHost),
		harbor.WithHTTPClient(c),
		harbor.WithBasicAuth(r.username, r.password))
	if err != nil {
		panic(err)
	}

	var harborTags []QuayTag
	currentPage := "1"
	for {
		withTags := true
		opts := schema.ArtifactsListOptions{
			ProjectName:    r.organisation,
			RepositoryName: imageName,
			WithTag:        &withTags,
			Page:           currentPage,
		}
		listArtifacts, err := harborClient.ListArtifacts(ctx, opts)

		if err != nil {
			r.logger.Log("level", "warn", "message", "Harbor repository does not exist. Retagger will try to create it")
			return harborTags, nil
		}

		if len(listArtifacts) == 0 {
			break
		}

		for _, artifact := range listArtifacts {
			for tagCount := 0; tagCount < len(artifact.Tags); tagCount++ {
				harborTags = append(harborTags, QuayTag{
					Name:           artifact.Tags[tagCount].Name,
					Size:           artifact.Size,
					ManifestDigest: artifact.Digest,
					ImageID:        imageName,
				})
			}
		}
		currentPage, err = nextPage(currentPage)
		if err != nil {
			r.logger.Log("level", "debug", "message", "Error converting page number")
			break
		}
	}

	return harborTags, nil
}

func nextPage(currentPage string) (string, error) {
	page, err := strconv.Atoi(currentPage)

	if err != nil {
		return "", err
	}

	page = page + 1
	return strconv.Itoa(page), nil
}
