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
package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/AaronSaikovski/armv/types"
)

func MoveApiLookup(ctx context.Context, wg *sync.WaitGroup, resultChan chan<- string, subscriptionID string, sourceResourceGroupName string, bearerToken string, requestBody *types.ApiBodyText) {

	// Obtain the Move API reference - Return code 409 means an error
	defer wg.Done()

	//Format the Validation API string
	validationUrl := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/validateMoveResources?api-version=2021-04-01", subscriptionID, sourceResourceGroupName)

	// Marshal the request body to JSON
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("Error marshalling request body:", err)
		return
	}

	// Build the request
	req, err := http.NewRequestWithContext(ctx, "POST", validationUrl, io.NopCloser(bytes.NewReader(bodyBytes)))
	if err != nil {
		logger.Error("Error creating request:", err)
		return
	}

	//set the request header
	setRequestHeaders(req, bearerToken)

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("Unexpected status code: %d", resp.StatusCode))
		return
	}

	// Read response body
	body := make([]byte, 0)
	_, err = resp.Body.Read(body)
	if err != nil {
		logger.Error("Error reading response body:", err)
		return
	}

	// Send response body through channel
	resultChan <- string(body)

}
