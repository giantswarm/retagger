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

// RunCloudMetricProfiling invokes the ehpc.RunCloudMetricProfiling API synchronously
// api document: https://help.aliyun.com/api/ehpc/runcloudmetricprofiling.html
func (client *Client) RunCloudMetricProfiling(request *RunCloudMetricProfilingRequest) (response *RunCloudMetricProfilingResponse, err error) {
	response = CreateRunCloudMetricProfilingResponse()
	err = client.DoAction(request, response)
	return
}

// RunCloudMetricProfilingWithChan invokes the ehpc.RunCloudMetricProfiling API asynchronously
// api document: https://help.aliyun.com/api/ehpc/runcloudmetricprofiling.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) RunCloudMetricProfilingWithChan(request *RunCloudMetricProfilingRequest) (<-chan *RunCloudMetricProfilingResponse, <-chan error) {
	responseChan := make(chan *RunCloudMetricProfilingResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.RunCloudMetricProfiling(request)
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

// RunCloudMetricProfilingWithCallback invokes the ehpc.RunCloudMetricProfiling API asynchronously
// api document: https://help.aliyun.com/api/ehpc/runcloudmetricprofiling.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) RunCloudMetricProfilingWithCallback(request *RunCloudMetricProfilingRequest, callback func(response *RunCloudMetricProfilingResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *RunCloudMetricProfilingResponse
		var err error
		defer close(result)
		response, err = client.RunCloudMetricProfiling(request)
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

// RunCloudMetricProfilingRequest is the request struct for api RunCloudMetricProfiling
type RunCloudMetricProfilingRequest struct {
	*requests.RpcRequest
	Freq      requests.Integer `position:"Query" name:"Freq"`
	ClusterId string           `position:"Query" name:"ClusterId"`
	Duration  requests.Integer `position:"Query" name:"Duration"`
	HostName  string           `position:"Query" name:"HostName"`
	ProcessId requests.Integer `position:"Query" name:"ProcessId"`
}

// RunCloudMetricProfilingResponse is the response struct for api RunCloudMetricProfiling
type RunCloudMetricProfilingResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateRunCloudMetricProfilingRequest creates a request to invoke RunCloudMetricProfiling API
func CreateRunCloudMetricProfilingRequest() (request *RunCloudMetricProfilingRequest) {
	request = &RunCloudMetricProfilingRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("EHPC", "2018-04-12", "RunCloudMetricProfiling", "ehs", "openAPI")
	return
}

// CreateRunCloudMetricProfilingResponse creates a response to parse from RunCloudMetricProfiling response
func CreateRunCloudMetricProfilingResponse() (response *RunCloudMetricProfilingResponse) {
	response = &RunCloudMetricProfilingResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
