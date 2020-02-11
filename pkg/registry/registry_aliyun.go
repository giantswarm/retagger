package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"
	"github.com/giantswarm/microerror"
)

type AliyunTag struct {
	ImageUpdate int64  `json:"imageUpdate"`
	ImageCreate int64  `json:"imageCreate"`
	ImageID     string `json:"imageId"`
	Digest      string `json:"digest"`
	ImageSize   int64  `json:"imageSize"`
	Tag         string `json:"tag"`
	Status      string `json:"status"`
}

type AliyunResponseWrapper struct {
	Data AliyunTagsResponse `json:"data"`
}

type AliyunTagsResponse struct {
	Tags         []AliyunTag `json:"tags"`
	Page         int64       `json:"page"`
	PageSize     int64       `json:"pageSize"`
	TotalRecords int64       `json:"total"`
}

func (t *AliyunTag) GetName() string {
	return t.Tag
}

func (t *AliyunTag) GetImageID() string {
	return t.ImageID
}

func (t *AliyunTag) GetDigest() string {
	return t.Digest
}

func (t *AliyunTag) GetSize() int64 {
	return t.ImageSize
}

func (r *Registry) GetAliyunTagsWithDetails(image string) (tags []QuayTag, err error) {
	crClient, err := cr.NewClientWithAccessKey(r.aliyunRegion, r.accessKey, r.accessSecret)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	path, err := r.GuessRegistryPath(image)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Remove leading slash and namespace.
	shortName := strings.Split(strings.Trim(path.Path, "/"), "/")[1]
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tagRequest := cr.CreateGetRepoTagsRequest()
	tagRequest.Domain = fmt.Sprintf("cr.%s.aliyuncs.com", r.aliyunRegion)
	tagRequest.RepoNamespace = "giantswarm"
	tagRequest.RepoName = shortName

	var aliTags []QuayTag // TODO: Make AliyunTag
	page := 1
	for {
		tagRequest.Page = requests.NewInteger(page)

		crResp, err := crClient.GetRepoTags(tagRequest)
		serverError, isAliyunError := err.(*errors.ServerError)
		if isAliyunError {
			if strings.Contains(serverError.Message(), "REPO_NOT_EXIST") {
				r.logger.Log("level", "warn", "message", "Aliyun repository does not exist. Retagger will try to create it")
				return aliTags, nil
			}
		}
		if err != nil {
			return nil, microerror.Mask(err)
		}

		apiResponse := &AliyunResponseWrapper{}

		resp := json.NewDecoder(bytes.NewReader(crResp.GetHttpContentBytes()))
		err = resp.Decode(apiResponse)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, tag := range apiResponse.Data.Tags {
			aliTags = append(aliTags, quayTagFromTag(&tag))
		}

		remaining := int(apiResponse.Data.TotalRecords) - len(aliTags)
		if remaining == 0 {
			break
		}

		page++
	}

	return aliTags, nil
}

func quayTagFromTag(t Tag) QuayTag {
	return QuayTag{
		Name:           t.GetName(),
		ImageID:        t.GetImageID(),
		ManifestDigest: t.GetDigest(),
		Size:           t.GetSize(),
	}
}
