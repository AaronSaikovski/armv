package main

import (
	"os"

	"github.com/AaronSaikovski/gostarter/app"
	"github.com/AaronSaikovski/gostarter/pkg/utils"
)

// main - program main
func main() {

	if err := app.Run(); err != nil {
		utils.HandleError(err)
		os.Exit(1)
	}
}
