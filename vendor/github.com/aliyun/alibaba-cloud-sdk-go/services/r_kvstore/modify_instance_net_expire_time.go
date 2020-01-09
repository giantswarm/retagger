package r_kvstore

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

// ModifyInstanceNetExpireTime invokes the r_kvstore.ModifyInstanceNetExpireTime API synchronously
// api document: https://help.aliyun.com/api/r-kvstore/modifyinstancenetexpiretime.html
func (client *Client) ModifyInstanceNetExpireTime(request *ModifyInstanceNetExpireTimeRequest) (response *ModifyInstanceNetExpireTimeResponse, err error) {
	response = CreateModifyInstanceNetExpireTimeResponse()
	err = client.DoAction(request, response)
	return
}

// ModifyInstanceNetExpireTimeWithChan invokes the r_kvstore.ModifyInstanceNetExpireTime API asynchronously
// api document: https://help.aliyun.com/api/r-kvstore/modifyinstancenetexpiretime.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ModifyInstanceNetExpireTimeWithChan(request *ModifyInstanceNetExpireTimeRequest) (<-chan *ModifyInstanceNetExpireTimeResponse, <-chan error) {
	responseChan := make(chan *ModifyInstanceNetExpireTimeResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ModifyInstanceNetExpireTime(request)
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

// ModifyInstanceNetExpireTimeWithCallback invokes the r_kvstore.ModifyInstanceNetExpireTime API asynchronously
// api document: https://help.aliyun.com/api/r-kvstore/modifyinstancenetexpiretime.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ModifyInstanceNetExpireTimeWithCallback(request *ModifyInstanceNetExpireTimeRequest, callback func(response *ModifyInstanceNetExpireTimeResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ModifyInstanceNetExpireTimeResponse
		var err error
		defer close(result)
		response, err = client.ModifyInstanceNetExpireTime(request)
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

// ModifyInstanceNetExpireTimeRequest is the request struct for api ModifyInstanceNetExpireTime
type ModifyInstanceNetExpireTimeRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ConnectionString     string           `position:"Query" name:"ConnectionString"`
	SecurityToken        string           `position:"Query" name:"SecurityToken"`
	ClassicExpiredDays   requests.Integer `position:"Query" name:"ClassicExpiredDays"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	InstanceId           string           `position:"Query" name:"InstanceId"`
}

// ModifyInstanceNetExpireTimeResponse is the response struct for api ModifyInstanceNetExpireTime
type ModifyInstanceNetExpireTimeResponse struct {
	*responses.BaseResponse
	RequestId    string                                    `json:"RequestId" xml:"RequestId"`
	InstanceId   string                                    `json:"InstanceId" xml:"InstanceId"`
	NetInfoItems NetInfoItemsInModifyInstanceNetExpireTime `json:"NetInfoItems" xml:"NetInfoItems"`
}

// CreateModifyInstanceNetExpireTimeRequest creates a request to invoke ModifyInstanceNetExpireTime API
func CreateModifyInstanceNetExpireTimeRequest() (request *ModifyInstanceNetExpireTimeRequest) {
	request = &ModifyInstanceNetExpireTimeRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("R-kvstore", "2015-01-01", "ModifyInstanceNetExpireTime", "", "")
	return
}

// CreateModifyInstanceNetExpireTimeResponse creates a response to parse from ModifyInstanceNetExpireTime response
func CreateModifyInstanceNetExpireTimeResponse() (response *ModifyInstanceNetExpireTimeResponse) {
	response = &ModifyInstanceNetExpireTimeResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
