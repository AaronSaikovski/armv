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

package test

import (
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

	const expectedSubstring = "Azure Resource Movability Validator"
	if len(utils.AppDescription) < len(expectedSubstring) {
		t.Errorf("AppDescription is too short, got length %d", len(utils.AppDescription))
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
