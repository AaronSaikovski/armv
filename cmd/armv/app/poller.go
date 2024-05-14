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
	"context"
	"time"

	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

// pollApi polls the AzureRM Validation API until the response is ready.
//
// It takes a context.Context object and a *runtime.Poller[T] object as parameters.
// The context.Context object is used to control the execution of the function.
// The *runtime.Poller[T] object represents the response poller.
//
// The function returns a PollerResponse object and an error.
// The PollerResponse object contains the response body, status code, and status.
// The error object represents any error that occurred during the polling process.
// func pollApi[T any](ctx context.Context, respPoller *runtime.Poller[T]) (pollerResp PollerResponse, err error) {

// 	poller := PollerResponse{}

// 	barCount := 0
// 	bar := progressbar.NewOptions(PROGRESS_BAR_MAX,
// 		progressbar.OptionSetWriter(ansi.NewAnsiStdout()), //you should install "github.com/k0kubun/go-ansi"
// 		progressbar.OptionEnableColorCodes(true),
// 		//progressbar.OptionSetWidth(50),
// 		progressbar.OptionSetDescription("[cyan][reset] Polling AzureRM Validation API..."),
// 		progressbar.OptionSetTheme(progressbar.Theme{
// 			Saucer:        "[green]=[reset]",
// 			SaucerHead:    "[green]>[reset]",
// 			SaucerPadding: " ",
// 			BarStart:      "[",
// 			BarEnd:        "]",
// 		}))

// 	// polling loop
// 	for {
// 		bar.Add(1)
// 		barCount += 1
// 		time.Sleep(5 * time.Millisecond)
// 		if barCount >= PROGRESS_BAR_MAX {
// 			bar.Reset()
// 			barCount = 0
// 		}

// 		//poll the response
// 		w, err := respPoller.Poll(ctx)
// 		if err != nil {
// 			return PollerResponse{}, err
// 		}

// 		//if done. update response codes
// 		if respPoller.Done() {
// 			poller.respBody = utils.FetchResponseBody(w.Body)
// 			poller.respStatusCode = w.StatusCode
// 			poller.respStatus = w.Status
// 			bar.Finish()
// 			break
// 		}

// 	}
// 	return poller, nil

// }

func pollApi[T any](ctx context.Context, respPoller *runtime.Poller[T]) (pollerResp PollerResponse, err error) {
	poller := PollerResponse{}

	const (
		progressBarMax = 100
		sleepDuration  = 5 * time.Millisecond
	)

	bar := progressbar.NewOptions(progressBarMax,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription("[cyan][reset] Polling AzureRM Validation API..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	defer bar.Finish()

	pollingLoop := func() (PollerResponse, error) {
		barCount := 0
		for {
			bar.Add(1)
			barCount++
			time.Sleep(sleepDuration)
			if barCount >= progressBarMax {
				bar.Reset()
				barCount = 0
			}

			w, err := respPoller.Poll(ctx)
			if err != nil {
				return PollerResponse{}, err
			}

			if respPoller.Done() {
				return PollerResponse{
					respBody:       utils.FetchResponseBody(w.Body),
					respStatusCode: w.StatusCode,
					respStatus:     w.Status,
				}, nil
			}
		}
	}

	poller, err = pollingLoop()
	if err != nil {
		return PollerResponse{}, err
	}

	return poller, nil
}
