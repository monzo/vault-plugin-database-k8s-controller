package kms

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

// DescribeCertificate invokes the kms.DescribeCertificate API synchronously
func (client *Client) DescribeCertificate(request *DescribeCertificateRequest) (response *DescribeCertificateResponse, err error) {
	response = CreateDescribeCertificateResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeCertificateWithChan invokes the kms.DescribeCertificate API asynchronously
func (client *Client) DescribeCertificateWithChan(request *DescribeCertificateRequest) (<-chan *DescribeCertificateResponse, <-chan error) {
	responseChan := make(chan *DescribeCertificateResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeCertificate(request)
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

// DescribeCertificateWithCallback invokes the kms.DescribeCertificate API asynchronously
func (client *Client) DescribeCertificateWithCallback(request *DescribeCertificateRequest, callback func(response *DescribeCertificateResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeCertificateResponse
		var err error
		defer close(result)
		response, err = client.DescribeCertificate(request)
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

// DescribeCertificateRequest is the request struct for api DescribeCertificate
type DescribeCertificateRequest struct {
	*requests.RpcRequest
	CertificateId string `position:"Query" name:"CertificateId"`
}

// DescribeCertificateResponse is the response struct for api DescribeCertificate
type DescribeCertificateResponse struct {
	*responses.BaseResponse
	Status                  string                 `json:"Status" xml:"Status"`
	Serial                  string                 `json:"Serial" xml:"Serial"`
	CreatedAt               string                 `json:"CreatedAt" xml:"CreatedAt"`
	Tags                    map[string]interface{} `json:"Tags" xml:"Tags"`
	SubjectKeyIdentifier    string                 `json:"SubjectKeyIdentifier" xml:"SubjectKeyIdentifier"`
	ExportablePrivateKey    bool                   `json:"ExportablePrivateKey" xml:"ExportablePrivateKey"`
	Issuer                  string                 `json:"Issuer" xml:"Issuer"`
	SignatureAlgorithm      string                 `json:"SignatureAlgorithm" xml:"SignatureAlgorithm"`
	NotAfter                string                 `json:"NotAfter" xml:"NotAfter"`
	Arn                     string                 `json:"Arn" xml:"Arn"`
	CertificateId           string                 `json:"CertificateId" xml:"CertificateId"`
	RequestId               string                 `json:"RequestId" xml:"RequestId"`
	KeySpec                 string                 `json:"KeySpec" xml:"KeySpec"`
	ProtectionLevel         string                 `json:"ProtectionLevel" xml:"ProtectionLevel"`
	SubjectPublicKey        string                 `json:"SubjectPublicKey" xml:"SubjectPublicKey"`
	Subject                 string                 `json:"Subject" xml:"Subject"`
	NotBefore               string                 `json:"NotBefore" xml:"NotBefore"`
	UpdatedAt               string                 `json:"UpdatedAt" xml:"UpdatedAt"`
	SubjectAlternativeNames []string               `json:"SubjectAlternativeNames" xml:"SubjectAlternativeNames"`
}

// CreateDescribeCertificateRequest creates a request to invoke DescribeCertificate API
func CreateDescribeCertificateRequest() (request *DescribeCertificateRequest) {
	request = &DescribeCertificateRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Kms", "2016-01-20", "DescribeCertificate", "kms", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDescribeCertificateResponse creates a response to parse from DescribeCertificate response
func CreateDescribeCertificateResponse() (response *DescribeCertificateResponse) {
	response = &DescribeCertificateResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
