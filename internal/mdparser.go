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
	RequirementSiteRegex = regexp.MustCompile(
		"`~([^~]+)~`" + // RequirementSiteLabel = "`" "~" RequirementName "~" "`"
			"(?:" + // Optional group for coverage status and footnote
			"\\s*([a-zA-Z]+)" + // Optional CoverageStatusWord
			"\\s*\\[\\^~([^~]+)~\\]" + // CoverageFootnoteReference
			"\\s*(✅|❓)?" + // Optional CoverageStatusEmoji
			")?")
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
		requirements := ParseRequirements(filePath, line, lineNum, &errors)
		structure.Requirements = append(structure.Requirements, requirements...)

		// Parse coverage footnotes
		footnote := ParseCoverageFootnote(filePath, line, lineNum, &errors)
		if footnote != nil {
			structure.CoverageFootnotes = append(structure.CoverageFootnotes, *footnote)
		}
	}

	if err := scanner.Err(); err != nil {
		errors = append(errors, SyntaxError{
			FilePath: filePath,
			Message:  "Error reading file: " + err.Error(),
		})
	}

	return structure, errors, nil
}

func ParseRequirements(filePath string, line string, lineNum int, errors *[]SyntaxError) []RequirementSite {
	var requirements []RequirementSite

	matches := RequirementSiteRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		// match[1] = RequirementName from RequirementSiteLabel
		// match[2] = CoverageStatusWord (optional)
		// match[3] = ReferenceName from CoverageFootnoteReference
		// match[4] = CoverageStatusEmoji (optional)

		req := RequirementSite{
			FilePath:            filePath,
			RequirementName:     match[1],
			CoverageStatusWord:  match[2],
			ReferenceName:       match[3],
			CoverageStatusEmoji: match[4],
			Line:                lineNum,
			IsAnnotated:         match[3] != "",
		}

		if req.IsAnnotated && (req.RequirementName != req.ReferenceName) {
			*errors = append(*errors, NewErrRequirementSiteIDEqual(filePath, req.Line, req.RequirementName, req.ReferenceName))
		}
		requirements = append(requirements, req)
	}

	return requirements
}

var (
	// "[^~REQ002~]: `[~com.example.basic/REQ002~impl]`[folder1/filename1:line1:impl](https://example.com/pkg1/filename1), [folder2/filename2:line2:test](https://example.com/pkg2/filename2)"
	CoverageFootnoteRegex = regexp.MustCompile(`^\s*\[\^~([^~]+)~\]:\s*` + //Footnote reference
		"`\\[~([^~/]+)/([^~]+)~([^\\]]+)\\]`" + // Hint with package and coverage type
		`(?:\s*(.+))?$`) // Optional coverer list
	CovererRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

func ParseCoverageFootnote(filePath string, line string, lineNum int, _ *[]SyntaxError) (footnote *CoverageFootnote) {

	matches := CoverageFootnoteRegex.FindStringSubmatch(line)
	if len(matches) > 0 {
		footnote = &CoverageFootnote{
			FilePath:      filePath,
			RequirementID: matches[1],
			PackageID:     matches[2],
			Line:          lineNum,
		}

		// Parse coverers if present
		if len(matches) > 5 && matches[5] != "" {
			covererMatches := CovererRegex.FindAllStringSubmatch(matches[5], -1)
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
	}
	return footnote
}
