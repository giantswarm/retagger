package quay

import (
	"encoding/json"
	"strings"
)

type Tag struct {
	ImageID        string `json:"image_id"`
	LastModified   string `json:"last_modified"`
	Name           string `json:"name"`
	ManifestDigest string `json:"manifest_digest"`
	Size           int    `json:"size"`
}
type Repository struct {
	TrustEnabled   bool           `json:"trust_enabled"`
	Description    string         `json:"description"`
	Tags           map[string]Tag `json:"tags"`
	TagExpirationS int            `json:"tag_expiration_s"`
	IsPublic       bool           `json:"is_public"`
	IsStarred      bool           `json:"is_starred"`
	Kind           string         `json:"kind"`
	Name           string         `json:"name"`
	Namespace      string         `json:"namespace"`
	IsOrganization bool           `json:"is_organization"`
	CanWrite       bool           `json:"can_write"`
	StatusToken    string         `json:"status_token"`
	CanAdmin       bool           `json:"can_admin"`
}

type RepositoryInput struct {
	RepoKind    string `json:"repo_kind"`
	Namespace   string `json:"namespace"`
	Visibility  string `json:"visibility"`
	Repository  string `json:"repository"`
	Description string `json:"description"`
}

func postRepositoryPath() (string, string) {
	return "POST", "/api/v1/repository"
}

func getRepositoryPath(repositoryName string) (string, string) {
	return "GET", "/api/v1/repository/" + repositoryName
}

func (c *Client) GetRepository(repositoryName string) (*Repository, error) {
	m, u := getRepositoryPath(repositoryName)
	statuscode, status, body, err := c.do(m, u, nil)
	if err != nil {
		return nil, err
	}

	if statuscode != 200 {
		if statuscode == 404 {
			return nil, nil
		}
		return nil, getAPIError(status, body)
	}

	var repo Repository
	err = json.Unmarshal(body, &repo)

	if err != nil {
		return nil, err
	}

	return &repo, nil

}

func (c *Client) CreateRepository(repositoryName string, visibility string) (*Repository, error) {
	rI := RepositoryInput{
		Namespace:  strings.Split(repositoryName, "/")[0],
		Repository: strings.Split(repositoryName, "/")[1],
		Visibility: visibility,
		RepoKind:   "image",
	}

	m, u := postRepositoryPath()
	statuscode, status, body, err := c.do(m, u, mustReader(rI))
	if err != nil {
		return nil, err
	}

	if statuscode != 201 {
		return nil, getAPIError(status, body)
	}

	return c.GetRepository(repositoryName)
}
