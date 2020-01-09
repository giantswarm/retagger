package smartag

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

// UnbindSmartAccessGateway invokes the smartag.UnbindSmartAccessGateway API synchronously
// api document: https://help.aliyun.com/api/smartag/unbindsmartaccessgateway.html
func (client *Client) UnbindSmartAccessGateway(request *UnbindSmartAccessGatewayRequest) (response *UnbindSmartAccessGatewayResponse, err error) {
	response = CreateUnbindSmartAccessGatewayResponse()
	err = client.DoAction(request, response)
	return
}

// UnbindSmartAccessGatewayWithChan invokes the smartag.UnbindSmartAccessGateway API asynchronously
// api document: https://help.aliyun.com/api/smartag/unbindsmartaccessgateway.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) UnbindSmartAccessGatewayWithChan(request *UnbindSmartAccessGatewayRequest) (<-chan *UnbindSmartAccessGatewayResponse, <-chan error) {
	responseChan := make(chan *UnbindSmartAccessGatewayResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.UnbindSmartAccessGateway(request)
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

// UnbindSmartAccessGatewayWithCallback invokes the smartag.UnbindSmartAccessGateway API asynchronously
// api document: https://help.aliyun.com/api/smartag/unbindsmartaccessgateway.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) UnbindSmartAccessGatewayWithCallback(request *UnbindSmartAccessGatewayRequest, callback func(response *UnbindSmartAccessGatewayResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *UnbindSmartAccessGatewayResponse
		var err error
		defer close(result)
		response, err = client.UnbindSmartAccessGateway(request)
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

// UnbindSmartAccessGatewayRequest is the request struct for api UnbindSmartAccessGateway
type UnbindSmartAccessGatewayRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	CcnId                string           `position:"Query" name:"CcnId"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	SmartAGUid           requests.Integer `position:"Query" name:"SmartAGUid"`
	SmartAGId            string           `position:"Query" name:"SmartAGId"`
}

// UnbindSmartAccessGatewayResponse is the response struct for api UnbindSmartAccessGateway
type UnbindSmartAccessGatewayResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateUnbindSmartAccessGatewayRequest creates a request to invoke UnbindSmartAccessGateway API
func CreateUnbindSmartAccessGatewayRequest() (request *UnbindSmartAccessGatewayRequest) {
	request = &UnbindSmartAccessGatewayRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Smartag", "2018-03-13", "UnbindSmartAccessGateway", "smartag", "openAPI")
	return
}

// CreateUnbindSmartAccessGatewayResponse creates a response to parse from UnbindSmartAccessGateway response
func CreateUnbindSmartAccessGatewayResponse() (response *UnbindSmartAccessGatewayResponse) {
	response = &UnbindSmartAccessGatewayResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
