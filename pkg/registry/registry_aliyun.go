package registry

import (
	"bytes"
	"encoding/json"
	"strings"

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
	crClient, err := cr.NewClientWithAccessKey("cn-shanghai", r.accessKey, r.accessSecret)
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
	tagRequest.Domain = "cr.cn-shanghai.aliyuncs.com"
	tagRequest.RepoNamespace = "giantswarm"
	tagRequest.RepoName = shortName

	var aliTags []QuayTag // TODO: Make AliyunTag
	done := false
	page := 1
	for !done {
		tagRequest.Page = requests.NewInteger(page)

		crResp, err := crClient.GetRepoTags(tagRequest)
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

		// We're done when we have consumed the last page
		if apiResponse.Data.TotalRecords%apiResponse.Data.PageSize == 0 {
			// Results fit perfectly into page size, so the last page is total / size
			if int64(page) == (apiResponse.Data.TotalRecords / apiResponse.Data.PageSize) {
				done = true
			}
		} else if int64(page) > (apiResponse.Data.TotalRecords / apiResponse.Data.PageSize) {
			// Results don't fit evenly into pagination, so last page is (total / size) + 1
			done = true
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
