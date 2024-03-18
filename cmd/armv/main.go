package main

import (
	"os"

	"github.com/AaronSaikovski/armv/app"
	"github.com/AaronSaikovski/armv/pkg/utils"
)

// main is the entry point of the program.
//
// No parameters.
// No return types.
func main() {

	if err := app.Run(); err != nil {
		utils.HandleError(err)
		os.Exit(1)
	}
}
