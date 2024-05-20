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

import (
	"regexp"
)

// validateId checks if the inputId string matches the specified pattern.
//
// inputId - the string to validate.
// bool - returns true if the inputId matches the pattern, false otherwise.
func validateId(inputId string) bool {
	pattern := regexp.MustCompile(`^(\{{0,1}([0-9a-fA-F]){8}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){12}\}{0,1})$`)
	return pattern.MatchString(inputId)
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
