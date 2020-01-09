package ehpc

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

// DescribeNFSClientStatus invokes the ehpc.DescribeNFSClientStatus API synchronously
// api document: https://help.aliyun.com/api/ehpc/describenfsclientstatus.html
func (client *Client) DescribeNFSClientStatus(request *DescribeNFSClientStatusRequest) (response *DescribeNFSClientStatusResponse, err error) {
	response = CreateDescribeNFSClientStatusResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeNFSClientStatusWithChan invokes the ehpc.DescribeNFSClientStatus API asynchronously
// api document: https://help.aliyun.com/api/ehpc/describenfsclientstatus.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeNFSClientStatusWithChan(request *DescribeNFSClientStatusRequest) (<-chan *DescribeNFSClientStatusResponse, <-chan error) {
	responseChan := make(chan *DescribeNFSClientStatusResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeNFSClientStatus(request)
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

// DescribeNFSClientStatusWithCallback invokes the ehpc.DescribeNFSClientStatus API asynchronously
// api document: https://help.aliyun.com/api/ehpc/describenfsclientstatus.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeNFSClientStatusWithCallback(request *DescribeNFSClientStatusRequest, callback func(response *DescribeNFSClientStatusResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeNFSClientStatusResponse
		var err error
		defer close(result)
		response, err = client.DescribeNFSClientStatus(request)
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

// DescribeNFSClientStatusRequest is the request struct for api DescribeNFSClientStatus
type DescribeNFSClientStatusRequest struct {
	*requests.RpcRequest
	InstanceId string `position:"Query" name:"InstanceId"`
}

// DescribeNFSClientStatusResponse is the response struct for api DescribeNFSClientStatus
type DescribeNFSClientStatusResponse struct {
	*responses.BaseResponse
	Status string `json:"Status" xml:"Status"`
	Result Result `json:"Result" xml:"Result"`
}

// CreateDescribeNFSClientStatusRequest creates a request to invoke DescribeNFSClientStatus API
func CreateDescribeNFSClientStatusRequest() (request *DescribeNFSClientStatusRequest) {
	request = &DescribeNFSClientStatusRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("EHPC", "2018-04-12", "DescribeNFSClientStatus", "ehs", "openAPI")
	return
}

// CreateDescribeNFSClientStatusResponse creates a response to parse from DescribeNFSClientStatus response
func CreateDescribeNFSClientStatusResponse() (response *DescribeNFSClientStatusResponse) {
	response = &DescribeNFSClientStatusResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
