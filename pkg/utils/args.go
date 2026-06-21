// Package utils provides utility functions and structures for command-line argument parsing,
// version management, and common helper functions for the ARMV application.
package utils

const (
	// AppDescription is the application description
	AppDescription = `ARMV - Azure Resource Movability Validator

Performs a Read-Only check whether resources in a source resource group
can be moved to a target resource group in the same subscription.`
)

// Args holds the command-line arguments
type Args struct {
	SourceSubscriptionId string
	SourceResourceGroup  string
	TargetSubscriptionId string
	TargetResourceGroup  string
	Debug                bool
	OutputPath           string
}

// FormatVersion returns the formatted version string for display.
func FormatVersion(version string) string {
	return "ARMV version: " + version
}
