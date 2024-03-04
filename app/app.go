package app

import (
	"fmt"

	"github.com/logrusorgru/aurora"

	"github.com/AaronSaikovski/armv/constants"
	"github.com/AaronSaikovski/armv/pkg/samplemodule"
	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/AaronSaikovski/armv/types"
	"github.com/alexflint/go-arg"
)

// run - main run method
func Run() error {

	//Get the args input data
	var args utils.Args
	p := arg.MustParse(&args)

	//Print the args
	fmt.Printf("Source Subscription Id: %s\n", args.SourceSubscriptionId)
	fmt.Printf("Source Resource Group: %s\n", args.SourceResourceGroup)
	fmt.Printf("Target Subscription Id: %s\n", args.TargetSubscriptionId)
	fmt.Printf("Target Resource Group: %s\n", args.TargetResourceGroup)

	//check for valid subscription Id
	if !utils.CheckValidSubscriptionID(args.SourceSubscriptionId) {
		p.Fail("Invalid Source Subscription ID format: - should be: 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX'.")
	}


	fmt.Println(aurora.BrightGreen(string("This is the main function.\n")))

	for i := 0; i < constants.LoopConstant; i++ {
		fmt.Printf("print using loop const \n")
	}

	samplemodule.SampleFunction()

	fmt.Print(types.Sample{SampleString: "hello from struct", SampleInt: 1})

	return nil
}
