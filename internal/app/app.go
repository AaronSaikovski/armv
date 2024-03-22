package app

import (
	"fmt"
	"strconv"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/mattn/go-colorable"
	//"github.com/AaronSaikovski/armv/types"
)

var (
	args utils.Args
	//inputParams types.Params
)

// run - main run method
func Run() error {

	//ctx := context.Background()

	restoreColorMode := colorable.EnableColorsStdout(nil)
	defer restoreColorMode()

	// check params
	if err := checkParams(); err != nil {
		return err
	}

	// populate the params struct
	// inputParams = types.Params{
	// 	SourceSubscriptionId: args.SourceSubscriptionId,
	// 	SourceResourceGroup:  args.SourceResourceGroup,
	// 	TargetSubscriptionId: args.TargetSubscriptionId,
	// 	TargetResourceGroup:  args.TargetResourceGroup,
	// }

	//Print the args
	fmt.Printf("Source Subscription Id: %s\n", args.SourceSubscriptionId)
	fmt.Printf("Source Resource Group: %s\n", args.SourceResourceGroup)
	fmt.Printf("Target Subscription Id: %s\n", args.TargetSubscriptionId)
	fmt.Printf("Target Resource Group: %s\n", args.TargetResourceGroup)

	isLoggedIn := auth.DoLogin(args.SourceSubscriptionId) //inputParams.SourceSubscriptionId)
	fmt.Printf("Logged in to Azure? %s \n", strconv.FormatBool(isLoggedIn))

	return nil
}
