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

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DirPermission grants owner rwx and group r-x on created directories.
	DirPermission os.FileMode = 0750
	// FilePermission grants owner rw- and group r-- on created files.
	FilePermission os.FileMode = 0640
)

// CheckExists reports whether a file or directory exists at filePath.
func CheckExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

// MakeFolder creates folderName if it does not already exist. It is safe
// against concurrent callers because os.MkdirAll is idempotent.
func MakeFolder(folderName string) error {
	if err := os.MkdirAll(folderName, DirPermission); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", folderName, err)
	}
	return nil
}

// WriteOutputFile writes output to outputPath/filename, creating the directory
// if necessary.
func WriteOutputFile(outputPath string, filename string, output string) error {
	if err := MakeFolder(outputPath); err != nil {
		return fmt.Errorf("output directory creation failed: %w", err)
	}

	fullPath := filepath.Join(outputPath, filename)
	if err := os.WriteFile(fullPath, []byte(output), FilePermission); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	return nil
}
