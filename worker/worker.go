package worker

// worker package is responsible for searching for a string in a file.

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// Result represents a matching result containing the line, line number, and file path.
type Result struct {
	Line    string
	LineNum int
	Path    string
}

// Results represents a collection of matching results.
type Results struct {
	Inner []Result
}

// NewResult creates a new Result instance.
func NewResult(line string, lineNum int, path string) Result {
	return Result{line, lineNum, path}
}

// FindInFile searches for occurrences of the specified string in a file.
// It returns a Results object containing matching results.
func FindInFile(path string, find string) *Results {
	// Open the file for reading.
	file, err := os.Open(path)
	if err != nil {
		log.Println("Error:", err)
		return nil
	}

	// Initialize the results container.
	results := Results{make([]Result, 0)}

	// Create a scanner to read the file line by line.
	scanner := bufio.NewScanner(file)
	lineNum := 1

	// Iterate through each line in the file.
	for scanner.Scan() {
		// Check if the line contains the specified string.
		if strings.Contains(scanner.Text(), find) {
			// Create a new Result and add it to the results container.
			r := NewResult(scanner.Text(), lineNum, path)
			results.Inner = append(results.Inner, r)
		}
		lineNum++
	}

	// Close the file.
	file.Close()

	// Check if any results were found.
	if len(results.Inner) == 0 {
		return nil
	}
	return &results

}
