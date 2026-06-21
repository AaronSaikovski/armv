package test

import (
	"strings"
	"testing"
	"time"

	"github.com/AaronSaikovski/armv/cmd/armv/poller"
)

func TestParseResourceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		target   string
		wantType string
		wantName string
	}{
		{
			name:     "valid full Azure resource ID",
			target:   "/subscriptions/abc/resourceGroups/src-rsg/providers/Microsoft.ContainerInstance/containerGroups/aciresource",
			wantType: "Microsoft.ContainerInstance/containerGroups",
			wantName: "aciresource",
		},
		{
			name:     "child resource with extra path segments",
			target:   "/subscriptions/abc/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/mysa/blobServices/default",
			wantType: "Microsoft.Storage/storageAccounts",
			wantName: "default",
		},
		{
			name:     "missing providers segment falls back to raw target",
			target:   "not-a-resource-id",
			wantType: "not-a-resource-id",
			wantName: "not-a-resource-id",
		},
		{
			name:     "empty target returns empty",
			target:   "",
			wantType: "",
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotType, gotName := poller.ParseResourceID(tt.target)
			if gotType != tt.wantType {
				t.Errorf("ParseResourceID type = %q, want %q", gotType, tt.wantType)
			}
			if gotName != tt.wantName {
				t.Errorf("ParseResourceID name = %q, want %q", gotName, tt.wantName)
			}
		})
	}
}

func TestRenderMarkdown_Success(t *testing.T) {
	t.Parallel()

	report := poller.ValidationReport{
		Success:     true,
		GeneratedAt: time.Date(2026, 4, 20, 10, 45, 12, 0, time.UTC),
		StatusCode:  204,
		StatusText:  "No Content",
		Context: poller.ReportContext{
			SourceSubscriptionID: "sub-src",
			SourceResourceGroup:  "rg-src",
			TargetSubscriptionID: "sub-dst",
			TargetResourceGroup:  "rg-dst",
			ResourceCount:        12,
		},
	}

	md := poller.RenderMarkdown(report)

	if !strings.Contains(md, "# Azure Resource Move Validation Report") {
		t.Error("expected report heading")
	}
	if !strings.Contains(md, "**Status:** SUCCESS") {
		t.Errorf("expected SUCCESS status, got:\n%s", md)
	}
	if !strings.Contains(md, "rg-src") || !strings.Contains(md, "rg-dst") {
		t.Error("expected source and target resource groups in header")
	}
	if !strings.Contains(md, "No validation issues found") {
		t.Error("expected success body text")
	}
	if strings.Contains(md, "## Details") {
		t.Error("success report should not contain a Details section")
	}
	if strings.Contains(md, "## Summary") {
		t.Error("success report should not contain a Summary section")
	}
}

func TestRenderMarkdown_Failure_FromRawBody(t *testing.T) {
	t.Parallel()

	rawBody := []byte(`{
  "error": {
    "code": "ResourceMoveValidationFailed",
    "message": "Move validation failed.",
    "details": [
      {
        "code": "ResourceMoveNotSupported",
        "target": "/subscriptions/abc/resourceGroups/src-rsg/providers/Microsoft.ContainerInstance/containerGroups/aciresource",
        "message": "Resource move is not supported for resource types 'Microsoft.ContainerInstance/containerGroups'."
      }
    ]
  }
}`)

	report := poller.BuildValidationReport(409, "Conflict", rawBody, string(rawBody), poller.ReportContext{
		SourceSubscriptionID: "sub-src",
		SourceResourceGroup:  "rg-src",
		TargetSubscriptionID: "sub-dst",
		TargetResourceGroup:  "rg-dst",
		ResourceCount:        3,
	})

	if report.Success {
		t.Fatal("expected Success=false for 409")
	}
	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 parsed error, got %d", len(report.Errors))
	}
	if report.Errors[0].ResourceType != "Microsoft.ContainerInstance/containerGroups" {
		t.Errorf("unexpected resource type: %q", report.Errors[0].ResourceType)
	}
	if report.Errors[0].ResourceName != "aciresource" {
		t.Errorf("unexpected resource name: %q", report.Errors[0].ResourceName)
	}
	if report.TopLevel.Code != "ResourceMoveValidationFailed" {
		t.Errorf("unexpected top-level code: %q", report.TopLevel.Code)
	}

	md := poller.RenderMarkdown(report)

	mustContain := []string{
		"**Status:** FAILED (1 error)",
		"**Top-level code:** `ResourceMoveValidationFailed`",
		"## Summary",
		"| 1 | Microsoft.ContainerInstance/containerGroups | aciresource | ResourceMoveNotSupported |",
		"## Details",
		"### 1. aciresource",
		"```json",
	}
	for _, s := range mustContain {
		if !strings.Contains(md, s) {
			t.Errorf("expected rendered markdown to contain %q, got:\n%s", s, md)
		}
	}
}

func TestRenderMarkdown_Failure_MalformedJSONDoesNotPanic(t *testing.T) {
	t.Parallel()

	raw := []byte(`not json`)
	report := poller.BuildValidationReport(409, "Conflict", raw, string(raw), poller.ReportContext{})

	// No parsed errors, but rendering must still succeed and include the raw body.
	md := poller.RenderMarkdown(report)
	if !strings.Contains(md, "**Status:** FAILED") {
		t.Error("expected FAILED status even when JSON is unparseable")
	}
	if !strings.Contains(md, "## Raw Azure API Response") {
		t.Error("expected raw response section as a fallback")
	}
}

func TestRenderMarkdown_Failure_MultipleErrors_PluralisesAndTabulates(t *testing.T) {
	t.Parallel()

	rawBody := []byte(`{
  "error": {
    "code": "ResourceMoveValidationFailed",
    "message": "Multiple failures.",
    "details": [
      {"code": "C1", "target": "/subscriptions/s/resourceGroups/r/providers/P/T/one",   "message": "m1"},
      {"code": "C2", "target": "/subscriptions/s/resourceGroups/r/providers/P/T/two",   "message": "m2"},
      {"code": "C3", "target": "/subscriptions/s/resourceGroups/r/providers/P/T/three", "message": "m3"}
    ]
  }
}`)
	report := poller.BuildValidationReport(409, "Conflict", rawBody, string(rawBody), poller.ReportContext{})

	md := poller.RenderMarkdown(report)

	if !strings.Contains(md, "**Status:** FAILED (3 errors)") {
		t.Errorf("expected plural 'errors' for 3 failures, got:\n%s", md)
	}
	for _, name := range []string{"one", "two", "three"} {
		if !strings.Contains(md, "### ") || !strings.Contains(md, name) {
			t.Errorf("expected details section for %q, got:\n%s", name, md)
		}
	}
}

func TestTopFailureNames(t *testing.T) {
	t.Parallel()

	report := poller.ValidationReport{
		Errors: []poller.ValidationError{
			{ResourceName: "a"},
			{ResourceName: "b"},
			{ResourceName: "c"},
			{ResourceName: "d"},
		},
	}

	got := poller.TopFailureNames(report, 3)
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("len=%d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}

	if poller.TopFailureNames(report, 0) != nil {
		t.Error("n=0 should return nil")
	}
	if poller.TopFailureNames(poller.ValidationReport{}, 3) != nil {
		t.Error("empty report should return nil")
	}
}
