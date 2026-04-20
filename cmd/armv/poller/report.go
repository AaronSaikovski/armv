package poller

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ReportContext holds the validation-run metadata used to populate the report header.
type ReportContext struct {
	SourceSubscriptionID string
	SourceResourceGroup  string
	TargetSubscriptionID string
	TargetResourceGroup  string
	ResourceCount        int
}

// AzureErrorResponse mirrors the shape of the JSON returned by the
// Azure Validate Move Resources API on a 409 response.
type AzureErrorResponse struct {
	Error struct {
		Code    string             `json:"code"`
		Message string             `json:"message"`
		Details []AzureErrorDetail `json:"details"`
	} `json:"error"`
}

// AzureErrorDetail is one entry from the Azure error response's details array.
type AzureErrorDetail struct {
	Code    string `json:"code"`
	Target  string `json:"target"`
	Message string `json:"message"`
}

// ValidationReport is the parsed, rendered form of the API response.
type ValidationReport struct {
	Success     bool
	GeneratedAt time.Time
	Context     ReportContext
	StatusCode  int
	StatusText  string
	TopLevel    AzureErrorDetail // code+message summarising the failure (empty on success)
	Errors      []ValidationError
	RawJSON     string
}

// ValidationError is one failing resource, flattened from AzureErrorDetail.
type ValidationError struct {
	ResourceID   string
	ResourceType string
	ResourceName string
	Code         string
	Message      string
}

// BuildValidationReport turns a raw API response into a ValidationReport.
// prettyJSON may be empty for the 204 case.
func BuildValidationReport(statusCode int, statusText string, rawBody []byte, prettyJSON string, ctx ReportContext) ValidationReport {
	report := ValidationReport{
		Success:     statusCode == StatusMoveOK,
		GeneratedAt: time.Now().UTC(),
		Context:     ctx,
		StatusCode:  statusCode,
		StatusText:  statusText,
		RawJSON:     prettyJSON,
	}

	if report.Success || len(rawBody) == 0 {
		return report
	}

	var parsed AzureErrorResponse
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		// Parsing failed — keep the raw JSON so operators can still diagnose.
		return report
	}

	report.TopLevel = AzureErrorDetail{
		Code:    parsed.Error.Code,
		Message: parsed.Error.Message,
	}
	report.Errors = make([]ValidationError, 0, len(parsed.Error.Details))
	for _, d := range parsed.Error.Details {
		resourceType, resourceName := ParseResourceID(d.Target)
		report.Errors = append(report.Errors, ValidationError{
			ResourceID:   d.Target,
			ResourceType: resourceType,
			ResourceName: resourceName,
			Code:         d.Code,
			Message:      d.Message,
		})
	}
	return report
}

// ParseResourceID extracts the provider/type and name from an Azure resource ID like
// /subscriptions/<sub>/resourceGroups/<rg>/providers/<ns>/<type>/<name>.
// If the shape is not recognised, both return values fall back to the original target.
func ParseResourceID(target string) (resourceType, resourceName string) {
	if target == "" {
		return "", ""
	}
	idx := strings.Index(target, "/providers/")
	if idx < 0 {
		return target, target
	}
	parts := strings.Split(strings.TrimPrefix(target[idx:], "/providers/"), "/")
	// Expect at least: <namespace>/<type>/<name>
	if len(parts) < 3 {
		return target, target
	}
	resourceType = parts[0] + "/" + parts[1]
	resourceName = parts[len(parts)-1]
	return resourceType, resourceName
}

// RenderMarkdown produces the Markdown report body.
func RenderMarkdown(r ValidationReport) string {
	var b strings.Builder
	b.WriteString("# Azure Resource Move Validation Report\n\n")

	fmt.Fprintf(&b, "- **Generated:** %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))
	if r.Success {
		b.WriteString("- **Status:** SUCCESS\n")
	} else {
		fmt.Fprintf(&b, "- **Status:** FAILED (%d %s)\n", len(r.Errors), pluralise("error", len(r.Errors)))
	}
	fmt.Fprintf(&b, "- **Source:** `%s` / `%s`\n", r.Context.SourceSubscriptionID, r.Context.SourceResourceGroup)
	fmt.Fprintf(&b, "- **Target:** `%s` / `%s`\n", r.Context.TargetSubscriptionID, r.Context.TargetResourceGroup)
	fmt.Fprintf(&b, "- **Resources validated:** %d\n", r.Context.ResourceCount)
	fmt.Fprintf(&b, "- **HTTP status:** %d %s\n", r.StatusCode, r.StatusText)
	if !r.Success && r.TopLevel.Code != "" {
		fmt.Fprintf(&b, "- **Top-level code:** `%s`\n", r.TopLevel.Code)
	}
	b.WriteString("\n")

	if r.Success {
		b.WriteString("No validation issues found. All resources are eligible to move.\n")
		return b.String()
	}

	if r.TopLevel.Message != "" {
		b.WriteString("> ")
		b.WriteString(r.TopLevel.Message)
		b.WriteString("\n\n")
	}

	if len(r.Errors) > 0 {
		b.WriteString("## Summary\n\n")
		b.WriteString("| # | Resource Type | Name | Code |\n")
		b.WriteString("|---|---|---|---|\n")
		for i, e := range r.Errors {
			fmt.Fprintf(&b, "| %d | %s | %s | %s |\n", i+1, mdEscape(e.ResourceType), mdEscape(e.ResourceName), mdEscape(e.Code))
		}
		b.WriteString("\n## Details\n\n")
		for i, e := range r.Errors {
			fmt.Fprintf(&b, "### %d. %s\n", i+1, e.ResourceName)
			fmt.Fprintf(&b, "- **Type:** `%s`\n", e.ResourceType)
			fmt.Fprintf(&b, "- **Resource ID:** `%s`\n", e.ResourceID)
			fmt.Fprintf(&b, "- **Code:** `%s`\n", e.Code)
			fmt.Fprintf(&b, "- **Message:** %s\n\n", e.Message)
		}
	}

	if r.RawJSON != "" {
		b.WriteString("## Raw Azure API Response\n\n")
		b.WriteString("```json\n")
		b.WriteString(r.RawJSON)
		if !strings.HasSuffix(r.RawJSON, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("```\n")
	}

	return b.String()
}

// TopFailureNames returns up to n resource names from the failed details,
// suitable for the console summary.
func TopFailureNames(r ValidationReport, n int) []string {
	if n <= 0 || len(r.Errors) == 0 {
		return nil
	}
	limit := min(n, len(r.Errors))
	names := make([]string, 0, limit)
	for i := range limit {
		names = append(names, r.Errors[i].ResourceName)
	}
	return names
}

// mdEscape escapes the pipe character inside Markdown table cells.
func mdEscape(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

func pluralise(word string, n int) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
