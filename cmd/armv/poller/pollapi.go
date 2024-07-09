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
package poller

import (
	"context"
	"sync"
	"time"

	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

var (
	pollResp = PollerResponseData{}
)

// ** OLD Sync Code **
// PollApi is a function that polls the AzureRM Validation API indefinitely until it receives a response.
//
// It takes the following parameters:
// - ctx: the context.Context object for cancellation and timeout control.
// - respPoller: a pointer to the runtime.Poller[T] object that handles the polling.
//
// It returns the following:
// - types.PollerResponse: the response from the API.
// - error: an error if any occurred during the polling process.
// func PollApi[T any](ctx context.Context, respPoller *runtime.Poller[T]) error {

// 	bar := progressbar.NewOptions(progressBarMax,
// 		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
// 		progressbar.OptionEnableColorCodes(true),
// 		progressbar.OptionSetDescription("[cyan][reset] Running Validation..."),
// 		progressbar.OptionSetTheme(progressbar.Theme{
// 			Saucer:        "[green]=[reset]",
// 			SaucerHead:    "[green]>[reset]",
// 			SaucerPadding: " ",
// 			BarStart:      "[",
// 			BarEnd:        "]",
// 		}))
// 	// defer func() {
// 	// 	_ = bar.Finish()
// 	// }()

// 	barCount := 0
// 	for {
// 		barCount++
// 		_ = bar.Add(1)
// 		time.Sleep(sleepDuration)

// 		if barCount >= progressBarMax {
// 			bar.Reset()
// 			barCount = 0
// 		}

// 		w, err := respPoller.Poll(ctx)
// 		if err != nil {
// 			return err
// 		}

// 		if respPoller.Done() {

// 			_ = bar.Finish()

// 			pollResp = PollerResponseData{
// 				RespBody:       utils.FetchResponseBody(w.Body),
// 				RespStatusCode: w.StatusCode,
// 				RespStatus:     w.Status,
// 			}

// 			pollResp.displayOutput()
// 			ctx.Done()
// 			return nil

// 		}
// 	}
// }

// PollApi polls the API and displays the response.
//
// It takes a context.Context and a *runtime.Poller[T] as parameters.
// The context is used for cancellation and timeout control.
// The respPoller is used to handle the polling.
//
// It returns an error if any occurred during the polling process.
func PollApi[T any](ctx context.Context, respPoller *runtime.Poller[T]) error {

	wg := new(sync.WaitGroup)
	wg.Add(1)

	//error channel
	errChan := make(chan error, 1)

	go func() {

		defer wg.Done() // Signal that this goroutine is done

		//progress bar
		bar := progressBar()

		barCount := 0
		for {

			barCount++
			err := bar.Add(1)
			if err != nil {
				errChan <- err
				return
			}

			time.Sleep(sleepDuration)

			if barCount >= progressBarMax {
				bar.Reset()
				barCount = 0
			}

			w, err := respPoller.Poll(ctx)
			if err != nil {
				errChan <- err
				return
			}

			if respPoller.Done() {

				err := bar.Finish()
				if err != nil {
					errChan <- err
					return
				}

				// create new PollerResponseData
				pollResp = NewPollerResponseData(utils.FetchResponseBody(w.Body), w.StatusCode, w.Status)

				// display output
				pollResp.displayOutput()

				// close context
				ctx.Done()

				return

			}

		}
	}()

	// Wait for the goroutine to finish
	wg.Wait()

	// Close the error channel
	close(errChan)

	// Check for errors
	if err := <-errChan; err != nil {
		return err
	}

	return nil
}
