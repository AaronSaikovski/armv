package app

import (
	"fmt"

	"github.com/AaronSaikovski/gostarter/pkg/samplemodule"
	"github.com/logrusorgru/aurora"

	"github.com/AaronSaikovski/gostarter/constants"
	"github.com/AaronSaikovski/gostarter/types"
)

// run - main run method
func Run() error {

	fmt.Println(aurora.BrightGreen(string("This is the main function.\n")))

	for i := 0; i < constants.LoopConstant; i++ {
		fmt.Printf("print using loop const \n")
	}

	samplemodule.SampleFunction()

	fmt.Print(types.Sample{SampleString: "hello from struct", SampleInt: 1})

	return nil
}
