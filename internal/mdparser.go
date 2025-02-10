package internal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Regular expressions for parsing markdown elements
var (
	headerRegex          = regexp.MustCompile(`^reqmd\.package:\s*(.+)$`)
	RequirementSiteRegex = regexp.MustCompile("`~([^~]+)~`(?:cov\\[\\^~([^~]+)~\\])?")
)

func ParseMarkdownFile(filePath string) (*FileStructure, []SyntaxError, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("ParseMarkdownFile: failed to open file: %w", err)
	}
	defer file.Close()

	var errors []SyntaxError
	structure := &FileStructure{
		Path: filePath,
		Type: FileTypeMarkdown,
	}

	// Parse header and content
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inHeader := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Handle header section
		if line == "---" {
			if !inHeader {
				inHeader = true
				continue
			} else {
				inHeader = false
				continue
			}
		}

		if inHeader {
			if matches := headerRegex.FindStringSubmatch(line); len(matches) > 1 {
				structure.PackageID = strings.TrimSpace(matches[1])
			}
			continue
		}

		// Parse requirements
		requirements := parseRequirements(line, lineNum, &errors)
		structure.Requirements = append(structure.Requirements, requirements...)
	}

	if err := scanner.Err(); err != nil {
		errors = append(errors, SyntaxError{
			FilePath: filePath,
			Message:  "Error reading file: " + err.Error(),
		})
	}

	return structure, errors, nil
}

func parseRequirements(line string, lineNum int, errors *[]SyntaxError) []RequirementSite {
	var requirements []RequirementSite

	// Find all requirement references
	matches := RequirementSiteRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) > 1 {
			req := RequirementSite{
				RequirementName: match[1],
				ReferenceName:   match[2],
				Line:            lineNum,
				IsAnnotated:     match[1] == match[2],
			}
			requirements = append(requirements, req)
		}
	}

	return requirements
}
