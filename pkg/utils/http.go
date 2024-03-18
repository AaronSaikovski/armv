package utils

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/AaronSaikovski/armv/types"
)

// CallAPI makes an HTTP request to the given URL and sends the response or error message through the provided channel.
//
// url string, wg *sync.WaitGroup, ch chan string
// No return type
func CallAPI(url string, wg *sync.WaitGroup, ch chan string) {
	defer wg.Done()

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprintf("Error: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	// Read response body
	// For simplicity, just printing the response body here
	// You can parse and process the response as needed
	ch <- fmt.Sprintf("Response from %s: %s", url, resp.Status)
}

// SetRequestHeaders sets the request headers for the HTTP request.
//
// r *http.Request - the HTTP request to set headers for
// cachedAccessToken string - the cached access token to use for authorization
func SetRequestHeaders(r *http.Request, cachedAccessToken string) {
	bearerToken := "Bearer ${cachedAccessToken}"
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", bearerToken)
}

// SetBodyText generates a types.ApiBodyText object based on the provided parameters.
//
// resourceIds - the string containing the resources.
// targetSubscriptionId - the string representing the target subscription ID.
// targetResourceGroup - the string containing the target resource group.
// types.ApiBodyText - returns the generated ApiBodyText object.
func SetBodyText(resourceIds string, targetSubscriptionId string, targetResourceGroup string) types.ApiBodyText {
	ApiBodyText := types.ApiBodyText{
		Resources:           resourceIds,
		TargetResourceGroup: "/subscriptions/${targetSubscriptionId/resourceGroups/${targetResourceGroup}",
	}
	return ApiBodyText
}
