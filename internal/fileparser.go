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

// ParseFile processes a file as both markdown and source file
// It combines the logic of ParseMarkdownFile and ParseSourceFile into a single pass
func ParseFile(mctx *MarkdownContext, filePath string) (*FileStructure, []ProcessingError, error) {
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
					pkgID := strings.TrimSpace(matches[1])
					if !identifierRegex.MatchString(pkgID) {
						errors = append(errors, NewErrPkgIdent(filePath, lineNum, pkgID))
					}
					structure.PackageID = pkgID

					// Ignore files with package "ignoreme"
					if strings.HasPrefix(pkgID, "ignoreme") {
						return structure, nil, nil
					}
				}
				continue
			}

			// Only parse requirements and footnotes when not in a code block
			if !inCodeBlock {
				// Parse requirements
				requirements := ParseRequirements(filePath, line, lineNum, &errors)
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
func parseCoverageTags(_ string, line string, lineNum int) []CoverageTag {
	var tags []CoverageTag
	matches := coverageTagRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) == 4 {
			tag := CoverageTag{
				RequirementId: RequirementId(match[1] + "/" + match[2]),
				CoverageType:  match[3],
				Line:          lineNum,
			}
			tags = append(tags, tag)
		}
	}
	return tags
}
