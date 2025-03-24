// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systest

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// parseReqGoldenData parses TestMarkdown files to extract golden test data
// It takes a path to the req folder and returns a structured goldenReqData object
func parseReqGoldenData(reqFolderPath string) (*goldenReqData, error) {
	// Initialize the goldenReqData structure
	result := &goldenReqData{
		errors:       make(map[string]map[int][]*goldenReqItem),
		reqsites:     make(map[string]map[int][]*goldenReqItem),
		footnotes:    make(map[string]map[int][]*goldenReqItem),
		newfootnotes: make(map[string][]*goldenReqItem),
		lines:        make(map[string][]string),
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

				for _, match := range matches {
					pattern := match[1]
					regex, err := regexp.Compile(pattern)
					if err != nil {
						return nil, fmt.Errorf("invalid error regex at %s:%d: %v", filePath, i+1, err)
					}

					item := &goldenReqItem{
						regex: regex,
					}

					// Initialize the line map if it doesn't exist
					if result.errors[filePath] == nil {
						result.errors[filePath] = make(map[int][]*goldenReqItem)
					}
					result.errors[filePath][previousLineN] = append(result.errors[filePath][previousLineN], item)
				}
				continue
			}

			// Process reqsite line
			if strings.HasPrefix(goldenContent, "reqsite:") {
				if previousLineN == 0 {
					return nil, fmt.Errorf("reqsite line found without preceding content at %s:%d", filePath, i+1)
				}

				data := strings.TrimSpace(strings.TrimPrefix(goldenContent, "reqsite:"))
				// Replace backticks with double quotes
				data = strings.ReplaceAll(data, "`", "\"")

				item := &goldenReqItem{
					data: data,
				}

				// Initialize the line map if it doesn't exist
				if result.reqsites[filePath] == nil {
					result.reqsites[filePath] = make(map[int][]*goldenReqItem)
				}
				result.reqsites[filePath][previousLineN] = append(result.reqsites[filePath][previousLineN], item)
				continue
			}

			// Process footnote line
			if strings.HasPrefix(goldenContent, "footnote:") {
				if previousLineN == 0 {
					return nil, fmt.Errorf("footnote line found without preceding content at %s:%d", filePath, i+1)
				}

				data := strings.TrimSpace(strings.TrimPrefix(goldenContent, "footnote:"))
				// Replace backticks with double quotes
				data = strings.ReplaceAll(data, "`", "\"")

				item := &goldenReqItem{
					data: data,
				}

				// Initialize the line map if it doesn't exist
				if result.footnotes[filePath] == nil {
					result.footnotes[filePath] = make(map[int][]*goldenReqItem)
				}
				result.footnotes[filePath][previousLineN] = append(result.footnotes[filePath][previousLineN], item)
				continue
			}

			// Process newfootnote line
			if strings.HasPrefix(goldenContent, "newfootnote:") {
				data := strings.TrimSpace(strings.TrimPrefix(goldenContent, "newfootnote:"))
				// Replace backticks with double quotes
				data = strings.ReplaceAll(data, "`", "\"")

				item := &goldenReqItem{
					data: data,
				}

				result.newfootnotes[filePath] = append(result.newfootnotes[filePath], item)
				continue
			}
		}
	}

	return result, nil
}

// goldenReqData holds the parsed golden data from TestMarkdown files
type goldenReqData struct {

	// Maps file paths to line numbers to goldenReqItem slices
	errors    map[string]map[int][]*goldenReqItem
	reqsites  map[string]map[int][]*goldenReqItem
	footnotes map[string]map[int][]*goldenReqItem

	// Maps file paths to the expected tail of the file
	newfootnotes map[string][]*goldenReqItem

	// Maps file paths to lines
	lines map[string][]string
}

// goldenReqItem represents a single golden data item with its context
type goldenReqItem struct {
	data  string         // Content for reqsites, footnotes, and newfootnotes
	regex *regexp.Regexp // Compiled regex for errors
}
