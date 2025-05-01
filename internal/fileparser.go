// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// parseFile processes a file as both markdown and source file
// It combines the logic of ParseMarkdownFile and ParseSourceFile into a single pass
func parseFile(mctx *MarkdownContext, filePath string) (*FileStructure, []ProcessingError, error) {
	if IsVerbose {
		Verbose("parseFile", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("ParseFile: failed to open file: %w", err)
	}
	defer file.Close()

	var errors []ProcessingError

	// Determine file type based on extension
	ext := strings.ToLower(filepath.Ext(filePath))
	fileType := FileTypeSource
	if ext == ".md" {
		fileType = FileTypeMarkdown
	}

	structure := &FileStructure{
		Path: filePath,
		Type: fileType,
	}

	// Parse file contents
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inHeader := false
	inCodeBlock := false
	var lastFenceLine int

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check if the line should be ignored based on ignore patterns
		if shouldIgnoreLine(mctx, line) {
			if IsVerbose {
				Verbose("parseFile: ignoring line", "line", lineNum, "file", filePath)
			}
			continue
		}

		// Source file parsing - always parse coverage tags regardless of file type
		if !inCodeBlock {
			tags := parseCoverageTags(filePath, line, lineNum)
			if len(tags) > 0 {
				structure.CoverageTags = append(structure.CoverageTags, tags...)
			}
		}

		// Markdown specific parsing - only if file is markdown
		if fileType == FileTypeMarkdown {
			// Check for code block markers
			if isCodeBlockMarker(line) {
				if !inCodeBlock {
					lastFenceLine = lineNum
					inCodeBlock = true
				} else {
					inCodeBlock = false
				}
				continue
			}

			// Handle header section
			if line == "---" {
				if lineNum == 1 {
					inHeader = true
					continue
				} else {
					inHeader = false
					continue
				}
			}

			if inHeader {
				if matches := headerRegex.FindStringSubmatch(line); len(matches) > 1 {
					pkgId := strings.TrimSpace(matches[1])
					if !identifierRegex.MatchString(pkgId) {
						errors = append(errors, NewErrPkgIdent(filePath, lineNum, pkgId))
					}
					structure.PackageId = PackageId(pkgId)

					// Ignore files with package "ignoreme"
					if strings.HasPrefix(pkgId, "ignoreme") {
						return structure, nil, nil
					}
				}
				continue
			}

			// Only parse requirements and footnotes when not in a code block
			if !inCodeBlock {
				// Parse requirements
				requirements := parseRequirements(filePath, line, lineNum, &errors)
				structure.Requirements = append(structure.Requirements, requirements...)

				// Parse coverage footnotes
				footnote := ParseCoverageFootnote(mctx, filePath, line, lineNum, &errors)
				if footnote != nil {
					structure.CoverageFootnotes = append(structure.CoverageFootnotes, *footnote)
				}
			}
		}
	}

	// Markdown checks for unmatched fence at end of file
	if fileType == FileTypeMarkdown && inCodeBlock {
		errors = append(errors, NewErrUnmatchedFence(filePath, lastFenceLine))
	}

	if err := scanner.Err(); err != nil {
		errors = append(errors, ProcessingError{
			FilePath: filePath,
			Message:  "Error reading file: " + err.Error(),
		})
	}

	return structure, errors, nil
}

// parseCoverageTags finds and returns all coverage tags in a given line.
func parseCoverageTags(filePath string, line string, lineNum int) []CoverageTag {
	var tags []CoverageTag
	matches := coverageTagRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) == 4 {
			tag := CoverageTag{
				RequirementId: StrToReqId(match[1] + "/" + match[2]),
				CoverageType:  match[3],
				Line:          lineNum,
			}
			tags = append(tags, tag)
			if IsVerbose {
				Verbose("parseCoverageTags: CoverageTag", "tag", match[0], "line", lineNum, "file", filePath)
			}
		}
	}
	return tags
}

// shouldIgnoreLine checks if a line should be ignored based on the ignore patterns
// in the MarkdownContext. Returns true if the line should be ignored.
func shouldIgnoreLine(mctx *MarkdownContext, line string) bool {
	// If no ignore patterns are defined, process all lines
	if mctx == nil || len(mctx.IgnorePatterns) == 0 {
		return false
	}

	// Check if the line matches any of the ignore patterns
	for _, pattern := range mctx.IgnorePatterns {
		if pattern.MatchString(line) {
			return true
		}
	}

	return false
}
