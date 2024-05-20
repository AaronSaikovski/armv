/*
MIT License

# Copyright (c) 2024 Aaron Saikovski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package app

import (
	"fmt"

	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/alexflint/go-arg"
)

// checkParams checks the input arguments for validity.
//
// No parameters.
// // Returns an error.
// func checkParams() error {
// 	//Get the args input data
// 	p := arg.MustParse(&args)

// 	//check for valid subscription Id
// 	if !utils.CheckValidSubscriptionID(args.SourceSubscriptionId) {
// 		p.Fail("Invalid Source Subscription ID format: - should be: '0000-0000-0000-000000000000'.")
// 	}

// 	//check for valid subscription Id
// 	if !utils.CheckValidSubscriptionID(args.TargetSubscriptionId) {
// 		p.Fail("Invalid Target Subscription ID format: - should be: '0000-0000-0000-000000000000'.")
// 	}
// 	return nil
// }

// checkParams validates the input arguments for validity.
//
// It retrieves the args input data using arg.MustParse(&args) and defines a helper function validateSubscriptionID to validate the subscription ID format.
// The helper function checks if the subscription ID is valid using utils.CheckValidSubscriptionID and fails if it is not.
// The function then validates the Source and Target Subscription IDs using the validateSubscriptionID function.
// Finally, it returns nil indicating that there were no errors.
//
// Returns:
// - error: nil if there are no errors, otherwise an error indicating the invalid subscription ID format.
func checkParams() error {
	// Get the args input data
	p := arg.MustParse(&args)

	// Helper function to validate subscription ID
	validateSubscriptionID := func(id, name string) {
		if !utils.CheckValidSubscriptionID(id) {
			p.Fail(fmt.Sprintf("Invalid %s Subscription ID format: - should be: '0000-0000-0000-000000000000'.", name))
		}
	}

	// Validate Source and Target Subscription IDs
	validateSubscriptionID(args.SourceSubscriptionId, "Source")
	validateSubscriptionID(args.TargetSubscriptionId, "Target")

	return nil
}
