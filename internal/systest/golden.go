// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Path = string

// goldenData holds the parsed golden data from TestMarkdown files
type goldenData struct {
	// Maps file paths to line numbers to goldenReqItem slices
	errors map[Path]map[int][]*regexp.Regexp
	// Golden file lines, per file
	lines map[Path][]string
}

// parseGoldenData
//
// Definitions
// - RootName is a file name without extension
// - GoldenFile is a file whose RootName ends with "_"
// - NormalFile is a file whose RootName does not end with "_"
// - NormalizedPath is the path with "_" removed from the RootName
//
// Description
// - Takes the path to the `req` folder as a parameter
// - NormalFiles that ends with ".md" are processed to extract GoldenErrors (see below)
// - NormalFiles that do not have GoldenFile counterpart are loaded to goldenData.lines
// - GoldenFiles are loaded to goldenData.lines, path is normalized ("_" is removed)
// - Processing of goldenData.lines:
//   - For each NormalFile without a GoldenFile counterpart, read the file content line by line and store in goldenData.lines[normalizedPath]
//   - For each GoldenFile, read the file content line by line and store in goldenData.lines[normalizedPath]
//   - The lines are stored in the same order as they appear in the file
//   - Empty lines and whitespace are preserved exactly as they appear in the files
func parseGoldenData(reqFolderPath string) (*goldenData, error) {
	// Initialize the goldenData structure
	result := &goldenData{
		errors: make(map[string]map[int][]*regexp.Regexp),
		lines:  make(map[string][]string),
	}

	// Walk through the req folder to find TestMarkdown files
	files, err := listFilePaths(reqFolderPath, `.*\.md`)
	if err != nil {
		return nil, fmt.Errorf("error finding TestMarkdown files: %v", err)
	}

	// Separate files into golden files and normal files
	goldenFiles := make(map[string]string) // normalized path -> file path
	normalFiles := make(map[string]string) // normalized path -> file path

	for _, filePath := range files {
		// Extract the filename from the path
		fileName := filepath.Base(filePath)
		ext := filepath.Ext(fileName)
		rootName := fileName[:len(fileName)-len(ext)]

		// Determine if this is a GoldenFile (ends with "_") or NormalFile
		isGoldenFile := strings.HasSuffix(rootName, "_")

		// Determine the normalized path (remove "_" for golden files)
		normalizedPath := filePath
		if isGoldenFile {
			// Remove the "_" from the root name
			normalizedRootName := rootName[:len(rootName)-1]
			normalizedFileName := normalizedRootName + ext
			normalizedPath = filepath.Join(filepath.Dir(filePath), normalizedFileName)
		}

		// Store in appropriate map
		if isGoldenFile {
			goldenFiles[normalizedPath] = filePath
		} else {
			normalFiles[normalizedPath] = filePath
		}
	}

	// Process normal files to extract errors
	for normalizedPath, filePath := range normalFiles {
		// Read file contents
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
		}

		// Process the file line by line
		lines := strings.Split(string(content), "\n")

		// Extract errors from the file
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

		// If no golden counterpart exists, store the normal file's lines
		if _, hasGoldenCounterpart := goldenFiles[normalizedPath]; !hasGoldenCounterpart {
			result.lines[normalizedPath] = lines
		}
	}

	// Process golden files
	for normalizedPath, goldenFilePath := range goldenFiles {
		content, err := os.ReadFile(goldenFilePath)
		if err != nil {
			return nil, fmt.Errorf("error reading golden file %s: %v", goldenFilePath, err)
		}

		// Store the lines using the normalized path
		result.lines[normalizedPath] = strings.Split(string(content), "\n")
	}

	return result, nil
}
