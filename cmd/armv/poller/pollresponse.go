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

// writeOutput builds a ValidationReport, renders it as Markdown, and writes
// it to a timestamped .md file under outputPath. The returned report is
// returned so the caller can drive the console summary from the same data.
func (pollResp *PollerResponseData) writeOutput(outputPath string, ctx ReportContext) (ValidationReport, error) {
	fileName := fmt.Sprintf("output-%s.md", time.Now().Format("2006-01-02-15-04-05"))

	// Pretty-print the raw Azure body (if any). Non-JSON bodies are kept
	// verbatim rather than failing the operation — the markdown rendering
	// handles unparseable JSON by showing a FAILED header with no table.
	var prettyJSON string
	if pollResp.RespStatusCode != StatusMoveOK && len(pollResp.RespBody) > 0 {
		if pj, err := utils.PrettyJsonString(string(pollResp.RespBody)); err == nil {
			prettyJSON = pj
		} else {
			prettyJSON = string(pollResp.RespBody)
		}
	}

	report := BuildValidationReport(pollResp.RespStatusCode, pollResp.RespStatus, pollResp.RespBody, prettyJSON, ctx)
	markdown := RenderMarkdown(report)

	if err := utils.WriteOutputFile(outputPath, fileName, markdown); err != nil {
		return ValidationReport{}, fmt.Errorf("failed to write output file: %w", err)
	}
	return report, nil
}
