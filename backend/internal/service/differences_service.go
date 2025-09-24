package service

import (
	"fmt"
	"os"

	"github.com/nsf/jsondiff"
)

type DifferencesService struct {
	// if reading from db, would add here
}

func NewDifferencesServicet() *DifferencesService {
	return &DifferencesService{}
}

// GetDifferences reads files from disk and creates a difference between them
//
// Summary: Reads files from disk and compares them using github.com/nsf/jsondiff
// Path Params:
//   - file1path: string (path to file 1)
//   - file2path: string (path to file 2)
//
// Responses:
//   - string: difference between the two files {diffStatus": "FullMatch"|"SupersetMatch"|"NoMatch"|"FirstArgIsInvalidJson"|"SecondArgIsInvalidJson"|"BothArgsAreInvalidJson"|"Invalid"|"" if error occurs}
//   - string: explanation of the difference {Color Coded Differences String | "" if error occurs}
//   - error: error if the files cannot be read {nil | error}
func (service *DifferencesService) GetDifferences(file1Path string, file2Path string) (string, string, error) {
	// Read both files
	file1, err := os.ReadFile(file1Path)
	if err != nil {
		// Do not include file path in response to protect against domain traversal attempts!!
		return "", "", fmt.Errorf("Failed to read contents of file1: %v", err.Error())
	}
	file2, err := os.ReadFile(file2Path)
	if err != nil {
		return "", "", fmt.Errorf("Failed to read contents of file2: %v", err.Error())
	}

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(file1, file2, &opts)
	return diff.String(), explanation, nil

}

// Excluded Functionality:
// - writing checked differences to db
// - checking if difference already exists
