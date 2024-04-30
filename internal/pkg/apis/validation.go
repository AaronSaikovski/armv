/*
MIT License

# Copyright (c) 2024 Aaron Saikovski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/AaronSaikovski/armv/pkg/utils"
)

// CallValidationApi calls the validation API to validate move resources.
//
// Parameters:
// - sourceSubscriptionId: the ID of the source subscription.
// - sourceResourceGroup: the name of the source resource group.
// - resourceIds: the IDs of the resources to validate.
// - ctx: the context for the API call.
//
// Returns:
// - []byte: the response body of the API call.
// - error: an error if the API call encounters any issues.
func CallValidationApi(sourceSubscriptionId string, sourceResourceGroup string, resourceIds string, ctx context.Context) ([]byte, error) {

	// Build the API and call it and get the response code
	validateMoveApiUrl := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/validateMoveResources?api-version=2021-04-01", sourceSubscriptionId, sourceResourceGroup)

	//get the request body
	requestBody, err := json.Marshal(CreateRequestBody(sourceSubscriptionId, sourceResourceGroup, resourceIds))
	if err != nil {
		fmt.Println("Error creating request body:", err)
		return nil, err
	}

	// Create a new http request
	//req, err := http.NewRequest(http.MethodPost, validateMoveApiUrl, bytes.NewBuffer(requestBody))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, validateMoveApiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	//get the access token
	token, err := auth.GetAzCachedAccessToken(ctx)
	if err != nil {
		fmt.Println("Error fetching access token:", err)
		return nil, err
	}

	//Add headers pass in the pointer to set the headers on the request object
	CreateRequestHeader(token, req)

	//make the API Call
	client := &http.Client{Timeout: time.Duration(20) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	//cleanup
	defer resp.Body.Close()

	// Get the response body
	respBody, err := utils.FetchResponseBody(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected status code:", resp.StatusCode)
		return nil, err
	}

	// Read response body
	// body := make([]byte, 0)
	// _, err = resp.Body.Read(body)
	// if err != nil {
	// 	fmt.Println("Error reading response body:", err)
	// 	return nil, err
	// }

	//marshall response to struct pointer
	// inverterDataerr := utils.UnmarshalDataToStruct(respBody, &InverterOutput)
	// if inverterDataerr != nil {
	// 	return inverterDataerr
	// }

	return respBody, nil

}

// func CallValidationApi(sourceSubscriptionId string, sourceResourceGroup string, resourceIds string, ctx context.Context, wg *sync.WaitGroup, resultChan chan<- string) {

// 	// Notify the WG
// 	defer wg.Done()

// 	// Build the API and call it and get the response code
// 	validateMoveApiUrl := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/validateMoveResources?api-version=2021-04-01", sourceSubscriptionId, sourceResourceGroup)

// 	req, err := http.NewRequestWithContext(ctx, "POST", validateMoveApiUrl, nil)
// 	if err != nil {
// 		fmt.Println("Error creating request:", err)
// 		return
// 	}

// 	//get the access token
// 	token, err := auth.GetAzCachedAccessToken(ctx)
// 	if err != nil {
// 		fmt.Println("Error fetching access token:", err)
// 		return
// 	}

// 	//Add headers pass in the pointer to set the headers on the request object
// 	CreateRequestHeader(token, req)

// 	//get the request body
// 	requestBody, err := json.Marshal(CreateRequestBody(sourceSubscriptionId, sourceResourceGroup, resourceIds))
// 	if err != nil {
// 		fmt.Println("Error creating request body:", err)
// 		return
// 	}

// 	//make the API Call
// 	// Make POST request
// 	resp, err := http.Post(validateMoveApiUrl, "application/json", bytes.NewBuffer(requestBody))
// 	if err != nil {
// 		fmt.Println("Error making POST request:", err)
// 		return
// 	}
// 	// resp, err := http.DefaultClient.Do(req)
// 	// if err != nil {
// 	// 	fmt.Println("Error making request:", err)
// 	// 	return
// 	// }
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		fmt.Println("Unexpected status code:", resp.StatusCode)
// 		return
// 	}

// 	// Read response body
// 	body := make([]byte, 0)
// 	_, err = resp.Body.Read(body)
// 	if err != nil {
// 		fmt.Println("Error reading response body:", err)
// 		return
// 	}

// 	// Send response body through channel
// 	resultChan <- string(body)

// }
