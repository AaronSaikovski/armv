package utils

import (
	"regexp"
)

// CheckValidSubscriptionID - check if subscription id is valid
func CheckValidSubscriptionID(subscriptionID string) bool {
	pattern := regexp.MustCompile(`^(\{{0,1}([0-9a-fA-F]){8}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){12}\}{0,1})$`)
	return pattern.MatchString(subscriptionID)

}
