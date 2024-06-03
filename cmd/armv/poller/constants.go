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
<<<<<<<< HEAD:cmd/armv/poller/constants.go
package poller

import (
	"time"
)

const (
	progressBarMax = 100
	sleepDuration  = 5 * time.Millisecond

	//API return codes
	API_SUCCESS            int = 202
	API_RESOURCE_MOVE_OK   int = 204
	API_RESOURCE_MOVE_FAIL int = 409

	//Progress bar Max
	PROGRESS_BAR_MAX int = 100
)
========
package validation

type AzureResourceMoveInfo struct {
	SourceSubscriptionId  string
	SourceResourceGroup   string
	TargetResourceGroup   string
	TargetResourceGroupId *string
	ResourceIds           []*string
}
>>>>>>>> 64e343d (optimised code and added types):internal/pkg/validation/azureresourcemoveinfo.go
