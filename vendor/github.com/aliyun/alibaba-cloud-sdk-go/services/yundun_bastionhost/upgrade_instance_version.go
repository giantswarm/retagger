package yundun_bastionhost

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

// UpgradeInstanceVersion invokes the yundun_bastionhost.UpgradeInstanceVersion API synchronously
// api document: https://help.aliyun.com/api/yundun-bastionhost/upgradeinstanceversion.html
func (client *Client) UpgradeInstanceVersion(request *UpgradeInstanceVersionRequest) (response *UpgradeInstanceVersionResponse, err error) {
	response = CreateUpgradeInstanceVersionResponse()
	err = client.DoAction(request, response)
	return
}

// UpgradeInstanceVersionWithChan invokes the yundun_bastionhost.UpgradeInstanceVersion API asynchronously
// api document: https://help.aliyun.com/api/yundun-bastionhost/upgradeinstanceversion.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) UpgradeInstanceVersionWithChan(request *UpgradeInstanceVersionRequest) (<-chan *UpgradeInstanceVersionResponse, <-chan error) {
	responseChan := make(chan *UpgradeInstanceVersionResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.UpgradeInstanceVersion(request)
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

// UpgradeInstanceVersionWithCallback invokes the yundun_bastionhost.UpgradeInstanceVersion API asynchronously
// api document: https://help.aliyun.com/api/yundun-bastionhost/upgradeinstanceversion.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) UpgradeInstanceVersionWithCallback(request *UpgradeInstanceVersionRequest, callback func(response *UpgradeInstanceVersionResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *UpgradeInstanceVersionResponse
		var err error
		defer close(result)
		response, err = client.UpgradeInstanceVersion(request)
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

// UpgradeInstanceVersionRequest is the request struct for api UpgradeInstanceVersion
type UpgradeInstanceVersionRequest struct {
	*requests.RpcRequest
	InstanceId string `position:"Query" name:"InstanceId"`
	SourceIp   string `position:"Query" name:"SourceIp"`
	Lang       string `position:"Query" name:"Lang"`
}

// UpgradeInstanceVersionResponse is the response struct for api UpgradeInstanceVersion
type UpgradeInstanceVersionResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateUpgradeInstanceVersionRequest creates a request to invoke UpgradeInstanceVersion API
func CreateUpgradeInstanceVersionRequest() (request *UpgradeInstanceVersionRequest) {
	request = &UpgradeInstanceVersionRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Yundun-bastionhost", "2018-10-10", "UpgradeInstanceVersion", "bastionhost", "openAPI")
	return
}

// CreateUpgradeInstanceVersionResponse creates a response to parse from UpgradeInstanceVersion response
func CreateUpgradeInstanceVersionResponse() (response *UpgradeInstanceVersionResponse) {
	response = &UpgradeInstanceVersionResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
