package app

import (
	"context"
	"fmt"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/AaronSaikovski/armv/internal/pkg/resourcegroups"
	"github.com/AaronSaikovski/armv/internal/pkg/resources"
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

	ctx := context.Background()

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
	// fmt.Printf("Source Subscription Id: %s\n", args.SourceSubscriptionId)
	// fmt.Printf("Source Resource Group: %s\n", args.SourceResourceGroup)
	// fmt.Printf("Target Subscription Id: %s\n", args.TargetSubscriptionId)
	// fmt.Printf("Target Resource Group: %s\n", args.TargetResourceGroup)

	/* ********************************************************************** */
	// check we are logged into the Azure source subscription
	isLoggedIn := auth.GetLogin(args.SourceSubscriptionId)
	if !isLoggedIn {
		return fmt.Errorf("not logged into azure, please login and retry operation.")
	}

	/* ********************************************************************** */

	//check source and destination resource groups exist
	srcRsgExists, err := resourcegroups.CheckResourceGroupExists(ctx, args.SourceResourceGroup)
	if err != nil {
		return err
	}
	if !srcRsgExists {
		return fmt.Errorf("source resource group %s does not exist", args.SourceResourceGroup)
	}

	/* ********************************************************************** */

	//check destination and destination resource groups exist
	dstRsgExists, err := resourcegroups.CheckResourceGroupExists(ctx, args.TargetResourceGroup)
	if err != nil {
		return err
	}
	if !dstRsgExists {
		return fmt.Errorf("destination resource group %s does not exist", args.TargetResourceGroup)
	}

	/* ********************************************************************** */

	// Get all resource IDs from source resource group
	resourceIds, err := resources.GetResourceIds(ctx, args.SourceResourceGroup)
	if err != nil {
		return err
	}

	fmt.Println(resourceIds)

	/* ********************************************************************** */

	/* ********************************************************************** */

	return nil
}
