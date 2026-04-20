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

// writeOutput writes the poller response to a timestamped file under outputPath.
func (pollResp *PollerResponseData) writeOutput(outputPath string) error {
	fileName := fmt.Sprintf("output-%s.txt", time.Now().Format("2006-01-02-15-04-05"))

	output := pollResp.Format()
	if err := utils.WriteOutputFile(outputPath, fileName, output); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	return nil
}

// Format renders the response for persistence, covering the observable
// terminal states: success (204), an error body that parses as JSON, an
// empty error body, and a non-JSON error body. Format never returns an error;
// malformed bodies are persisted verbatim so an operator has something to
// diagnose.
func (pollResp *PollerResponseData) Format() string {
	if pollResp.RespStatusCode == StatusMoveOK {
		return fmt.Sprintf(
			"*** SUCCESS - No Azure Resource Validation issues found. ***\n*** Response Status Code OK: %s ***",
			pollResp.RespStatus,
		)
	}

	if len(pollResp.RespBody) == 0 {
		return fmt.Sprintf(
			"*** Azure Resource Validation returned status %d %q with no response body. ***",
			pollResp.RespStatusCode, pollResp.RespStatus,
		)
	}

	pretty, err := utils.PrettyJsonString(string(pollResp.RespBody))
	if err != nil {
		// Non-JSON body: persist verbatim rather than failing the operation.
		return fmt.Sprintf(
			"*** Azure Resource Validation returned status %d %q. Raw body: ***\n%s",
			pollResp.RespStatusCode, pollResp.RespStatus, string(pollResp.RespBody),
		)
	}
	return pretty
}
