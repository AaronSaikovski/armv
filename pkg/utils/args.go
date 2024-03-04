package utils

var (
	// Version string
	VersionString string = "armv v0.0.1"

	infoString string = `
        Validates a source Azure resource group
        and all child resources to check for moveability support into a target
        resource group within a target subscription.
    `
)

// Args - struct using go-arg- https://github.com/alexflint/go-arg
type Args struct {
	SourceSubscriptionId string `arg:"required,-s,--srcsubid" help:"Source Subscription Id.."`
	SourceResourceGroup  string `arg:"required,-r,--srcrsg" help:"Source Resource Group."`
	TargetSubscriptionId string `arg:"required,-t,--targsubid" help:"Target Subscription Id."`
	TargetResourceGroup  string `arg:"required,-d,--targrsg" help:"Target Resource Group."`
}

// Description - App description
func (Args) Description() string {
	return infoString
}

// Version - Version info
func (Args) Version() string {
	return VersionString
}
