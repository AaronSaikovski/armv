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
package utils

import (
	"fmt"

	"github.com/logrusorgru/aurora"
)

// OutputSuccess prints a success message indicating no Azure Resource move validation issues found.
//
// No parameters.
func OutputSuccess(respStatus string) {
	fmt.Println(aurora.Bold(aurora.Green("\n*****************************************************************")))
	fmt.Println(aurora.Bold(aurora.Green("*** SUCCESS - No Azure Resource Move Validation issues found. ***")))
	fmt.Println(aurora.Sprintf(aurora.Green("*** Response Status OK - %s ***"), aurora.Green(respStatus)))
	fmt.Println(aurora.Bold(aurora.Green("*****************************************************************")))
}

// OutputFail prints an error message with the specified SourceResourceGroup and ErrorDetails.
//
// Parameters:
// - SourceResourceGroup: the name of the source resource group.
// - ErrorDetails: details about the error that occurred.
func OutputFail(SourceResourceGroup string, ErrorDetails string) {

	fmt.Println(aurora.Bold(aurora.Red("\n*****************************************************************")))
	fmt.Println(aurora.Sprintf(aurora.Red("*** Source ResourceGroup - '%s' ***"), SourceResourceGroup))
	fmt.Println(aurora.Sprintf(aurora.Red("*** Error Details: \n %s "), aurora.Red(ErrorDetails)))
	fmt.Println(aurora.Bold(aurora.Red("*****************************************************************")))

}
