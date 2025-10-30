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
	"fmt"
	"time"

	"github.com/AaronSaikovski/armv/pkg/utils"
)

// writeOutput writes the output to a file with a timestamp in the filename.
//
// No parameters.
// Returns an error if writing fails.
func (pollResp *PollerResponseData) writeOutput(outputPath string) error {
	fileName := fmt.Sprintf("output-%s.txt", time.Now().Format("2006-01-02-15-04-05"))

	var output string
	var err error

	if pollResp.RespStatusCode == API_RESOURCE_MOVE_OK {
		output = fmt.Sprintf("*** SUCCESS - No Azure Resource Validation issues found. ***\n*** Response Status Code OK: %s ***", pollResp.RespStatus)
	} else {
		output, err = utils.PrettyJsonString(string(pollResp.RespBody))
		if err != nil {
			return fmt.Errorf("failed to format JSON output: %w", err)
		}
	}

	if err := utils.WriteOutputFile(outputPath, fileName, output); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
