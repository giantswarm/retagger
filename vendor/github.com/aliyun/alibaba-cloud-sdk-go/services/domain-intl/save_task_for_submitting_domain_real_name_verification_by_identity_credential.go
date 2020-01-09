package domain_intl

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

// SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential invokes the domain_intl.SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential API synchronously
// api document: https://help.aliyun.com/api/domain-intl/savetaskforsubmittingdomainrealnameverificationbyidentitycredential.html
func (client *Client) SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential(request *SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialRequest) (response *SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse, err error) {
	response = CreateSaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse()
	err = client.DoAction(request, response)
	return
}

// SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialWithChan invokes the domain_intl.SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential API asynchronously
// api document: https://help.aliyun.com/api/domain-intl/savetaskforsubmittingdomainrealnameverificationbyidentitycredential.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialWithChan(request *SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialRequest) (<-chan *SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse, <-chan error) {
	responseChan := make(chan *SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential(request)
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

// SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialWithCallback invokes the domain_intl.SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential API asynchronously
// api document: https://help.aliyun.com/api/domain-intl/savetaskforsubmittingdomainrealnameverificationbyidentitycredential.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialWithCallback(request *SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialRequest, callback func(response *SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse
		var err error
		defer close(result)
		response, err = client.SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential(request)
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

// SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialRequest is the request struct for api SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential
type SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialRequest struct {
	*requests.RpcRequest
	IdentityCredentialType string    `position:"Query" name:"IdentityCredentialType"`
	UserClientIp           string    `position:"Query" name:"UserClientIp"`
	IdentityCredential     string    `position:"Body" name:"IdentityCredential"`
	DomainName             *[]string `position:"Query" name:"DomainName"  type:"Repeated"`
	Lang                   string    `position:"Query" name:"Lang"`
	IdentityCredentialNo   string    `position:"Query" name:"IdentityCredentialNo"`
}

// SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse is the response struct for api SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential
type SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
	TaskNo    string `json:"TaskNo" xml:"TaskNo"`
}

// CreateSaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialRequest creates a request to invoke SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential API
func CreateSaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialRequest() (request *SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialRequest) {
	request = &SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Domain-intl", "2017-12-18", "SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential", "domain", "openAPI")
	return
}

// CreateSaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse creates a response to parse from SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredential response
func CreateSaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse() (response *SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse) {
	response = &SaveTaskForSubmittingDomainRealNameVerificationByIdentityCredentialResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
