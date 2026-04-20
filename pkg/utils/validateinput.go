package utils

import (
	"regexp"
)

// uuidPattern is compiled once at package initialization for performance
var uuidPattern = regexp.MustCompile(`^(\{{0,1}([0-9a-fA-F]){8}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){12}\}{0,1})$`)

// validateId checks if the inputId string matches the specified pattern.
//
// inputId - the string to validate.
// bool - returns true if the inputId matches the pattern, false otherwise.
func validateId(inputId string) bool {
	return uuidPattern.MatchString(inputId)
}

// CheckValidSubscriptionID checks if the provided subscription ID is valid.
//
// subscriptionID: a string representing a subscription ID
// bool: returns true if the subscription ID is valid, false otherwise
func CheckValidSubscriptionID(subscriptionID string) bool {
	return validateId(subscriptionID)
}

// CheckValidTenantID checks if the provided tenant ID is valid.
//
// tenantID: a string representing a tenant ID
// bool: returns true if the tenant ID is valid, false otherwise
func CheckValidTenantID(tenantID string) bool {
	return validateId(tenantID)
}
