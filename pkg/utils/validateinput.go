package utils

import (
	"regexp"
)

// uuidPattern matches a canonical (unbraced) UUID. Compiled once at package
// initialization for performance. Azure subscription IDs are always bare UUIDs.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// CheckValidSubscriptionID reports whether subscriptionID is a well-formed UUID.
func CheckValidSubscriptionID(subscriptionID string) bool {
	return uuidPattern.MatchString(subscriptionID)
}
