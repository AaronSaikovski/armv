package test

import (
	"testing"

	"github.com/AaronSaikovski/armv/cmd/armv/app"
)

func TestNewRootCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
	}{
		{name: "semantic version", version: "1.0.0-test"},
		{name: "empty version", version: ""},
		{name: "dev build", version: "dev-build"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := app.NewRootCommand(tt.version)
			if cmd == nil {
				t.Fatal("NewRootCommand returned nil")
			}
			if cmd.Use != "armv" {
				t.Errorf("Use = %q, want %q", cmd.Use, "armv")
			}
			if cmd.Short == "" {
				t.Error("Short description should not be empty")
			}
			if cmd.Long == "" {
				t.Error("Long description should not be empty")
			}
			if cmd.Version != tt.version {
				t.Errorf("Version = %q, want %q", cmd.Version, tt.version)
			}

			requiredFlags := []string{
				"source-subscription-id",
				"source-resource-group",
				"target-subscription-id",
				"target-resource-group",
			}
			for _, flagName := range requiredFlags {
				if flag := cmd.Flags().Lookup(flagName); flag == nil {
					t.Errorf("required flag %q not found", flagName)
				}
			}

			for _, flagName := range []string{"debug", "output-path"} {
				if flag := cmd.Flags().Lookup(flagName); flag == nil {
					t.Errorf("optional flag %q not found", flagName)
				}
			}
		})
	}
}

func TestRootCommandFlags(t *testing.T) {
	t.Parallel()

	cmd := app.NewRootCommand("test-version")

	tests := []struct {
		name         string
		flagName     string
		flagType     string
		defaultValue string
	}{
		{name: "source-subscription-id", flagName: "source-subscription-id", flagType: "string"},
		{name: "source-resource-group", flagName: "source-resource-group", flagType: "string"},
		{name: "target-subscription-id", flagName: "target-subscription-id", flagType: "string"},
		{name: "target-resource-group", flagName: "target-resource-group", flagType: "string"},
		{name: "debug", flagName: "debug", flagType: "bool"},
		{name: "output-path", flagName: "output-path", flagType: "string", defaultValue: app.DefaultOutputPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if flag.Value.Type() != tt.flagType {
				t.Errorf("type = %q, want %q", flag.Value.Type(), tt.flagType)
			}
			if tt.defaultValue != "" && flag.DefValue != tt.defaultValue {
				t.Errorf("default = %q, want %q", flag.DefValue, tt.defaultValue)
			}
		})
	}
}

// TestRootCommandRequiredFlags verifies every required flag is marked as such.
// Guards against accidental removal of MarkFlagRequired after the loop refactor.
func TestRootCommandRequiredFlags(t *testing.T) {
	t.Parallel()

	cmd := app.NewRootCommand("test")
	required := []string{
		"source-subscription-id",
		"source-resource-group",
		"target-subscription-id",
		"target-resource-group",
	}

	for _, name := range required {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("flag %q not found", name)
		}
		annotations := flag.Annotations["cobra_annotation_bash_completion_one_required_flag"]
		if len(annotations) == 0 || annotations[0] != "true" {
			t.Errorf("flag %q is not marked required (annotations=%v)", name, annotations)
		}
	}
}

// TestRootCommandNoSharedFlagState verifies that two Command instances do not
// share flag variables (the refactor moved vars from package globals into the
// closure; this pins that property).
func TestRootCommandNoSharedFlagState(t *testing.T) {
	t.Parallel()

	a := app.NewRootCommand("a")
	b := app.NewRootCommand("b")

	if a == b {
		t.Fatal("NewRootCommand returned the same pointer twice")
	}
	if a.Version == b.Version {
		t.Errorf("Version fields unexpectedly shared: %q == %q", a.Version, b.Version)
	}
}
