// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systest

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Path = string

// goldenReqData holds the parsed golden data from TestMarkdown files
type goldenData struct {

	// Maps file paths to line numbers to goldenReqItem slices
	errors map[Path]map[int][]*regexp.Regexp
	// Golden file lines, per file
	lines map[Path][]string
}

// parseReqGoldenData parses TestMarkdown files to extract golden test data
// It takes a path to the req folder and returns a structured goldenReqData object
func parseGoldenData(reqFolderPath string) (*goldenData, error) {
	// Initialize the goldenReqData structure
	result := &goldenData{
		errors: make(map[string]map[int][]*regexp.Regexp),
		lines:  make(map[string][]string),
	}

	// Walk through the req folder to find TestMarkdown files
	files, err := listFilePaths(reqFolderPath, `.*\.md`)
	if err != nil {
		return nil, fmt.Errorf("error finding TestMarkdown files: %v", err)
	}

	for _, filePath := range files {
		// Read file contents
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
		}

		// Process the file line by line
		lines := strings.Split(string(content), "\n")

		// Store the lines in the result
		result.lines[filePath] = lines

		previousLineN := 0

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			// Skip if not a golden line
			if !strings.HasPrefix(trimmedLine, "//") {
				previousLineN = i + 1 // Store current line number for reference
				continue
			}

			// Remove the "//" prefix and trim whitespace
			goldenContent := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "//"))

			// Process errors line
			if strings.HasPrefix(goldenContent, "errors:") {
				if previousLineN == 0 {
					return nil, fmt.Errorf("errors line found without preceding content at %s:%d", filePath, i+1)
				}

				// Extract error regexes from the line
				errorPart := strings.TrimSpace(strings.TrimPrefix(goldenContent, "errors:"))
				reErrPattern := regexp.MustCompile(`"([^"]*)"`)
				matches := reErrPattern.FindAllStringSubmatch(errorPart, -1)

				// Initialize the line map if it doesn't exist
				if result.errors[filePath] == nil {
					result.errors[filePath] = make(map[int][]*regexp.Regexp)
				}

				for _, match := range matches {
					pattern := match[1]
					regex, err := regexp.Compile(pattern)
					if err != nil {
						return nil, fmt.Errorf("invalid error regex at %s:%d: %v", filePath, i+1, err)
					}
					result.errors[filePath][previousLineN] = append(result.errors[filePath][previousLineN], regex)
				}
				continue
			}
		}
	}

	return result, nil
}
