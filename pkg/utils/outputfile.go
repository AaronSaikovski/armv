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
	"os"
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

// deleteFile deletes a file at the specified filePath if it exists.
//
// filePath: the path to the file to be deleted.
func DeleteFile(filePath string) {
	if CheckExists(filePath) {
		err := os.Remove(filePath)
		if err != nil {
			HandleError(err)
		}
	}
}

// MakeFolder creates a new folder with the specified folderName.
//
// folderName: the name of the folder to be created.
func MakeFolder(folderName string) {

	if !CheckExists(folderName) {
		errFolder := os.Mkdir(folderName, 0755)
		if errFolder != nil {
			HandleError(errFolder)
		}
	}

}

// WriteOutputFile writes the output to a file with the specified filename.
//
// filename: the name of the file to write the output to.
// output: the content to be written to the file.
func WriteOutputFile(outputPath string, filename string, output string) {

	//create the folder if it doesn't exist
	MakeFolder(outputPath)

	//write the output to the file
	errWrite := os.WriteFile(outputPath+"/"+filename, []byte(output), 0644)
	if errWrite != nil {
		HandleError(errWrite)
	}

}
