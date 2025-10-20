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
	// Directory permission: owner rwx, group r-x, others ---
	dirPermission os.FileMode = 0750
	// File permission: owner rw-, group r--, others ---
	filePermission os.FileMode = 0640
)

// CheckExists checks if a file or folder exists at the specified filePath.
//
// It returns a boolean indicating the existence of the file.
//
// Parameters:
// - filePath: the path to the file to check.
//
// Returns:
// - a boolean indicating whether the file exists or not.
func CheckExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}

// MakeFolder creates a new folder with the specified folderName.
//
// folderName: the name of the folder to be created.
// Returns an error if folder creation fails.
func MakeFolder(folderName string) error {

	if !CheckExists(folderName) {
		if err := os.Mkdir(folderName, dirPermission); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", folderName, err)
		}
	}
	return nil
}

// WriteOutputFile writes the output to a file with the specified filename.
//
// filename: the name of the file to write the output to.
// output: the content to be written to the file.
// Returns an error if writing fails.
func WriteOutputFile(outputPath string, filename string, output string) error {

	//create the folder if it doesn't exist
	if err := MakeFolder(outputPath); err != nil {
		return fmt.Errorf("output directory creation failed: %w", err)
	}

	//write the output to the file
	fullPath := filepath.Join(outputPath, filename)
	if err := os.WriteFile(fullPath, []byte(output), filePermission); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	return nil
}
