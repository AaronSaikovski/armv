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
	"testing"

	"github.com/AaronSaikovski/armv/pkg/utils"
)

const (
	dirPermission os.FileMode = 0750
)

func TestCheckExists(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() string
		teardown func(string)
		want     bool
	}{
		{
			name: "file exists",
			setup: func() string {
				tmpFile, _ := os.CreateTemp("", "test-*.txt")
				path := tmpFile.Name()
				tmpFile.Close()
				return path
			},
			teardown: func(path string) {
				os.Remove(path)
			},
			want: true,
		},
		{
			name: "file does not exist",
			setup: func() string {
				return filepath.Join(os.TempDir(), "non-existent-file-12345.txt")
			},
			teardown: func(path string) {},
			want:     false,
		},
		{
			name: "directory exists",
			setup: func() string {
				tmpDir, _ := os.MkdirTemp("", "test-dir-*")
				return tmpDir
			},
			teardown: func(path string) {
				os.RemoveAll(path)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			defer tt.teardown(path)

			got := utils.CheckExists(path)
			if got != tt.want {
				t.Errorf("CheckExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeFolder(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		setup      func(string)
		teardown   func(string)
		wantErr    bool
	}{
		{
			name:       "create new folder",
			folderName: filepath.Join(os.TempDir(), "test-folder-new"),
			setup:      func(path string) {},
			teardown: func(path string) {
				os.RemoveAll(path)
			},
			wantErr: false,
		},
		{
			name:       "folder already exists",
			folderName: filepath.Join(os.TempDir(), "test-folder-exists"),
			setup: func(path string) {
				os.Mkdir(path, dirPermission)
			},
			teardown: func(path string) {
				os.RemoveAll(path)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.folderName)
			defer tt.teardown(tt.folderName)

			err := utils.MakeFolder(tt.folderName)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeFolder() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify folder exists after creation
			if !tt.wantErr && !utils.CheckExists(tt.folderName) {
				t.Errorf("MakeFolder() did not create folder at %s", tt.folderName)
			}
		})
	}
}

func TestWriteOutputFile(t *testing.T) {
	tests := []struct {
		name       string
		outputPath string
		filename   string
		content    string
		setup      func(string)
		teardown   func(string)
		wantErr    bool
	}{
		{
			name:       "write file to new directory",
			outputPath: filepath.Join(os.TempDir(), "test-output-new"),
			filename:   "test.txt",
			content:    "test content",
			setup:      func(path string) {},
			teardown: func(path string) {
				os.RemoveAll(path)
			},
			wantErr: false,
		},
		{
			name:       "write file to existing directory",
			outputPath: filepath.Join(os.TempDir(), "test-output-exists"),
			filename:   "test.txt",
			content:    "test content 2",
			setup: func(path string) {
				os.Mkdir(path, dirPermission)
			},
			teardown: func(path string) {
				os.RemoveAll(path)
			},
			wantErr: false,
		},
		{
			name:       "write empty content",
			outputPath: filepath.Join(os.TempDir(), "test-output-empty"),
			filename:   "empty.txt",
			content:    "",
			setup:      func(path string) {},
			teardown: func(path string) {
				os.RemoveAll(path)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.outputPath)
			defer tt.teardown(tt.outputPath)

			err := utils.WriteOutputFile(tt.outputPath, tt.filename, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteOutputFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify file exists and content matches
				fullPath := filepath.Join(tt.outputPath, tt.filename)
				data, err := os.ReadFile(fullPath)
				if err != nil {
					t.Errorf("Failed to read written file: %v", err)
				}
				if string(data) != tt.content {
					t.Errorf("File content = %q, want %q", string(data), tt.content)
				}
			}
		})
	}
}
