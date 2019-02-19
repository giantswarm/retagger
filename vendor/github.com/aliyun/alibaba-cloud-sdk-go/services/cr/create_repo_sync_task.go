package cr

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

// CreateRepoSyncTask invokes the cr.CreateRepoSyncTask API synchronously
// api document: https://help.aliyun.com/api/cr/createreposynctask.html
func (client *Client) CreateRepoSyncTask(request *CreateRepoSyncTaskRequest) (response *CreateRepoSyncTaskResponse, err error) {
	response = CreateCreateRepoSyncTaskResponse()
	err = client.DoAction(request, response)
	return
}

// CreateRepoSyncTaskWithChan invokes the cr.CreateRepoSyncTask API asynchronously
// api document: https://help.aliyun.com/api/cr/createreposynctask.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CreateRepoSyncTaskWithChan(request *CreateRepoSyncTaskRequest) (<-chan *CreateRepoSyncTaskResponse, <-chan error) {
	responseChan := make(chan *CreateRepoSyncTaskResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.CreateRepoSyncTask(request)
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

// CreateRepoSyncTaskWithCallback invokes the cr.CreateRepoSyncTask API asynchronously
// api document: https://help.aliyun.com/api/cr/createreposynctask.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CreateRepoSyncTaskWithCallback(request *CreateRepoSyncTaskRequest, callback func(response *CreateRepoSyncTaskResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *CreateRepoSyncTaskResponse
		var err error
		defer close(result)
		response, err = client.CreateRepoSyncTask(request)
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

// CreateRepoSyncTaskRequest is the request struct for api CreateRepoSyncTask
type CreateRepoSyncTaskRequest struct {
	*requests.RoaRequest
	RepoNamespace string `position:"Path" name:"RepoNamespace"`
	RepoName      string `position:"Path" name:"RepoName"`
}

// CreateRepoSyncTaskResponse is the response struct for api CreateRepoSyncTask
type CreateRepoSyncTaskResponse struct {
	*responses.BaseResponse
}

// CreateCreateRepoSyncTaskRequest creates a request to invoke CreateRepoSyncTask API
func CreateCreateRepoSyncTaskRequest() (request *CreateRepoSyncTaskRequest) {
	request = &CreateRepoSyncTaskRequest{
		RoaRequest: &requests.RoaRequest{},
	}
	request.InitWithApiInfo("cr", "2016-06-07", "CreateRepoSyncTask", "/repos/[RepoNamespace]/[RepoName]/syncTasks", "", "")
	request.Method = requests.PUT
	return
}

// CreateCreateRepoSyncTaskResponse creates a response to parse from CreateRepoSyncTask response
func CreateCreateRepoSyncTaskResponse() (response *CreateRepoSyncTaskResponse) {
	response = &CreateRepoSyncTaskResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
