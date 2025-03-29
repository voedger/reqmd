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
// Ref. design.md, the "System tests" section
func parseGoldenData(reqFolderPath string) (*goldenData, error) {
	gd := &goldenData{
		errors: make(map[Path]map[int][]*regexp.Regexp),
		lines:  make(map[Path][]string),
	}

	// Get all files in the reqFolderPath
	files, err := listFilePaths(reqFolderPath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}

	// Group files by their normalized paths
	normalizedPathMap := make(map[string][]string)
	for _, path := range files {
		relPath, err := filepath.Rel(reqFolderPath, path)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path: %v", err)
		}

		relPath = filepath.ToSlash(relPath)
		normalizedPath := getNormalizedPath(relPath)
		normalizedPathMap[normalizedPath] = append(normalizedPathMap[normalizedPath], path)
	}

	// Process each file
	for normalizedPath, filePaths := range normalizedPathMap {
		// Check if there's a golden file
		hasGoldenFile := false
		for _, path := range filePaths {
			if isGoldenFile(path) {
				hasGoldenFile = true
				break
			}
		}

		// Process normal files
		for _, path := range filePaths {
			if !isGoldenFile(path) {
				// If it's a Markdown file, process it to extract errors
				if filepath.Ext(path) == ".md" {
					if err := extractGoldenErrors(path, gd); err != nil {
						return nil, fmt.Errorf("failed to extract golden errors: %v", err)
					}
				}

				// If there's no golden file counterpart, load lines
				if !hasGoldenFile {
					if err := loadFileLines(path, normalizedPath, gd); err != nil {
						return nil, fmt.Errorf("failed to load file lines: %v", err)
					}
				}
			} else {
				// Load golden file lines
				if err := loadFileLines(path, normalizedPath, gd); err != nil {
					return nil, fmt.Errorf("failed to load golden file lines: %v", err)
				}
			}
		}
	}

	return gd, nil
}

// getNormalizedPath removes the "_" from the RootName (file name without extension)
func getNormalizedPath(path string) string {
	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	rootName := strings.TrimSuffix(filename, ext)

	// Remove trailing "_" from the root name
	normalizedRootName := strings.TrimSuffix(rootName, "_")

	if dir == "." {
		return normalizedRootName + ext
	}
	return filepath.ToSlash(filepath.Join(dir, normalizedRootName+ext))
}

// isGoldenFile checks if a file is a golden file (ends with "_" before extension)
func isGoldenFile(path string) bool {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	rootName := strings.TrimSuffix(filename, ext)
	return strings.HasSuffix(rootName, "_")
}

// loadFileLines loads the lines from a file into goldenData.lines
func loadFileLines(filePath, normalizedPath string, gd *goldenData) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Split content into lines, preserving exact whitespace
	lines := strings.Split(string(content), "\n")
	gd.lines[normalizedPath] = lines

	return nil
}

// extractGoldenErrors extracts error patterns from markdown files
func extractGoldenErrors(filePath string, gd *goldenData) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	// Regular expression for golden error lines: "// errors: "regex" "regex" ..."
	errLineRegex := regexp.MustCompile(`^\s*//\s*errors:\s*(.*)$`)

	for i := 0; i < len(lines); i++ {
		matches := errLineRegex.FindStringSubmatch(lines[i])
		if len(matches) > 1 {
			// This is a golden error line, process it for the previous line
			if i == 0 {
				return fmt.Errorf("golden error line found at the beginning of the file")
			}

			// Extract error regexes
			errorRegexes := extractErrorRegexes(matches[1])
			if len(errorRegexes) == 0 {
				continue
			}

			// Compile the regexes
			compiledRegexes := make([]*regexp.Regexp, 0, len(errorRegexes))
			for _, regex := range errorRegexes {
				compiled, err := regexp.Compile(regex)
				if err != nil {
					return fmt.Errorf("invalid error regex '%s': %v", regex, err)
				}
				compiledRegexes = append(compiledRegexes, compiled)
			}

			// Store the error regexes for the previous line
			if gd.errors[filePath] == nil {
				gd.errors[filePath] = make(map[int][]*regexp.Regexp)
			}
			gd.errors[filePath][i] = compiledRegexes
		}
	}

	return nil
}

// extractErrorRegexes parses a string containing quoted regex patterns
func extractErrorRegexes(s string) []string {
	var regexes []string
	var inQuote bool
	var currentRegex strings.Builder

	for _, r := range s {
		if r == '"' {
			inQuote = !inQuote
			if !inQuote {
				// End of a regex
				regexStr := currentRegex.String()
				if regexStr != "" {
					regexes = append(regexes, regexStr)
				}
				currentRegex.Reset()
			}
		} else if inQuote {
			currentRegex.WriteRune(r)
		}
	}

	return regexes
}
