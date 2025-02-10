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
	headerRegex           = regexp.MustCompile(`^reqmd\.package:\s*(.+)$`)
	RequirementSiteRegex  = regexp.MustCompile("`~([^~]+)~`(?:cov\\[\\^~([^~]+)~\\])?")
	CoverageFootnoteRegex = regexp.MustCompile(`^\[^~([^~]+)~\]:\s*` + // Footnote reference
		"`\\[~([^~]+)~([^\\]]+)\\]`" + // Hint with package and coverage type
		`(?:\s*(.+))?$`) // Optional coverer list
	CovererRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
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
		requirements := parseRequirements(filePath, line, lineNum, &errors)
		structure.Requirements = append(structure.Requirements, requirements...)

		// Parse coverage footnotes
		footnotes := parseCoverageFootnotes(filePath, line, lineNum, &errors)
		structure.CoverageFootnotes = append(structure.CoverageFootnotes, footnotes...)
	}

	if err := scanner.Err(); err != nil {
		errors = append(errors, SyntaxError{
			FilePath: filePath,
			Message:  "Error reading file: " + err.Error(),
		})
	}

	return structure, errors, nil
}

func parseRequirements(filePath string, line string, lineNum int, errors *[]SyntaxError) []RequirementSite {
	var requirements []RequirementSite

	// Find all requirement references
	matches := RequirementSiteRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) > 1 {
			req := RequirementSite{
				FilePath:        filePath,
				RequirementName: match[1],
				ReferenceName:   match[2],
				Line:            lineNum,
				IsAnnotated:     match[1] == match[2],
			}
			if req.IsAnnotated && (req.RequirementName != req.ReferenceName) {
				*errors = append(*errors, NewErrRequirementSiteIDEqual(filePath, req.Line, req.RequirementName, req.ReferenceName))
			}
			requirements = append(requirements, req)
		}
	}

	return requirements
}

func parseCoverageFootnotes(filePath string, line string, lineNum int, _ *[]SyntaxError) []CoverageFootnote {
	var footnotes []CoverageFootnote

	matches := CoverageFootnoteRegex.FindStringSubmatch(line)
	if len(matches) > 0 {
		footnote := CoverageFootnote{
			FilePath:          filePath,
			RequirementSiteID: matches[1],
			PackageID:         matches[2],
			Line:              lineNum,
		}

		// Parse coverers if present
		if len(matches) > 4 && matches[4] != "" {
			covererMatches := CovererRegex.FindAllStringSubmatch(matches[4], -1)
			for _, covMatch := range covererMatches {
				if len(covMatch) > 2 {
					coverer := Coverer{
						CoverageLabel: covMatch[1],
						CoverageURL:   covMatch[2],
					}
					footnote.Coverers = append(footnote.Coverers, coverer)
				}
			}
		}

		footnotes = append(footnotes, footnote)
	}

	return footnotes
}
