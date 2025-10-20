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

	"github.com/AaronSaikovski/armv/cmd/armv/app"
)

func TestNewRootCommand(t *testing.T) {
	version := "1.0.0-test"

	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "create root command with version",
			version: version,
		},
		{
			name:    "create root command with empty version",
			version: "",
		},
		{
			name:    "create root command with dev version",
			version: "dev-build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := app.NewRootCommand(tt.version)

			if cmd == nil {
				t.Fatal("NewRootCommand() returned nil")
			}

			if cmd.Use != "armv" {
				t.Errorf("Use = %v, want %v", cmd.Use, "armv")
			}

			if cmd.Short == "" {
				t.Error("Short description should not be empty")
			}

			if cmd.Long == "" {
				t.Error("Long description should not be empty")
			}

			if cmd.Version != tt.version {
				t.Errorf("Version = %v, want %v", cmd.Version, tt.version)
			}

			// Check that required flags exist
			requiredFlags := []string{
				"source-subscription-id",
				"source-resource-group",
				"target-subscription-id",
				"target-resource-group",
			}

			for _, flagName := range requiredFlags {
				flag := cmd.Flags().Lookup(flagName)
				if flag == nil {
					t.Errorf("Required flag %s not found", flagName)
				}
			}

			// Check that optional flags exist
			optionalFlags := []string{"debug", "output-path"}
			for _, flagName := range optionalFlags {
				flag := cmd.Flags().Lookup(flagName)
				if flag == nil {
					t.Errorf("Optional flag %s not found", flagName)
				}
			}
		})
	}
}

func TestRootCommandFlags(t *testing.T) {
	cmd := app.NewRootCommand("test-version")

	tests := []struct {
		name         string
		flagName     string
		flagType     string
		shouldExist  bool
		defaultValue string
	}{
		{
			name:        "source-subscription-id exists",
			flagName:    "source-subscription-id",
			flagType:    "string",
			shouldExist: true,
		},
		{
			name:        "source-resource-group exists",
			flagName:    "source-resource-group",
			flagType:    "string",
			shouldExist: true,
		},
		{
			name:        "target-subscription-id exists",
			flagName:    "target-subscription-id",
			flagType:    "string",
			shouldExist: true,
		},
		{
			name:        "target-resource-group exists",
			flagName:    "target-resource-group",
			flagType:    "string",
			shouldExist: true,
		},
		{
			name:        "debug flag exists",
			flagName:    "debug",
			flagType:    "bool",
			shouldExist: true,
		},
		{
			name:         "output-path has default",
			flagName:     "output-path",
			flagType:     "string",
			shouldExist:  true,
			defaultValue: app.DefaultOutputPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)

			if tt.shouldExist && flag == nil {
				t.Errorf("Flag %s should exist but was not found", tt.flagName)
				return
			}

			if !tt.shouldExist && flag != nil {
				t.Errorf("Flag %s should not exist but was found", tt.flagName)
				return
			}

			if flag != nil && tt.flagType != "" && flag.Value.Type() != tt.flagType {
				t.Errorf("Flag %s type = %v, want %v", tt.flagName, flag.Value.Type(), tt.flagType)
			}

			if tt.defaultValue != "" && flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %s default = %v, want %v", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}
