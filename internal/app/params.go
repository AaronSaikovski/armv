package app

import (
	"github.com/AaronSaikovski/armv/internal/pkg/subscriptions"
	"github.com/alexflint/go-arg"
)

// checkParams checks the parameters of the function.
//
// It does not take any parameters.
// It returns an error.
func checkParams() error {
	//Get the args input data
	p := arg.MustParse(&args)

	//check for valid subscription Id
	if !subscriptions.CheckValidSubscriptionID(args.SourceSubscriptionId) {
		p.Fail("Invalid Source Subscription ID format: - should be: '0000-0000-0000-000000000000'.")
	}

	//check for valid subscription Id
	if !subscriptions.CheckValidSubscriptionID(args.TargetSubscriptionId) {
		p.Fail("Invalid Target Subscription ID format: - should be: '0000-0000-0000-000000000000'.")
	}
	return nil
}
