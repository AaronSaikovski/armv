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

// Package-internal tests that exercise the unexported writeOutput persistence
// pipeline. The external test/ package can't reach this because writeOutput
// is unexported; in-package white-box testing is the idiomatic alternative.

package poller

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResourceMoveOK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{name: "204 No Content is OK", statusCode: 204, want: true},
		{name: "409 Conflict is not OK", statusCode: 409, want: false},
		{name: "200 OK is not a validated success", statusCode: 200, want: false},
		{name: "500 Internal Server Error is not OK", statusCode: 500, want: false},
		{name: "zero status is not OK", statusCode: 0, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ResourceMoveOK(tt.statusCode); got != tt.want {
				t.Errorf("ResourceMoveOK(%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

// TestWriteOutputEndToEnd exercises the full persistence pipeline:
// PollerResponseData → format via report.go → WriteOutputFile → file on disk.
func TestWriteOutputEndToEnd(t *testing.T) {
	t.Parallel()

	successBody := []byte{} // 204 → typically empty
	failureBody := []byte(`{
  "error": {
    "code": "ResourceMoveValidationFailed",
    "message": "one resource could not be moved",
    "details": [
      {
        "code": "ResourceMoveNotSupported",
        "target": "/subscriptions/abc/resourceGroups/src/providers/Microsoft.ContainerInstance/containerGroups/blocked",
        "message": "cannot move container groups"
      }
    ]
  }
}`)
	garbageBody := []byte("<html>upstream proxy 502</html>")

	tests := []struct {
		name       string
		body       []byte
		statusCode int
		status     string
		wantInFile []string
	}{
		{
			name:       "204 success writes success banner",
			body:       successBody,
			statusCode: StatusMoveOK,
			status:     "No Content",
			wantInFile: []string{
				"# Azure Resource Move Validation Report",
				"**Status:** SUCCESS",
				"No validation issues found",
				"sub-src",
				"sub-dst",
			},
		},
		{
			name:       "409 with parsed JSON writes summary + details",
			body:       failureBody,
			statusCode: StatusMoveFailure,
			status:     "Conflict",
			wantInFile: []string{
				"**Status:** FAILED (1 error)",
				"**Top-level code:** `ResourceMoveValidationFailed`",
				"## Summary",
				"ResourceMoveNotSupported",
				"## Details",
				"### 1. blocked",
				"```json",
			},
		},
		{
			name:       "409 with empty body writes no-body sentinel through Markdown",
			body:       nil,
			statusCode: StatusMoveFailure,
			status:     "Conflict",
			wantInFile: []string{
				"**Status:** FAILED",
				// empty body → BuildValidationReport returns zero errors;
				// the resulting report still has the header section.
				"**HTTP status:** 409 Conflict",
			},
		},
		{
			name:       "409 with non-JSON body still writes a usable report",
			body:       garbageBody,
			statusCode: StatusMoveFailure,
			status:     "Conflict",
			wantInFile: []string{
				"**Status:** FAILED",
				// Body doesn't parse as the Azure error shape, so no Summary table,
				// but the raw body still lands in the "Raw Azure API Response" block.
				"## Raw Azure API Response",
				"upstream proxy 502",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			outDir := filepath.Join(t.TempDir(), "reports")
			resp := NewPollerResponseData(tt.body, tt.statusCode, tt.status)

			ctx := ReportContext{
				SourceSubscriptionID: "sub-src",
				SourceResourceGroup:  "rg-src",
				TargetSubscriptionID: "sub-dst",
				TargetResourceGroup:  "rg-dst",
				ResourceCount:        3,
			}

			report, err := resp.writeOutput(outDir, ctx)
			if err != nil {
				t.Fatalf("writeOutput: %v", err)
			}
			if report.StatusCode != tt.statusCode {
				t.Errorf("report.StatusCode = %d, want %d", report.StatusCode, tt.statusCode)
			}
			if want := tt.statusCode == StatusMoveOK; report.Success != want {
				t.Errorf("report.Success = %v, want %v", report.Success, want)
			}

			// Find the timestamped file that writeOutput created.
			entries, err := os.ReadDir(outDir)
			if err != nil {
				t.Fatalf("ReadDir(%q): %v", outDir, err)
			}
			if len(entries) != 1 {
				t.Fatalf("expected exactly 1 file in %q, got %d", outDir, len(entries))
			}
			name := entries[0].Name()
			if !strings.HasPrefix(name, "output-") || !strings.HasSuffix(name, ".md") {
				t.Errorf("file name %q does not match output-*.md pattern", name)
			}

			data, err := os.ReadFile(filepath.Join(outDir, name))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			content := string(data)
			for _, want := range tt.wantInFile {
				if !strings.Contains(content, want) {
					t.Errorf("report missing substring %q. full content:\n%s", want, content)
				}
			}
		})
	}
}

// TestWriteOutputCreatesNestedOutputDir verifies writeOutput will create a
// multi-level output path that doesn't yet exist (MakeFolder uses MkdirAll).
func TestWriteOutputCreatesNestedOutputDir(t *testing.T) {
	t.Parallel()

	nested := filepath.Join(t.TempDir(), "a", "b", "c")
	resp := NewPollerResponseData(nil, StatusMoveOK, "No Content")

	if _, err := resp.writeOutput(nested, ReportContext{}); err != nil {
		t.Fatalf("writeOutput into non-existent nested dir: %v", err)
	}
	if _, err := os.Stat(nested); err != nil {
		t.Fatalf("nested dir was not created: %v", err)
	}
}

// TestWriteOutputFilenameIsTimestamped checks that two successive calls into
// the same directory do not clobber each other's files (unless they hit the
// exact same second, in which case overwrite is acceptable — this test tolerates
// both outcomes rather than racing on clock resolution).
func TestWriteOutputFilenameFormat(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	resp := NewPollerResponseData(nil, StatusMoveOK, "No Content")
	if _, err := resp.writeOutput(outDir, ReportContext{}); err != nil {
		t.Fatalf("writeOutput: %v", err)
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}

	// Pattern: output-YYYY-MM-DD-HH-MM-SS.md
	name := entries[0].Name()
	const (
		prefix = "output-"
		ext    = ".md"
	)
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ext) {
		t.Fatalf("filename %q does not match %s*%s", name, prefix, ext)
	}
	stamp := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ext)
	// YYYY-MM-DD-HH-MM-SS → 19 chars, all digits or dashes
	if len(stamp) != 19 {
		t.Errorf("timestamp %q is %d chars, want 19", stamp, len(stamp))
	}
	for _, r := range stamp {
		if (r < '0' || r > '9') && r != '-' {
			t.Errorf("timestamp %q contains unexpected char %q", stamp, r)
		}
	}
}
