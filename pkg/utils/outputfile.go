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
