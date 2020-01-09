package rtc

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

// CreateChannel invokes the rtc.CreateChannel API synchronously
// api document: https://help.aliyun.com/api/rtc/createchannel.html
func (client *Client) CreateChannel(request *CreateChannelRequest) (response *CreateChannelResponse, err error) {
	response = CreateCreateChannelResponse()
	err = client.DoAction(request, response)
	return
}

// CreateChannelWithChan invokes the rtc.CreateChannel API asynchronously
// api document: https://help.aliyun.com/api/rtc/createchannel.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CreateChannelWithChan(request *CreateChannelRequest) (<-chan *CreateChannelResponse, <-chan error) {
	responseChan := make(chan *CreateChannelResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.CreateChannel(request)
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

// CreateChannelWithCallback invokes the rtc.CreateChannel API asynchronously
// api document: https://help.aliyun.com/api/rtc/createchannel.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CreateChannelWithCallback(request *CreateChannelRequest, callback func(response *CreateChannelResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *CreateChannelResponse
		var err error
		defer close(result)
		response, err = client.CreateChannel(request)
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

// CreateChannelRequest is the request struct for api CreateChannel
type CreateChannelRequest struct {
	*requests.RpcRequest
	OwnerId   requests.Integer `position:"Query" name:"OwnerId"`
	AppId     string           `position:"Query" name:"AppId"`
	ChannelId string           `position:"Query" name:"ChannelId"`
}

// CreateChannelResponse is the response struct for api CreateChannel
type CreateChannelResponse struct {
	*responses.BaseResponse
	RequestId  string `json:"RequestId" xml:"RequestId"`
	ChannelKey string `json:"ChannelKey" xml:"ChannelKey"`
	Nonce      string `json:"Nonce" xml:"Nonce"`
	Timestamp  int    `json:"Timestamp" xml:"Timestamp"`
}

// CreateCreateChannelRequest creates a request to invoke CreateChannel API
func CreateCreateChannelRequest() (request *CreateChannelRequest) {
	request = &CreateChannelRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("rtc", "2018-01-11", "CreateChannel", "", "")
	return
}

// CreateCreateChannelResponse creates a response to parse from CreateChannel response
func CreateCreateChannelResponse() (response *CreateChannelResponse) {
	response = &CreateChannelResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
