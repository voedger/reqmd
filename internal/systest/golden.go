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

	err := filepath.Walk(reqFolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Convert path to slash format for consistency
		path = filepath.ToSlash(path)
		normalizedPath := getNormalizedPath(path)

		// Load file contents
		lines, err := loadFileLines(path)
		if err != nil {
			return fmt.Errorf("loading file %s: %v", path, err)
		}

		// Only store lines for markdown files we need to check for errors
		// and golden files we need to compare against
		if (strings.HasSuffix(strings.ToLower(path), ".md") && !isGoldenFile(path)) || isGoldenFile(path) {
			gd.lines[path] = lines
			if isGoldenFile(path) {
				gd.lines[normalizedPath] = lines
			}
		}

		// Process markdown files for golden errors
		if strings.HasSuffix(strings.ToLower(path), ".md") && !isGoldenFile(path) {
			if err := extractGoldenErrors(path, lines, gd); err != nil {
				return fmt.Errorf("extracting golden errors from %s: %v", path, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking directory %s: %v", reqFolderPath, err)
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

// loadFileLines loads and splits the file content into lines
func loadFileLines(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Split content into lines, preserving exact whitespace
	return strings.Split(string(content), "\n"), nil
}

// extractGoldenErrors extracts error patterns from markdown files
func extractGoldenErrors(filePath string, lines []string, gd *goldenData) error {
	// Regular expression for golden error lines: "// errors: "regex" "regex" ..."
	errLineRegex := regexp.MustCompile(`^\s*//\s*errors:\s*(.*)$`)

	for i := range lines {
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