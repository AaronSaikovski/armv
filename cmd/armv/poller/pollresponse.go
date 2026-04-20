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
