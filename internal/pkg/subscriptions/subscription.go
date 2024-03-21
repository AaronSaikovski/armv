package subscriptions

import (
	"regexp"
)

// CheckValidSubscriptionID checks if the provided subscription ID is valid.
//
// subscriptionID: a string representing a subscription ID
// bool: returns true if the subscription ID is valid, false otherwise
func CheckValidSubscriptionID(subscriptionID string) bool {
	pattern := regexp.MustCompile(`^(\{{0,1}([0-9a-fA-F]){8}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){12}\}{0,1})$`)
	return pattern.MatchString(subscriptionID)

}
