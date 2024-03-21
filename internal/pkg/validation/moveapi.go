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
