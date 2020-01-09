package ccc

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

// CancelJobs invokes the ccc.CancelJobs API synchronously
// api document: https://help.aliyun.com/api/ccc/canceljobs.html
func (client *Client) CancelJobs(request *CancelJobsRequest) (response *CancelJobsResponse, err error) {
	response = CreateCancelJobsResponse()
	err = client.DoAction(request, response)
	return
}

// CancelJobsWithChan invokes the ccc.CancelJobs API asynchronously
// api document: https://help.aliyun.com/api/ccc/canceljobs.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CancelJobsWithChan(request *CancelJobsRequest) (<-chan *CancelJobsResponse, <-chan error) {
	responseChan := make(chan *CancelJobsResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.CancelJobs(request)
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

// CancelJobsWithCallback invokes the ccc.CancelJobs API asynchronously
// api document: https://help.aliyun.com/api/ccc/canceljobs.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CancelJobsWithCallback(request *CancelJobsRequest, callback func(response *CancelJobsResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *CancelJobsResponse
		var err error
		defer close(result)
		response, err = client.CancelJobs(request)
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

// CancelJobsRequest is the request struct for api CancelJobs
type CancelJobsRequest struct {
	*requests.RpcRequest
	All            requests.Boolean `position:"Query" name:"All"`
	JobReferenceId *[]string        `position:"Query" name:"JobReferenceId"  type:"Repeated"`
	GroupId        string           `position:"Query" name:"GroupId"`
	JobId          *[]string        `position:"Query" name:"JobId"  type:"Repeated"`
	InstanceId     string           `position:"Query" name:"InstanceId"`
	ScenarioId     string           `position:"Query" name:"ScenarioId"`
}

// CancelJobsResponse is the response struct for api CancelJobs
type CancelJobsResponse struct {
	*responses.BaseResponse
	RequestId      string `json:"RequestId" xml:"RequestId"`
	Success        bool   `json:"Success" xml:"Success"`
	Code           string `json:"Code" xml:"Code"`
	Message        string `json:"Message" xml:"Message"`
	HttpStatusCode int    `json:"HttpStatusCode" xml:"HttpStatusCode"`
}

// CreateCancelJobsRequest creates a request to invoke CancelJobs API
func CreateCancelJobsRequest() (request *CancelJobsRequest) {
	request = &CancelJobsRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("CCC", "2017-07-05", "CancelJobs", "", "")
	return
}

// CreateCancelJobsResponse creates a response to parse from CancelJobs response
func CreateCancelJobsResponse() (response *CancelJobsResponse) {
	response = &CancelJobsResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
