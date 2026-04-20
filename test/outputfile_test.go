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
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/AaronSaikovski/armv/pkg/utils"
)

func TestCheckExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T) string
		want  bool
	}{
		{
			name: "file exists",
			setup: func(t *testing.T) string {
				t.Helper()
				f, err := os.CreateTemp(t.TempDir(), "test-*.txt")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				_ = f.Close()
				return f.Name()
			},
			want: true,
		},
		{
			name: "file does not exist",
			setup: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "non-existent.txt")
			},
			want: false,
		},
		{
			name: "directory exists",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := tt.setup(t)
			if got := utils.CheckExists(path); got != tt.want {
				t.Errorf("CheckExists(%q) = %v, want %v", path, got, tt.want)
			}
		})
	}
}

func TestMakeFolder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T, path string)
	}{
		{
			name:  "create new folder",
			setup: func(t *testing.T, path string) {},
		},
		{
			name: "folder already exists",
			setup: func(t *testing.T, path string) {
				t.Helper()
				if err := os.Mkdir(path, utils.DirPermission); err != nil {
					t.Fatalf("pre-create: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "target")
			tt.setup(t, path)

			if err := utils.MakeFolder(path); err != nil {
				t.Fatalf("MakeFolder: %v", err)
			}
			if !utils.CheckExists(path) {
				t.Errorf("MakeFolder did not create %q", path)
			}
		})
	}
}

func TestMakeFolderIdempotent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "idem")
	for i := range 3 {
		if err := utils.MakeFolder(path); err != nil {
			t.Fatalf("iter %d MakeFolder: %v", i, err)
		}
	}
}

func TestMakeFolderCreatesNestedPath(t *testing.T) {
	t.Parallel()

	// MakeFolder uses os.MkdirAll so intermediate dirs must be created too.
	nested := filepath.Join(t.TempDir(), "a", "b", "c")
	if err := utils.MakeFolder(nested); err != nil {
		t.Fatalf("MakeFolder(%q): %v", nested, err)
	}
	if !utils.CheckExists(nested) {
		t.Errorf("nested path %q not created", nested)
	}
}

func TestWriteOutputFileOverwritesExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	const name = "report.txt"

	if err := utils.WriteOutputFile(dir, name, "first"); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := utils.WriteOutputFile(dir, name, "second"); err != nil {
		t.Fatalf("second write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "second" {
		t.Errorf("content = %q, want %q (file should be overwritten)", string(data), "second")
	}
}

func TestPermissionConstants(t *testing.T) {
	t.Parallel()

	// Pin the hardened permission values. A regression here widens the blast
	// radius of generated output files or directories.
	if utils.DirPermission != 0o750 {
		t.Errorf("DirPermission = %o, want 0750", utils.DirPermission)
	}
	if utils.FilePermission != 0o640 {
		t.Errorf("FilePermission = %o, want 0640", utils.FilePermission)
	}
}

func TestWriteOutputFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{name: "new directory", filename: "test.txt", content: "test content"},
		{name: "empty content", filename: "empty.txt", content: ""},
		{name: "multiline content", filename: "multi.txt", content: "line1\nline2\nline3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			outputPath := filepath.Join(t.TempDir(), "out")

			if err := utils.WriteOutputFile(outputPath, tt.filename, tt.content); err != nil {
				t.Fatalf("WriteOutputFile: %v", err)
			}

			fullPath := filepath.Join(outputPath, tt.filename)
			data, err := os.ReadFile(fullPath)
			if err != nil {
				t.Fatalf("ReadFile(%q): %v", fullPath, err)
			}
			if string(data) != tt.content {
				t.Errorf("content = %q, want %q", string(data), tt.content)
			}

			if runtime.GOOS != "windows" {
				info, err := os.Stat(fullPath)
				if err != nil {
					t.Fatalf("Stat(%q): %v", fullPath, err)
				}
				if perm := info.Mode().Perm(); perm != utils.FilePermission {
					t.Errorf("permission = %o, want %o", perm, utils.FilePermission)
				}
			}
		})
	}
}
