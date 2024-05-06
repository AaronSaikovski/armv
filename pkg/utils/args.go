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

var (
	// Version string
	VersionString string = "armv v0.0.1-beta"

	infoString string = `
   		*** Azure Resource Movability Validator (ARMV)* **
		This utility checks whether the specified Azure resources can be moved to the target location in the same tenant.  
		If validation succeeds, it returns HTTP response code 204 (no content).
		If validation fails, it returns HTTP response code 409 (Conflict) with an error message. 
		** This only performs a read of resources and NO changes are made. **
    `
)

// Args - struct using go-arg- https://github.com/alexflint/go-arg
type Args struct {
	SourceSubscriptionId string `arg:"required,--SourceSubscriptionId" help:"Source Subscription Id."`
	SourceResourceGroup  string `arg:"required,--SourceResourceGroup" help:"Source Resource Group."`
	TargetSubscriptionId string `arg:"required,--TargetSubscriptionId" help:"Target Subscription Id."`
	TargetResourceGroup  string `arg:"required,--TargetResourceGroup" help:"Target Resource Group."`
}

// Description - App description
func (Args) Description() string {
	return infoString
}

// Version - Version info
func (Args) Version() string {
	return VersionString
}
