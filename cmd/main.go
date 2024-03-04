package main

import (
	"os"

	"github.com/AaronSaikovski/armv/app"
	"github.com/AaronSaikovski/armv/pkg/utils"
)

// main - program main
func main() {

	if err := app.Run(); err != nil {
		utils.HandleError(err)
		os.Exit(1)
	}
}
