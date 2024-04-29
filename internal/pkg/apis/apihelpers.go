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
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateRequestHeader builds the request header dictionary.
//
// cachedAccessToken string - the cached access token used for authorization.
// map[string]string - returns the request header dictionary.
func CreateRequestHeader(cachedAccessToken string, r *http.Request) {

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cachedAccessToken))
}

// CreateRequestBody generates a JSON body based on the provided parameters.
//
// targetSubscriptionID - the ID of the target subscription.
// targetResourceGroup - the name of the target resource group.
// resourceIDs - the IDs of the resources to include in the body.
// string - returns the generated JSON body as a string.
func CreateRequestBody(targetSubscriptionID string, targetResourceGroup string, resourceIDs string) string {
	bodyMap := map[string]interface{}{
		"resources":           resourceIDs,
		"targetResourceGroup": fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", targetSubscriptionID, targetResourceGroup),
	}

	bodyJSON, _ := json.Marshal(bodyMap)
	return string(bodyJSON)

}
