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

package utils

import "testing"

func TestSetVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "set version successfully",
			version:  "1.0.0",
			expected: "ARMV version: 1.0.0",
		},
		{
			name:     "set empty version",
			version:  "",
			expected: "ARMV version: ",
		},
		{
			name:     "set version with build info",
			version:  "1.2.3-beta+20240101",
			expected: "ARMV version: 1.2.3-beta+20240101",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVersion(tt.version)
			got := GetVersion()
			if got != tt.expected {
				t.Errorf("GetVersion() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppDescription(t *testing.T) {
	if AppDescription == "" {
		t.Error("AppDescription should not be empty")
	}
	// Check if description contains key information
	expectedSubstring := "Azure Resource Movability Validator"
	if len(AppDescription) < len(expectedSubstring) {
		t.Errorf("AppDescription is too short, got length %d", len(AppDescription))
	}
}

func TestArgsStruct(t *testing.T) {
	args := Args{
		SourceSubscriptionId: "12345678-1234-1234-1234-123456789012",
		SourceResourceGroup:  "source-rg",
		TargetSubscriptionId: "87654321-4321-4321-4321-210987654321",
		TargetResourceGroup:  "target-rg",
		Debug:                true,
		OutputPath:           "./output",
	}

	if args.SourceSubscriptionId != "12345678-1234-1234-1234-123456789012" {
		t.Error("SourceSubscriptionId not set correctly")
	}
	if args.SourceResourceGroup != "source-rg" {
		t.Error("SourceResourceGroup not set correctly")
	}
	if args.TargetSubscriptionId != "87654321-4321-4321-4321-210987654321" {
		t.Error("TargetSubscriptionId not set correctly")
	}
	if args.TargetResourceGroup != "target-rg" {
		t.Error("TargetResourceGroup not set correctly")
	}
	if !args.Debug {
		t.Error("Debug not set correctly")
	}
	if args.OutputPath != "./output" {
		t.Error("OutputPath not set correctly")
	}
}
