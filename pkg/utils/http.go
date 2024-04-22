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
package utils

// import (
// 	"fmt"
// 	"net/http"
// 	"sync"

// 	"github.com/AaronSaikovski/armv/types"
// )

// // CallAPI makes an HTTP request to the given URL and sends the response or error message through the provided channel.
// //
// // url string, wg *sync.WaitGroup, ch chan string
// // No return type
// func CallAPI(url string, wg *sync.WaitGroup, ch chan string) {
// 	defer wg.Done()

// 	// Make HTTP request
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		ch <- fmt.Sprintf("Error: %s", err.Error())
// 		return
// 	}
// 	defer resp.Body.Close()

// 	// Read response body
// 	// For simplicity, just printing the response body here
// 	// You can parse and process the response as needed
// 	ch <- fmt.Sprintf("Response from %s: %s", url, resp.Status)
// }

// SetRequestHeaders sets the request headers for the HTTP request.
//
// r *http.Request - the HTTP request to set headers for
// cachedAccessToken string - the cached access token to use for authorization
// func SetRequestHeaders(r *http.Request, cachedAccessToken string) {
// 	bearerToken := "Bearer ${cachedAccessToken}"
// 	r.Header.Add("Content-Type", "application/json")
// 	r.Header.Add("Authorization", bearerToken)
// }

// // SetBodyText generates a types.ApiBodyText object based on the provided parameters.
// //
// // resourceIds - the string containing the resources.
// // targetSubscriptionId - the string representing the target subscription ID.
// // targetResourceGroup - the string containing the target resource group.
// // types.ApiBodyText - returns the generated ApiBodyText object.
// func SetBodyText(resourceIds string, targetSubscriptionId string, targetResourceGroup string) types.ApiBodyText {
// 	ApiBodyText := types.ApiBodyText{
// 		Resources:           resourceIds,
// 		TargetResourceGroup: "/subscriptions/${targetSubscriptionId/resourceGroups/${targetResourceGroup}",
// 	}
// 	return ApiBodyText
// }
