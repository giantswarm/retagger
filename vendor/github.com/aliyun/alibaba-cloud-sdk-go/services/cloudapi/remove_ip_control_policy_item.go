package cloudapi

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

// RemoveIpControlPolicyItem invokes the cloudapi.RemoveIpControlPolicyItem API synchronously
// api document: https://help.aliyun.com/api/cloudapi/removeipcontrolpolicyitem.html
func (client *Client) RemoveIpControlPolicyItem(request *RemoveIpControlPolicyItemRequest) (response *RemoveIpControlPolicyItemResponse, err error) {
	response = CreateRemoveIpControlPolicyItemResponse()
	err = client.DoAction(request, response)
	return
}

// RemoveIpControlPolicyItemWithChan invokes the cloudapi.RemoveIpControlPolicyItem API asynchronously
// api document: https://help.aliyun.com/api/cloudapi/removeipcontrolpolicyitem.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) RemoveIpControlPolicyItemWithChan(request *RemoveIpControlPolicyItemRequest) (<-chan *RemoveIpControlPolicyItemResponse, <-chan error) {
	responseChan := make(chan *RemoveIpControlPolicyItemResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.RemoveIpControlPolicyItem(request)
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

// RemoveIpControlPolicyItemWithCallback invokes the cloudapi.RemoveIpControlPolicyItem API asynchronously
// api document: https://help.aliyun.com/api/cloudapi/removeipcontrolpolicyitem.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) RemoveIpControlPolicyItemWithCallback(request *RemoveIpControlPolicyItemRequest, callback func(response *RemoveIpControlPolicyItemResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *RemoveIpControlPolicyItemResponse
		var err error
		defer close(result)
		response, err = client.RemoveIpControlPolicyItem(request)
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

// RemoveIpControlPolicyItemRequest is the request struct for api RemoveIpControlPolicyItem
type RemoveIpControlPolicyItemRequest struct {
	*requests.RpcRequest
	PolicyItemIds string `position:"Query" name:"PolicyItemIds"`
	IpControlId   string `position:"Query" name:"IpControlId"`
	SecurityToken string `position:"Query" name:"SecurityToken"`
}

// RemoveIpControlPolicyItemResponse is the response struct for api RemoveIpControlPolicyItem
type RemoveIpControlPolicyItemResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateRemoveIpControlPolicyItemRequest creates a request to invoke RemoveIpControlPolicyItem API
func CreateRemoveIpControlPolicyItemRequest() (request *RemoveIpControlPolicyItemRequest) {
	request = &RemoveIpControlPolicyItemRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("CloudAPI", "2016-07-14", "RemoveIpControlPolicyItem", "apigateway", "openAPI")
	return
}

// CreateRemoveIpControlPolicyItemResponse creates a response to parse from RemoveIpControlPolicyItem response
func CreateRemoveIpControlPolicyItemResponse() (response *RemoveIpControlPolicyItemResponse) {
	response = &RemoveIpControlPolicyItemResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
