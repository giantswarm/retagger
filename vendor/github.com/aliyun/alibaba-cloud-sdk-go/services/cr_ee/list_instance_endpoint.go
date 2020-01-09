package cr_ee

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

// ListInstanceEndpoint invokes the cr.ListInstanceEndpoint API synchronously
// api document: https://help.aliyun.com/api/cr/listinstanceendpoint.html
func (client *Client) ListInstanceEndpoint(request *ListInstanceEndpointRequest) (response *ListInstanceEndpointResponse, err error) {
	response = CreateListInstanceEndpointResponse()
	err = client.DoAction(request, response)
	return
}

// ListInstanceEndpointWithChan invokes the cr.ListInstanceEndpoint API asynchronously
// api document: https://help.aliyun.com/api/cr/listinstanceendpoint.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ListInstanceEndpointWithChan(request *ListInstanceEndpointRequest) (<-chan *ListInstanceEndpointResponse, <-chan error) {
	responseChan := make(chan *ListInstanceEndpointResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ListInstanceEndpoint(request)
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

// ListInstanceEndpointWithCallback invokes the cr.ListInstanceEndpoint API asynchronously
// api document: https://help.aliyun.com/api/cr/listinstanceendpoint.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ListInstanceEndpointWithCallback(request *ListInstanceEndpointRequest, callback func(response *ListInstanceEndpointResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ListInstanceEndpointResponse
		var err error
		defer close(result)
		response, err = client.ListInstanceEndpoint(request)
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

// ListInstanceEndpointRequest is the request struct for api ListInstanceEndpoint
type ListInstanceEndpointRequest struct {
	*requests.RpcRequest
	InstanceId string `position:"Query" name:"InstanceId"`
	ModuleName string `position:"Query" name:"ModuleName"`
}

// ListInstanceEndpointResponse is the response struct for api ListInstanceEndpoint
type ListInstanceEndpointResponse struct {
	*responses.BaseResponse
	ListInstanceEndpointIsSuccess bool            `json:"IsSuccess" xml:"IsSuccess"`
	Code                          string          `json:"Code" xml:"Code"`
	RequestId                     string          `json:"RequestId" xml:"RequestId"`
	Endpoints                     []EndpointsItem `json:"Endpoints" xml:"Endpoints"`
}

// CreateListInstanceEndpointRequest creates a request to invoke ListInstanceEndpoint API
func CreateListInstanceEndpointRequest() (request *ListInstanceEndpointRequest) {
	request = &ListInstanceEndpointRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("cr", "2018-12-01", "ListInstanceEndpoint", "cr", "openAPI")
	return
}

// CreateListInstanceEndpointResponse creates a response to parse from ListInstanceEndpoint response
func CreateListInstanceEndpointResponse() (response *ListInstanceEndpointResponse) {
	response = &ListInstanceEndpointResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
