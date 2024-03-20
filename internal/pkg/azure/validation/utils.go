package validation

import (
	"net/http"
)

// SetRequestHeaders sets the request headers for the HTTP request.
//
// r *http.Request - the HTTP request to set headers for
// cachedAccessToken string - the cached access token to use for authorization
func setRequestHeaders(r *http.Request, cachedAccessToken string) {
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
// func setBodyText(resourceIds string, targetSubscriptionId string, targetResourceGroup string) types.ApiBodyText {
// 	ApiBodyText := types.ApiBodyText{
// 		Resources:           resourceIds,
// 		TargetResourceGroup: "/subscriptions/${targetSubscriptionId/resourceGroups/${targetResourceGroup}",
// 	}
// 	return ApiBodyText
// }
