package csb

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

// GetService invokes the csb.GetService API synchronously
// api document: https://help.aliyun.com/api/csb/getservice.html
func (client *Client) GetService(request *GetServiceRequest) (response *GetServiceResponse, err error) {
	response = CreateGetServiceResponse()
	err = client.DoAction(request, response)
	return
}

// GetServiceWithChan invokes the csb.GetService API asynchronously
// api document: https://help.aliyun.com/api/csb/getservice.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetServiceWithChan(request *GetServiceRequest) (<-chan *GetServiceResponse, <-chan error) {
	responseChan := make(chan *GetServiceResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.GetService(request)
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

// GetServiceWithCallback invokes the csb.GetService API asynchronously
// api document: https://help.aliyun.com/api/csb/getservice.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetServiceWithCallback(request *GetServiceRequest, callback func(response *GetServiceResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *GetServiceResponse
		var err error
		defer close(result)
		response, err = client.GetService(request)
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

// GetServiceRequest is the request struct for api GetService
type GetServiceRequest struct {
	*requests.RpcRequest
	CsbId     requests.Integer `position:"Query" name:"CsbId"`
	ServiceId requests.Integer `position:"Query" name:"ServiceId"`
}

// GetServiceResponse is the response struct for api GetService
type GetServiceResponse struct {
	*responses.BaseResponse
	Code      int    `json:"Code" xml:"Code"`
	Message   string `json:"Message" xml:"Message"`
	RequestId string `json:"RequestId" xml:"RequestId"`
	Data      Data   `json:"Data" xml:"Data"`
}

// CreateGetServiceRequest creates a request to invoke GetService API
func CreateGetServiceRequest() (request *GetServiceRequest) {
	request = &GetServiceRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("CSB", "2017-11-18", "GetService", "", "")
	return
}

// CreateGetServiceResponse creates a response to parse from GetService response
func CreateGetServiceResponse() (response *GetServiceResponse) {
	response = &GetServiceResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
