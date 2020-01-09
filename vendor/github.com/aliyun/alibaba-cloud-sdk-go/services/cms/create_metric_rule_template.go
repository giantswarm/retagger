package cms

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

// CreateMetricRuleTemplate invokes the cms.CreateMetricRuleTemplate API synchronously
// api document: https://help.aliyun.com/api/cms/createmetricruletemplate.html
func (client *Client) CreateMetricRuleTemplate(request *CreateMetricRuleTemplateRequest) (response *CreateMetricRuleTemplateResponse, err error) {
	response = CreateCreateMetricRuleTemplateResponse()
	err = client.DoAction(request, response)
	return
}

// CreateMetricRuleTemplateWithChan invokes the cms.CreateMetricRuleTemplate API asynchronously
// api document: https://help.aliyun.com/api/cms/createmetricruletemplate.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CreateMetricRuleTemplateWithChan(request *CreateMetricRuleTemplateRequest) (<-chan *CreateMetricRuleTemplateResponse, <-chan error) {
	responseChan := make(chan *CreateMetricRuleTemplateResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.CreateMetricRuleTemplate(request)
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

// CreateMetricRuleTemplateWithCallback invokes the cms.CreateMetricRuleTemplate API asynchronously
// api document: https://help.aliyun.com/api/cms/createmetricruletemplate.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CreateMetricRuleTemplateWithCallback(request *CreateMetricRuleTemplateRequest, callback func(response *CreateMetricRuleTemplateResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *CreateMetricRuleTemplateResponse
		var err error
		defer close(result)
		response, err = client.CreateMetricRuleTemplate(request)
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

// CreateMetricRuleTemplateRequest is the request struct for api CreateMetricRuleTemplate
type CreateMetricRuleTemplateRequest struct {
	*requests.RpcRequest
	Description    string                                    `position:"Query" name:"Description"`
	Name           string                                    `position:"Query" name:"Name"`
	AlertTemplates *[]CreateMetricRuleTemplateAlertTemplates `position:"Query" name:"AlertTemplates"  type:"Repeated"`
}

// CreateMetricRuleTemplateAlertTemplates is a repeated param struct in CreateMetricRuleTemplateRequest
type CreateMetricRuleTemplateAlertTemplates struct {
	Period                                string `name:"Period"`
	EscalationsWarnThreshold              string `name:"Escalations.Warn.Threshold"`
	Webhook                               string `name:"Webhook"`
	EscalationsWarnComparisonOperator     string `name:"Escalations.Warn.ComparisonOperator"`
	EscalationsCriticalStatistics         string `name:"Escalations.Critical.Statistics"`
	EscalationsInfoTimes                  string `name:"Escalations.Info.Times"`
	RuleName                              string `name:"RuleName"`
	EscalationsInfoStatistics             string `name:"Escalations.Info.Statistics"`
	EscalationsCriticalTimes              string `name:"Escalations.Critical.Times"`
	EscalationsInfoComparisonOperator     string `name:"Escalations.Info.ComparisonOperator"`
	EscalationsWarnStatistics             string `name:"Escalations.Warn.Statistics"`
	EscalationsInfoThreshold              string `name:"Escalations.Info.Threshold"`
	Namespace                             string `name:"Namespace"`
	Selector                              string `name:"Selector"`
	MetricName                            string `name:"MetricName"`
	Category                              string `name:"Category"`
	EscalationsCriticalComparisonOperator string `name:"Escalations.Critical.ComparisonOperator"`
	EscalationsWarnTimes                  string `name:"Escalations.Warn.Times"`
	EscalationsCriticalThreshold          string `name:"Escalations.Critical.Threshold"`
}

// CreateMetricRuleTemplateResponse is the response struct for api CreateMetricRuleTemplate
type CreateMetricRuleTemplateResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
	Success   bool   `json:"Success" xml:"Success"`
	Code      int    `json:"Code" xml:"Code"`
	Message   string `json:"Message" xml:"Message"`
	Id        int64  `json:"Id" xml:"Id"`
}

// CreateCreateMetricRuleTemplateRequest creates a request to invoke CreateMetricRuleTemplate API
func CreateCreateMetricRuleTemplateRequest() (request *CreateMetricRuleTemplateRequest) {
	request = &CreateMetricRuleTemplateRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Cms", "2019-01-01", "CreateMetricRuleTemplate", "cms", "openAPI")
	return
}

// CreateCreateMetricRuleTemplateResponse creates a response to parse from CreateMetricRuleTemplate response
func CreateCreateMetricRuleTemplateResponse() (response *CreateMetricRuleTemplateResponse) {
	response = &CreateMetricRuleTemplateResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
