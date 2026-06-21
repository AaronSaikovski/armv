package poller

import (
	"testing"
)

func TestMdEscape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "no pipes", in: "normal text", want: "normal text"},
		{name: "single pipe", in: "has|pipe", want: `has\|pipe`},
		{name: "multiple pipes", in: "a|b|c", want: `a\|b\|c`},
		{name: "empty string", in: "", want: ""},
		{name: "only pipes", in: "|||", want: `\|\|\|`},
		{name: "pipes at boundaries", in: "|start|end|", want: `\|start\|end\|`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := mdEscape(tt.in); got != tt.want {
				t.Errorf("mdEscape(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestPluralise(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		word string
		n    int
		want string
	}{
		{name: "singular one", word: "error", n: 1, want: "error"},
		{name: "plural zero", word: "error", n: 0, want: "errors"},
		{name: "plural two", word: "error", n: 2, want: "errors"},
		{name: "plural large", word: "error", n: 100, want: "errors"},
		{name: "different word singular", word: "resource", n: 1, want: "resource"},
		{name: "different word plural", word: "resource", n: 5, want: "resources"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := pluralise(tt.word, tt.n); got != tt.want {
				t.Errorf("pluralise(%q, %d) = %q, want %q", tt.word, tt.n, got, tt.want)
			}
		})
	}
}

func TestBuildValidationReport_EmptyBody(t *testing.T) {
	t.Parallel()

	report := BuildValidationReport(409, "Conflict", []byte{}, "", ReportContext{})

	if report.Success {
		t.Error("expected Success=false for 409 even with empty body")
	}
	if len(report.Errors) != 0 {
		t.Errorf("expected 0 errors with empty body, got %d", len(report.Errors))
	}
}

func TestBuildValidationReport_SuccessNoErrors(t *testing.T) {
	t.Parallel()

	report := BuildValidationReport(204, "No Content", []byte{}, "", ReportContext{
		SourceSubscriptionID: "sub-src",
		SourceResourceGroup:  "rg-src",
		TargetSubscriptionID: "sub-dst",
		TargetResourceGroup:  "rg-dst",
		ResourceCount:        5,
	})

	if !report.Success {
		t.Error("expected Success=true for 204")
	}
	if report.StatusCode != 204 {
		t.Errorf("StatusCode = %d, want 204", report.StatusCode)
	}
	if report.StatusText != "No Content" {
		t.Errorf("StatusText = %q, want %q", report.StatusText, "No Content")
	}
}

func TestBuildValidationReport_MalformedJSON(t *testing.T) {
	t.Parallel()

	raw := []byte(`{invalid json}`)
	report := BuildValidationReport(409, "Conflict", raw, string(raw), ReportContext{})

	if report.Success {
		t.Error("expected Success=false for 409")
	}
	if len(report.Errors) != 0 {
		t.Error("expected 0 parsed errors from malformed JSON")
	}
	if report.RawJSON != string(raw) {
		t.Error("expected RawJSON to contain the original input")
	}
}

func TestBuildValidationReport_EmptyErrorDetails(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"error": {
			"code": "EmptyDetails",
			"message": "No details provided",
			"details": []
		}
	}`)
	report := BuildValidationReport(409, "Conflict", raw, string(raw), ReportContext{})

	if report.Success {
		t.Error("expected Success=false for 409")
	}
	if report.TopLevel.Code != "EmptyDetails" {
		t.Errorf("TopLevel.Code = %q, want %q", report.TopLevel.Code, "EmptyDetails")
	}
	if len(report.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(report.Errors))
	}
}

func TestBuildValidationReport_MissingErrorFields(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"error": {
			"message": "Partial error"
		}
	}`)
	report := BuildValidationReport(409, "Conflict", raw, string(raw), ReportContext{})

	if report.Success {
		t.Error("expected Success=false for 409")
	}
	if report.TopLevel.Code != "" {
		t.Errorf("TopLevel.Code = %q, want empty (missing in JSON)", report.TopLevel.Code)
	}
	if report.TopLevel.Message != "Partial error" {
		t.Errorf("TopLevel.Message = %q, want %q", report.TopLevel.Message, "Partial error")
	}
}

func TestTopFailureNames_BoundaryConditions(t *testing.T) {
	t.Parallel()

	report := ValidationReport{
		Errors: []ValidationError{
			{ResourceName: "a"},
			{ResourceName: "b"},
			{ResourceName: "c"},
		},
	}

	tests := []struct {
		name string
		n    int
		want []string
	}{
		{name: "n equals length", n: 3, want: []string{"a", "b", "c"}},
		{name: "n greater than length", n: 10, want: []string{"a", "b", "c"}},
		{name: "n equals 1", n: 1, want: []string{"a"}},
		{name: "n negative", n: -5, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TopFailureNames(report, tt.n)
			if tt.want == nil {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len=%d, want %d", len(got), len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("index %d: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseResourceID_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		target   string
		wantType string
		wantName string
	}{
		{
			name:     "providers but less than 3 parts",
			target:   "/providers/Microsoft.Compute",
			wantType: "/providers/Microsoft.Compute",
			wantName: "/providers/Microsoft.Compute",
		},
		{
			name:     "providers with exactly 3 parts",
			target:   "/providers/Microsoft.Compute/virtualMachines/myvm",
			wantType: "Microsoft.Compute/virtualMachines",
			wantName: "myvm",
		},
		{
			name:     "providers with nested resources",
			target:   "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/myvm/extensions/disk",
			wantType: "Microsoft.Compute/virtualMachines",
			wantName: "disk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotType, gotName := ParseResourceID(tt.target)
			if gotType != tt.wantType {
				t.Errorf("type = %q, want %q", gotType, tt.wantType)
			}
			if gotName != tt.wantName {
				t.Errorf("name = %q, want %q", gotName, tt.wantName)
			}
		})
	}
}
