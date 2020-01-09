package airec

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

// PushDocument invokes the airec.PushDocument API synchronously
// api document: https://help.aliyun.com/api/airec/pushdocument.html
func (client *Client) PushDocument(request *PushDocumentRequest) (response *PushDocumentResponse, err error) {
	response = CreatePushDocumentResponse()
	err = client.DoAction(request, response)
	return
}

// PushDocumentWithChan invokes the airec.PushDocument API asynchronously
// api document: https://help.aliyun.com/api/airec/pushdocument.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) PushDocumentWithChan(request *PushDocumentRequest) (<-chan *PushDocumentResponse, <-chan error) {
	responseChan := make(chan *PushDocumentResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.PushDocument(request)
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

// PushDocumentWithCallback invokes the airec.PushDocument API asynchronously
// api document: https://help.aliyun.com/api/airec/pushdocument.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) PushDocumentWithCallback(request *PushDocumentRequest, callback func(response *PushDocumentResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *PushDocumentResponse
		var err error
		defer close(result)
		response, err = client.PushDocument(request)
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

// PushDocumentRequest is the request struct for api PushDocument
type PushDocumentRequest struct {
	*requests.RoaRequest
	InstanceId string `position:"Path" name:"InstanceId"`
	TableName  string `position:"Path" name:"TableName"`
}

// PushDocumentResponse is the response struct for api PushDocument
type PushDocumentResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
	Code      string `json:"Code" xml:"Code"`
	Message   string `json:"Message" xml:"Message"`
	Result    bool   `json:"Result" xml:"Result"`
}

// CreatePushDocumentRequest creates a request to invoke PushDocument API
func CreatePushDocumentRequest() (request *PushDocumentRequest) {
	request = &PushDocumentRequest{
		RoaRequest: &requests.RoaRequest{},
	}
	request.InitWithApiInfo("Airec", "2018-10-12", "PushDocument", "/openapi/instances/[InstanceId]/tables/[TableName]/actions/bulk", "airec", "openAPI")
	request.Method = requests.POST
	return
}

// CreatePushDocumentResponse creates a response to parse from PushDocument response
func CreatePushDocumentResponse() (response *PushDocumentResponse) {
	response = &PushDocumentResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
