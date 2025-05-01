// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systrun

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
	// Maps relativefile paths to line numbers to goldenReqItem slices
	errors map[Path]map[int][]*regexp.Regexp
	// Golden file lines, per file
	lines map[Path][]string
}

// parseGoldenData
// Ref. design.md, the "System tests" section
func parseGoldenData(reqFolderPaths []string) (*goldenData, error) {
	gd := &goldenData{
		errors: make(map[Path]map[int][]*regexp.Regexp),
		lines:  make(map[Path][]string),
	}

	for _, reqFolderPath := range reqFolderPaths {
		err := filepath.Walk(reqFolderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			// Get path relative to reqFolderPath
			relPath, err := filepath.Rel(reqFolderPath, path)
			if err != nil {
				return fmt.Errorf("getting relative path for %s: %v", path, err)
			}

			// Convert path to slash format for consistency
			relPath = filepath.ToSlash(relPath)
			normalizedPath := getNormalizedPath(relPath)

			// Load file contents
			lines, err := loadFileLines(path)
			if err != nil {
				return fmt.Errorf("loading file %s: %v", path, err)
			}

			// Only store lines for markdown files we need to check for errors
			// and golden files we need to compare against
			if isGoldenFile(path) || !hasGoldenCounterpart(path) {
				gd.lines[normalizedPath] = lines
			}

			// Process markdown files for golden errors
			if strings.HasSuffix(strings.ToLower(path), ".md") && !isGoldenFile(path) {
				if err := extractGoldenErrors(relPath, lines, gd); err != nil {
					return fmt.Errorf("extracting golden errors from %s: %v", relPath, err)
				}
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("walking directory %s: %v", reqFolderPath, err)
		}
	}

	return gd, nil
}

// getNormalizedPath removes the "_" suffix, if exists, and converts the path to slash format
func getNormalizedPath(path string) string {
	// If the path ends with an underscore, remove it
	path = strings.TrimSuffix(path, "_")

	// Convert to slash format for consistency across platforms
	return filepath.ToSlash(path)
}

// isGoldenFile checks if a file is a golden file (ends with "_" before extension)
// Based on design.md: "GoldenFile is a file whose name ends with '_', e.q. `req.md_`"
func isGoldenFile(path string) bool {
	return strings.HasSuffix(path, "_")
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

// extractGoldenEmbedding processes the embedded golden data in markdown files
func extractGoldenEmbedding(lines []string) []string {
	// Define regex patterns for different directive types
	const (
		lineReplacePrefix = `^\s*//\s*line:\s*`
		lineRemovePrefix  = `^\s*//\s*line-\s*`
		lineAddPrefix     = `^\s*//\s*line\+:\s*`
		lineAtBeginPrefix = `^\s*//\s*line1:\s*`
		lineAtEndPrefix   = `^\s*//\s*line>>:\s*`
	)

	// Compile regex patterns once
	lineReplaceRegex := regexp.MustCompile(lineReplacePrefix)
	lineRemoveRegex := regexp.MustCompile(lineRemovePrefix)
	lineAddRegex := regexp.MustCompile(lineAddPrefix)
	lineBeginRegex := regexp.MustCompile(lineAtBeginPrefix)
	lineEndRegex := regexp.MustCompile(lineAtEndPrefix)

	var transformedLines []string
	var sourceLines []string
	var beginningLines []string
	var endLines []string

	// Clone original lines as source
	sourceLines = append(sourceLines, lines...)

	// First pass: collect transformed lines and handle directives
	for _, line := range lines {
		// Skip processing if the line is a GoldenAnnotation
		isGoldenAnnotation := strings.HasPrefix(strings.TrimSpace(line), "//")

		if isGoldenAnnotation {
			// Process directives
			switch {
			case lineReplaceRegex.MatchString(line):
				// Replace the last non-annotation line
				for j := len(transformedLines) - 1; j >= 0; j-- {
					if !strings.HasPrefix(strings.TrimSpace(transformedLines[j]), "//") {
						transformedLines[j] = lineReplaceRegex.ReplaceAllString(line, "")
						break
					}
				}
			case lineRemoveRegex.MatchString(line):
				// Remove the last non-annotation line
				for j := len(transformedLines) - 1; j >= 0; j-- {
					if !strings.HasPrefix(strings.TrimSpace(transformedLines[j]), "//") {
						transformedLines = append(transformedLines[:j], transformedLines[j+1:]...)
						break
					}
				}
			case lineAddRegex.MatchString(line):
				// Add a line after the last non-annotation line
				contentToAdd := lineAddRegex.ReplaceAllString(line, "")
				for j := len(transformedLines) - 1; j >= 0; j-- {
					if !strings.HasPrefix(strings.TrimSpace(transformedLines[j]), "//") {
						transformedLines = append(transformedLines[:j+1], append([]string{contentToAdd}, transformedLines[j+1:]...)...)
						break
					}
				}
			case lineBeginRegex.MatchString(line):
				// Collect lines to be added at the beginning
				beginningLines = append(beginningLines, lineBeginRegex.ReplaceAllString(line, ""))
			case lineEndRegex.MatchString(line):
				// Collect lines to be added at the end
				endLines = append(endLines, lineEndRegex.ReplaceAllString(line, ""))
			default:
				// Other annotations should be preserved
				transformedLines = append(transformedLines, line)
			}
		} else {
			// Normal line, just add it
			transformedLines = append(transformedLines, line)
		}
	}

	// Apply beginning lines (prepend)
	if len(beginningLines) > 0 {
		transformedLines = append(beginningLines, transformedLines...)
	}

	// Apply end lines (append)
	if len(endLines) > 0 {
		transformedLines = append(transformedLines, endLines...)
	}

	return transformedLines
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

// hasGoldenCounterpart checks if a file has a corresponding golden file
// Based on design.md: "GoldenFile is a file whose path  ends with '_', e.q. `req.md_`"
func hasGoldenCounterpart(path string) bool {
	// Check if the golden version of this file exists
	// A golden file has the same path but with an underscore appended
	goldenPath := path + "_"
	_, err := os.Stat(goldenPath)
	return err == nil
}
