package app

import (
	"fmt"
	"strconv"

	"github.com/AaronSaikovski/armv/internal/pkg/azure/auth"
	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/AaronSaikovski/armv/types"
)

var (
	args        utils.Args
	inputParams types.Params
)

// run - main run method
func Run() error {

	// check params
	if err := checkParams(); err != nil {
		return err
	}

	// populate the params struct
	inputParams = types.Params{
		SourceSubscriptionId: args.SourceSubscriptionId,
		SourceResourceGroup:  args.SourceResourceGroup,
		TargetSubscriptionId: args.TargetSubscriptionId,
		TargetResourceGroup:  args.TargetResourceGroup,
	}

	//Print the args
	fmt.Printf("Source Subscription Id: %s\n", inputParams.SourceSubscriptionId)
	fmt.Printf("Source Resource Group: %s\n", inputParams.SourceResourceGroup)
	fmt.Printf("Target Subscription Id: %s\n", inputParams.TargetSubscriptionId)
	fmt.Printf("Target Resource Group: %s\n", inputParams.TargetResourceGroup)

	isLoggedIn := auth.DoLogin(inputParams.SourceSubscriptionId)
	fmt.Printf("Logged in to Azure? %s \n", strconv.FormatBool(isLoggedIn))

	return nil
}
