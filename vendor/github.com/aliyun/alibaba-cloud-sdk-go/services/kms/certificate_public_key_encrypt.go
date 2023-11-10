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

// CertificatePublicKeyEncrypt invokes the kms.CertificatePublicKeyEncrypt API synchronously
func (client *Client) CertificatePublicKeyEncrypt(request *CertificatePublicKeyEncryptRequest) (response *CertificatePublicKeyEncryptResponse, err error) {
	response = CreateCertificatePublicKeyEncryptResponse()
	err = client.DoAction(request, response)
	return
}

// CertificatePublicKeyEncryptWithChan invokes the kms.CertificatePublicKeyEncrypt API asynchronously
func (client *Client) CertificatePublicKeyEncryptWithChan(request *CertificatePublicKeyEncryptRequest) (<-chan *CertificatePublicKeyEncryptResponse, <-chan error) {
	responseChan := make(chan *CertificatePublicKeyEncryptResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.CertificatePublicKeyEncrypt(request)
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

// CertificatePublicKeyEncryptWithCallback invokes the kms.CertificatePublicKeyEncrypt API asynchronously
func (client *Client) CertificatePublicKeyEncryptWithCallback(request *CertificatePublicKeyEncryptRequest, callback func(response *CertificatePublicKeyEncryptResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *CertificatePublicKeyEncryptResponse
		var err error
		defer close(result)
		response, err = client.CertificatePublicKeyEncrypt(request)
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

// CertificatePublicKeyEncryptRequest is the request struct for api CertificatePublicKeyEncrypt
type CertificatePublicKeyEncryptRequest struct {
	*requests.RpcRequest
	CertificateId string `position:"Query" name:"CertificateId"`
	Plaintext     string `position:"Query" name:"Plaintext"`
	Algorithm     string `position:"Query" name:"Algorithm"`
}

// CertificatePublicKeyEncryptResponse is the response struct for api CertificatePublicKeyEncrypt
type CertificatePublicKeyEncryptResponse struct {
	*responses.BaseResponse
	CiphertextBlob string `json:"CiphertextBlob" xml:"CiphertextBlob"`
	RequestId      string `json:"RequestId" xml:"RequestId"`
	CertificateId  string `json:"CertificateId" xml:"CertificateId"`
}

// CreateCertificatePublicKeyEncryptRequest creates a request to invoke CertificatePublicKeyEncrypt API
func CreateCertificatePublicKeyEncryptRequest() (request *CertificatePublicKeyEncryptRequest) {
	request = &CertificatePublicKeyEncryptRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Kms", "2016-01-20", "CertificatePublicKeyEncrypt", "kms", "openAPI")
	request.Method = requests.POST
	return
}

// CreateCertificatePublicKeyEncryptResponse creates a response to parse from CertificatePublicKeyEncrypt response
func CreateCertificatePublicKeyEncryptResponse() (response *CertificatePublicKeyEncryptResponse) {
	response = &CertificatePublicKeyEncryptResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
