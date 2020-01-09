package green

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

// DescribeAuditSetting invokes the green.DescribeAuditSetting API synchronously
// api document: https://help.aliyun.com/api/green/describeauditsetting.html
func (client *Client) DescribeAuditSetting(request *DescribeAuditSettingRequest) (response *DescribeAuditSettingResponse, err error) {
	response = CreateDescribeAuditSettingResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeAuditSettingWithChan invokes the green.DescribeAuditSetting API asynchronously
// api document: https://help.aliyun.com/api/green/describeauditsetting.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeAuditSettingWithChan(request *DescribeAuditSettingRequest) (<-chan *DescribeAuditSettingResponse, <-chan error) {
	responseChan := make(chan *DescribeAuditSettingResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeAuditSetting(request)
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

// DescribeAuditSettingWithCallback invokes the green.DescribeAuditSetting API asynchronously
// api document: https://help.aliyun.com/api/green/describeauditsetting.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeAuditSettingWithCallback(request *DescribeAuditSettingRequest, callback func(response *DescribeAuditSettingResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeAuditSettingResponse
		var err error
		defer close(result)
		response, err = client.DescribeAuditSetting(request)
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

// DescribeAuditSettingRequest is the request struct for api DescribeAuditSetting
type DescribeAuditSettingRequest struct {
	*requests.RpcRequest
	SourceIp string `position:"Query" name:"SourceIp"`
	Lang     string `position:"Query" name:"Lang"`
}

// DescribeAuditSettingResponse is the response struct for api DescribeAuditSetting
type DescribeAuditSettingResponse struct {
	*responses.BaseResponse
	RequestId  string     `json:"RequestId" xml:"RequestId"`
	Seed       string     `json:"Seed" xml:"Seed"`
	Callback   string     `json:"Callback" xml:"Callback"`
	AuditRange AuditRange `json:"AuditRange" xml:"AuditRange"`
}

// CreateDescribeAuditSettingRequest creates a request to invoke DescribeAuditSetting API
func CreateDescribeAuditSettingRequest() (request *DescribeAuditSettingRequest) {
	request = &DescribeAuditSettingRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Green", "2017-08-23", "DescribeAuditSetting", "green", "openAPI")
	return
}

// CreateDescribeAuditSettingResponse creates a response to parse from DescribeAuditSetting response
func CreateDescribeAuditSettingResponse() (response *DescribeAuditSettingResponse) {
	response = &DescribeAuditSettingResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
