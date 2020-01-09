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

// PullImage invokes the ehpc.PullImage API synchronously
// api document: https://help.aliyun.com/api/ehpc/pullimage.html
func (client *Client) PullImage(request *PullImageRequest) (response *PullImageResponse, err error) {
	response = CreatePullImageResponse()
	err = client.DoAction(request, response)
	return
}

// PullImageWithChan invokes the ehpc.PullImage API asynchronously
// api document: https://help.aliyun.com/api/ehpc/pullimage.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) PullImageWithChan(request *PullImageRequest) (<-chan *PullImageResponse, <-chan error) {
	responseChan := make(chan *PullImageResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.PullImage(request)
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

// PullImageWithCallback invokes the ehpc.PullImage API asynchronously
// api document: https://help.aliyun.com/api/ehpc/pullimage.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) PullImageWithCallback(request *PullImageRequest, callback func(response *PullImageResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *PullImageResponse
		var err error
		defer close(result)
		response, err = client.PullImage(request)
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

// PullImageRequest is the request struct for api PullImage
type PullImageRequest struct {
	*requests.RpcRequest
	ClusterId     string `position:"Query" name:"ClusterId"`
	Repository    string `position:"Query" name:"Repository"`
	ContainerType string `position:"Query" name:"ContainerType"`
	ImageTag      string `position:"Query" name:"ImageTag"`
}

// PullImageResponse is the response struct for api PullImage
type PullImageResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreatePullImageRequest creates a request to invoke PullImage API
func CreatePullImageRequest() (request *PullImageRequest) {
	request = &PullImageRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("EHPC", "2018-04-12", "PullImage", "ehs", "openAPI")
	return
}

// CreatePullImageResponse creates a response to parse from PullImage response
func CreatePullImageResponse() (response *PullImageResponse) {
	response = &PullImageResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
