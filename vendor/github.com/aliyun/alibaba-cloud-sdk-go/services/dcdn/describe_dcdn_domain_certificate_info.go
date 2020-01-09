package dcdn

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// DescribeDcdnDomainCertificateInfo invokes the dcdn.DescribeDcdnDomainCertificateInfo API synchronously
// api document: https://help.aliyun.com/api/dcdn/describedcdndomaincertificateinfo.html
func (client *Client) DescribeDcdnDomainCertificateInfo(request *DescribeDcdnDomainCertificateInfoRequest) (response *DescribeDcdnDomainCertificateInfoResponse, err error) {
	response = CreateDescribeDcdnDomainCertificateInfoResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeDcdnDomainCertificateInfoWithChan invokes the dcdn.DescribeDcdnDomainCertificateInfo API asynchronously
// api document: https://help.aliyun.com/api/dcdn/describedcdndomaincertificateinfo.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeDcdnDomainCertificateInfoWithChan(request *DescribeDcdnDomainCertificateInfoRequest) (<-chan *DescribeDcdnDomainCertificateInfoResponse, <-chan error) {
	responseChan := make(chan *DescribeDcdnDomainCertificateInfoResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeDcdnDomainCertificateInfo(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// DescribeDcdnDomainCertificateInfoWithCallback invokes the dcdn.DescribeDcdnDomainCertificateInfo API asynchronously
// api document: https://help.aliyun.com/api/dcdn/describedcdndomaincertificateinfo.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeDcdnDomainCertificateInfoWithCallback(request *DescribeDcdnDomainCertificateInfoRequest, callback func(response *DescribeDcdnDomainCertificateInfoResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeDcdnDomainCertificateInfoResponse
		var err error
		defer close(result)
		response, err = client.DescribeDcdnDomainCertificateInfo(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// DescribeDcdnDomainCertificateInfoRequest is the request struct for api DescribeDcdnDomainCertificateInfo
type DescribeDcdnDomainCertificateInfoRequest struct {
	*requests.RpcRequest
	DomainName string           `position:"Query" name:"DomainName"`
	OwnerId    requests.Integer `position:"Query" name:"OwnerId"`
}

// DescribeDcdnDomainCertificateInfoResponse is the response struct for api DescribeDcdnDomainCertificateInfo
type DescribeDcdnDomainCertificateInfoResponse struct {
	*responses.BaseResponse
	RequestId string    `json:"RequestId" xml:"RequestId"`
	CertInfos CertInfos `json:"CertInfos" xml:"CertInfos"`
}

// CreateDescribeDcdnDomainCertificateInfoRequest creates a request to invoke DescribeDcdnDomainCertificateInfo API
func CreateDescribeDcdnDomainCertificateInfoRequest() (request *DescribeDcdnDomainCertificateInfoRequest) {
	request = &DescribeDcdnDomainCertificateInfoRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("dcdn", "2018-01-15", "DescribeDcdnDomainCertificateInfo", "", "")
	return
}

// CreateDescribeDcdnDomainCertificateInfoResponse creates a response to parse from DescribeDcdnDomainCertificateInfo response
func CreateDescribeDcdnDomainCertificateInfoResponse() (response *DescribeDcdnDomainCertificateInfoResponse) {
	response = &DescribeDcdnDomainCertificateInfoResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
