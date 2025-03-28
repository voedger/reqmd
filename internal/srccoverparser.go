// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Global regex for parsing source file coverage tags.
// A CoverageTag is expected in the form: [~PackageID/RequirementName~CoverageType]
var coverageTagRegex = regexp.MustCompile(`\[\~([^/]+)/([^~]+)\~([^\]]+)\]`)

type SourceFileContext struct {
	Git IGit
}

func ParseSourceFile(filePath string) (*FileStructure, []ProcessingError, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("ParseSourceFile: failed to open file: %w", err)
	}
	defer file.Close()

	var errors []ProcessingError
	structure := &FileStructure{
		Path: filePath,
		Type: FileTypeSource,
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Parse coverage tags in the source line.
		tags := ParseCoverageTags(filePath, line, lineNum)
		structure.CoverageTags = append(structure.CoverageTags, tags...)
	}

	if err := scanner.Err(); err != nil {
		errors = append(errors, ProcessingError{
			FilePath: filePath,
			Message:  "Error reading file: " + err.Error(),
		})
	}

	return structure, errors, nil
}

// ParseCoverageTags finds and returns all coverage tags in a given line.
func ParseCoverageTags(filePath, line string, lineNum int) []CoverageTag {
	var tags []CoverageTag
	var matches [][]string
	if strings.Contains(line, "[~") { // TODO remove this check
		matches = coverageTagRegex.FindAllStringSubmatch(line, -1)
	}
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
