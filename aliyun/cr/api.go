package cr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"
)

type Client struct {
	client *cr.Client
	domain string
}

type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

func DefaultClient(accesskey, accessSecret, region string) (*Client, error) {
	crClient, err := cr.NewClientWithAccessKey(
		region,
		accesskey,
		accessSecret)

	if err != nil {
		return nil, err
	}
	return &Client{
		client: crClient,
		domain: fmt.Sprintf("cr.%s.aliyuncs.com", region),
	}, nil
}

func (c *Client) GetRepository(repositoryName string) (*Repository, error) {

	crReq := cr.CreateGetRepoRequest()
	crReq.SetDomain(c.domain)
	crReq.RepoNamespace = strings.Split(repositoryName, "/")[0]
	crReq.RepoName = strings.Split(repositoryName, "/")[1]

	resp, err := c.client.GetRepo(crReq)

	if err != nil {

		if resp.GetHttpStatus() == 404 {
			return nil, nil
		}
		return nil, err
	}

	var r RepoResponse

	err = json.Unmarshal(resp.GetHttpContentBytes(), &r)
	if err != nil {
		return nil, err
	}

	return &r.Data.Repo, nil
}

func (c *Client) CreateRepository(repositoryName string, isPublic bool) (*Repository, error) {

	t := "PRIVATE"
	if isPublic {
		t = "PUBLIC"
	}

	r := RepositoryRequest{
		Repo: RepositoryInput{
			RepoType:      t,
			RepoNamespace: strings.Split(repositoryName, "/")[0],
			RepoName:      strings.Split(repositoryName, "/")[1],
			Summary:       repositoryName + " image",
		},
	}

	crReq := cr.CreateCreateRepoRequest()
	crReq.SetDomain(c.domain)
	crReq.SetContent(mustMarshal(r))
	crReq.SetContentType("application/json")

	_, err := c.client.CreateRepo(crReq)

	if err != nil {
		return nil, err
	}

	return c.GetRepository(repositoryName)
}

func mustMarshal(i interface{}) []byte {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}

	return b
}
