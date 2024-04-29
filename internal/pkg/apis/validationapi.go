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
	"sync"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
)

func CallValidationApi(sourceSubscriptionId string, sourceResourceGroup string, resourceIds string, ctx context.Context, wg *sync.WaitGroup, resultChan chan<- string) {

	// Notify the WG
	defer wg.Done()

	// Build the API and call it and get the response code
	validateMoveApiUrl := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/validateMoveResources?api-version=2021-04-01", sourceSubscriptionId, sourceResourceGroup)

	req, err := http.NewRequestWithContext(ctx, "POST", validateMoveApiUrl, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	//get the access token
	token, err := auth.GetAzCachedAccessToken(ctx)
	if err != nil {
		fmt.Println("Error fetching access token:", err)
		return
	}

	//Add headers pass in the pointer to set the headers on the request object
	CreateRequestHeader(token, req)

	//get the request body
	requestBody, err := json.Marshal(CreateRequestBody(sourceSubscriptionId, sourceResourceGroup, resourceIds))
	if err != nil {
		fmt.Println("Error creating request body:", err)
		return
	}

	//make the API Call
	// Make POST request
	resp, err := http.Post(validateMoveApiUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return
	}
	// resp, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	fmt.Println("Error making request:", err)
	// 	return
	// }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected status code:", resp.StatusCode)
		return
	}

	// Read response body
	body := make([]byte, 0)
	_, err = resp.Body.Read(body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Send response body through channel
	resultChan <- string(body)

}
