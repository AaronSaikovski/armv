package test

import (
	"strings"
	"testing"

	"github.com/AaronSaikovski/armv/pkg/utils"
)

func TestFormatVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "semantic version", version: "1.0.0", want: "ARMV version: 1.0.0"},
		{name: "v-prefix", version: "v1.3.0", want: "ARMV version: v1.3.0"},
		{name: "empty version", version: "", want: "ARMV version: "},
		{name: "dev sentinel", version: "dev", want: "ARMV version: dev"},
		{name: "prerelease with metadata", version: "1.2.3-beta+20240101", want: "ARMV version: 1.2.3-beta+20240101"},
		{name: "composite version string", version: "v1.3.0 (commit abc1234, built 2026-04-20)", want: "ARMV version: v1.3.0 (commit abc1234, built 2026-04-20)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := utils.FormatVersion(tt.version); got != tt.want {
				t.Errorf("FormatVersion(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestAppDescription(t *testing.T) {
	t.Parallel()

	if utils.AppDescription == "" {
		t.Fatal("AppDescription should not be empty")
	}

	if !strings.Contains(utils.AppDescription, "Azure Resource Movability Validator") {
		t.Errorf("AppDescription missing product name:\n%s", utils.AppDescription)
	}
	if !strings.Contains(utils.AppDescription, "Read-Only") {
		t.Errorf("AppDescription should state the tool is Read-Only:\n%s", utils.AppDescription)
	}
	// The tool supports cross-subscription moves; the description must not
	// claim source and target must share a subscription (a prior doc bug).
	if strings.Contains(utils.AppDescription, "same subscription") {
		t.Errorf("AppDescription wrongly claims 'same subscription':\n%s", utils.AppDescription)
	}
}

func TestArgsZeroValue(t *testing.T) {
	t.Parallel()

	var a utils.Args

	if a.SourceSubscriptionId != "" || a.SourceResourceGroup != "" ||
		a.TargetSubscriptionId != "" || a.TargetResourceGroup != "" ||
		a.OutputPath != "" {
		t.Errorf("Args zero value has non-empty string fields: %+v", a)
	}
	if a.Debug {
		t.Errorf("Args.Debug zero value = true, want false")
	}
}

// TestArgsFieldAssignment pins the public field set. If a field is renamed or
// removed in a future refactor, this test fails to compile, flagging the break.
func TestArgsFieldAssignment(t *testing.T) {
	t.Parallel()

	a := utils.Args{
		SourceSubscriptionId: "a",
		SourceResourceGroup:  "b",
		TargetSubscriptionId: "c",
		TargetResourceGroup:  "d",
		Debug:                true,
		OutputPath:           "e",
	}

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"SourceSubscriptionId", a.SourceSubscriptionId, "a"},
		{"SourceResourceGroup", a.SourceResourceGroup, "b"},
		{"TargetSubscriptionId", a.TargetSubscriptionId, "c"},
		{"TargetResourceGroup", a.TargetResourceGroup, "d"},
		{"OutputPath", a.OutputPath, "e"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
	if !a.Debug {
		t.Errorf("Debug = false, want true")
	}
}
