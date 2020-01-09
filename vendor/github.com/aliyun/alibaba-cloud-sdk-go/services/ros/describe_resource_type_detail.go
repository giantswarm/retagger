package ros

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

// DescribeResourceTypeDetail invokes the ros.DescribeResourceTypeDetail API synchronously
// api document: https://help.aliyun.com/api/ros/describeresourcetypedetail.html
func (client *Client) DescribeResourceTypeDetail(request *DescribeResourceTypeDetailRequest) (response *DescribeResourceTypeDetailResponse, err error) {
	response = CreateDescribeResourceTypeDetailResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeResourceTypeDetailWithChan invokes the ros.DescribeResourceTypeDetail API asynchronously
// api document: https://help.aliyun.com/api/ros/describeresourcetypedetail.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeResourceTypeDetailWithChan(request *DescribeResourceTypeDetailRequest) (<-chan *DescribeResourceTypeDetailResponse, <-chan error) {
	responseChan := make(chan *DescribeResourceTypeDetailResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeResourceTypeDetail(request)
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

// DescribeResourceTypeDetailWithCallback invokes the ros.DescribeResourceTypeDetail API asynchronously
// api document: https://help.aliyun.com/api/ros/describeresourcetypedetail.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeResourceTypeDetailWithCallback(request *DescribeResourceTypeDetailRequest, callback func(response *DescribeResourceTypeDetailResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeResourceTypeDetailResponse
		var err error
		defer close(result)
		response, err = client.DescribeResourceTypeDetail(request)
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

// DescribeResourceTypeDetailRequest is the request struct for api DescribeResourceTypeDetail
type DescribeResourceTypeDetailRequest struct {
	*requests.RoaRequest
	TypeName string `position:"Path" name:"TypeName"`
}

// DescribeResourceTypeDetailResponse is the response struct for api DescribeResourceTypeDetail
type DescribeResourceTypeDetailResponse struct {
	*responses.BaseResponse
}

// CreateDescribeResourceTypeDetailRequest creates a request to invoke DescribeResourceTypeDetail API
func CreateDescribeResourceTypeDetailRequest() (request *DescribeResourceTypeDetailRequest) {
	request = &DescribeResourceTypeDetailRequest{
		RoaRequest: &requests.RoaRequest{},
	}
	request.InitWithApiInfo("ROS", "2015-09-01", "DescribeResourceTypeDetail", "/resource_types/[TypeName]", "ROS", "openAPI")
	request.Method = requests.GET
	return
}

// CreateDescribeResourceTypeDetailResponse creates a response to parse from DescribeResourceTypeDetail response
func CreateDescribeResourceTypeDetailResponse() (response *DescribeResourceTypeDetailResponse) {
	response = &DescribeResourceTypeDetailResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
