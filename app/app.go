package app

import (
	"fmt"

	"github.com/AaronSaikovski/armv/internal/pkg/azure/subscriptions"
	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/alexflint/go-arg"
)

// checkParams checks the parameters of the function.
//
// It does not take any parameters.
// It returns an error.
func checkParams() error {
	//Get the args input data
	var args utils.Args
	p := arg.MustParse(&args)

	//Print the args
	fmt.Printf("Source Subscription Id: %s\n", args.SourceSubscriptionId)
	fmt.Printf("Source Resource Group: %s\n", args.SourceResourceGroup)
	fmt.Printf("Target Subscription Id: %s\n", args.TargetSubscriptionId)
	fmt.Printf("Target Resource Group: %s\n", args.TargetResourceGroup)

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

// run - main run method
func Run() error {

	// check params
	if err := checkParams(); err != nil {
		return err
	}

	return nil
}
