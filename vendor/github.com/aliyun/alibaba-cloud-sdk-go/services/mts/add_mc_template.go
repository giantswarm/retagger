package mts

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

// AddMCTemplate invokes the mts.AddMCTemplate API synchronously
// api document: https://help.aliyun.com/api/mts/addmctemplate.html
func (client *Client) AddMCTemplate(request *AddMCTemplateRequest) (response *AddMCTemplateResponse, err error) {
	response = CreateAddMCTemplateResponse()
	err = client.DoAction(request, response)
	return
}

// AddMCTemplateWithChan invokes the mts.AddMCTemplate API asynchronously
// api document: https://help.aliyun.com/api/mts/addmctemplate.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) AddMCTemplateWithChan(request *AddMCTemplateRequest) (<-chan *AddMCTemplateResponse, <-chan error) {
	responseChan := make(chan *AddMCTemplateResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.AddMCTemplate(request)
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

// AddMCTemplateWithCallback invokes the mts.AddMCTemplate API asynchronously
// api document: https://help.aliyun.com/api/mts/addmctemplate.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) AddMCTemplateWithCallback(request *AddMCTemplateRequest, callback func(response *AddMCTemplateResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *AddMCTemplateResponse
		var err error
		defer close(result)
		response, err = client.AddMCTemplate(request)
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

// AddMCTemplateRequest is the request struct for api AddMCTemplate
type AddMCTemplateRequest struct {
	*requests.RpcRequest
	Politics             string           `position:"Query" name:"Politics"`
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	Contraband           string           `position:"Query" name:"Contraband"`
	Ad                   string           `position:"Query" name:"Ad"`
	Abuse                string           `position:"Query" name:"Abuse"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	Qrcode               string           `position:"Query" name:"Qrcode"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	Porn                 string           `position:"Query" name:"Porn"`
	Terrorism            string           `position:"Query" name:"Terrorism"`
	Name                 string           `position:"Query" name:"Name"`
	Logo                 string           `position:"Query" name:"Logo"`
	Spam                 string           `position:"Query" name:"spam"`
	Live                 string           `position:"Query" name:"Live"`
}

// AddMCTemplateResponse is the response struct for api AddMCTemplate
type AddMCTemplateResponse struct {
	*responses.BaseResponse
	RequestId string   `json:"RequestId" xml:"RequestId"`
	Template  Template `json:"Template" xml:"Template"`
}

// CreateAddMCTemplateRequest creates a request to invoke AddMCTemplate API
func CreateAddMCTemplateRequest() (request *AddMCTemplateRequest) {
	request = &AddMCTemplateRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Mts", "2014-06-18", "AddMCTemplate", "mts", "openAPI")
	return
}

// CreateAddMCTemplateResponse creates a response to parse from AddMCTemplate response
func CreateAddMCTemplateResponse() (response *AddMCTemplateResponse) {
	response = &AddMCTemplateResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
