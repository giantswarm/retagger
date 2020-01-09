package iot

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

// QueryEdgeInstanceHistoricDeployment invokes the iot.QueryEdgeInstanceHistoricDeployment API synchronously
// api document: https://help.aliyun.com/api/iot/queryedgeinstancehistoricdeployment.html
func (client *Client) QueryEdgeInstanceHistoricDeployment(request *QueryEdgeInstanceHistoricDeploymentRequest) (response *QueryEdgeInstanceHistoricDeploymentResponse, err error) {
	response = CreateQueryEdgeInstanceHistoricDeploymentResponse()
	err = client.DoAction(request, response)
	return
}

// QueryEdgeInstanceHistoricDeploymentWithChan invokes the iot.QueryEdgeInstanceHistoricDeployment API asynchronously
// api document: https://help.aliyun.com/api/iot/queryedgeinstancehistoricdeployment.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) QueryEdgeInstanceHistoricDeploymentWithChan(request *QueryEdgeInstanceHistoricDeploymentRequest) (<-chan *QueryEdgeInstanceHistoricDeploymentResponse, <-chan error) {
	responseChan := make(chan *QueryEdgeInstanceHistoricDeploymentResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.QueryEdgeInstanceHistoricDeployment(request)
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

// QueryEdgeInstanceHistoricDeploymentWithCallback invokes the iot.QueryEdgeInstanceHistoricDeployment API asynchronously
// api document: https://help.aliyun.com/api/iot/queryedgeinstancehistoricdeployment.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) QueryEdgeInstanceHistoricDeploymentWithCallback(request *QueryEdgeInstanceHistoricDeploymentRequest, callback func(response *QueryEdgeInstanceHistoricDeploymentResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *QueryEdgeInstanceHistoricDeploymentResponse
		var err error
		defer close(result)
		response, err = client.QueryEdgeInstanceHistoricDeployment(request)
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

// QueryEdgeInstanceHistoricDeploymentRequest is the request struct for api QueryEdgeInstanceHistoricDeployment
type QueryEdgeInstanceHistoricDeploymentRequest struct {
	*requests.RpcRequest
	EndTime       requests.Integer `position:"Query" name:"EndTime"`
	CurrentPage   requests.Integer `position:"Query" name:"CurrentPage"`
	StartTime     requests.Integer `position:"Query" name:"StartTime"`
	InstanceId    string           `position:"Query" name:"InstanceId"`
	IotInstanceId string           `position:"Query" name:"IotInstanceId"`
	PageSize      requests.Integer `position:"Query" name:"PageSize"`
}

// QueryEdgeInstanceHistoricDeploymentResponse is the response struct for api QueryEdgeInstanceHistoricDeployment
type QueryEdgeInstanceHistoricDeploymentResponse struct {
	*responses.BaseResponse
	RequestId    string                                    `json:"RequestId" xml:"RequestId"`
	Success      bool                                      `json:"Success" xml:"Success"`
	Code         string                                    `json:"Code" xml:"Code"`
	ErrorMessage string                                    `json:"ErrorMessage" xml:"ErrorMessage"`
	Data         DataInQueryEdgeInstanceHistoricDeployment `json:"Data" xml:"Data"`
}

// CreateQueryEdgeInstanceHistoricDeploymentRequest creates a request to invoke QueryEdgeInstanceHistoricDeployment API
func CreateQueryEdgeInstanceHistoricDeploymentRequest() (request *QueryEdgeInstanceHistoricDeploymentRequest) {
	request = &QueryEdgeInstanceHistoricDeploymentRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Iot", "2018-01-20", "QueryEdgeInstanceHistoricDeployment", "iot", "openAPI")
	return
}

// CreateQueryEdgeInstanceHistoricDeploymentResponse creates a response to parse from QueryEdgeInstanceHistoricDeployment response
func CreateQueryEdgeInstanceHistoricDeploymentResponse() (response *QueryEdgeInstanceHistoricDeploymentResponse) {
	response = &QueryEdgeInstanceHistoricDeploymentResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
