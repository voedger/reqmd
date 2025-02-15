package internal

import (
	"bufio"
	"fmt"
	"net/url"
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
	identifierRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(?:\.[a-zA-Z][a-zA-Z0-9_]*)*$`)
)

type MarkdownContext struct {
	rfiles ReqmdfilesMap
}

func ParseMarkdownFile(mctx *MarkdownContext, filePath string) (*FileStructure, []ProcessingError, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("ParseMarkdownFile: failed to open file: %w", err)
	}
	defer file.Close()

	var errors []ProcessingError
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
		if line == "---" && lineNum == 1 {
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
				pkgID := strings.TrimSpace(matches[1])
				if !identifierRegex.MatchString(pkgID) {
					errors = append(errors, NewErrPkgIdent(filePath, lineNum, pkgID))
				}
				structure.PackageID = pkgID
			}
			continue
		}

		// Parse requirements
		requirements := ParseRequirements(filePath, line, lineNum, &errors)
		structure.Requirements = append(structure.Requirements, requirements...)

		// Parse coverage footnotes
		footnote := ParseCoverageFootnote(mctx, filePath, line, lineNum, &errors)
		if footnote != nil {
			structure.CoverageFootnotes = append(structure.CoverageFootnotes, *footnote)
		}
	}

	if err := scanner.Err(); err != nil {
		errors = append(errors, ProcessingError{
			FilePath: filePath,
			Message:  "Error reading file: " + err.Error(),
		})
	}

	return structure, errors, nil
}

func ParseRequirements(filePath string, line string, lineNum int, errors *[]ProcessingError) []RequirementSite {
	var requirements []RequirementSite

	matches := RequirementSiteRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		reqName := match[1]
		if !identifierRegex.MatchString(reqName) {
			*errors = append(*errors, NewErrReqIdent(filePath, lineNum))
		}

		covStatus := match[2]
		if covStatus != "" && covStatus != "covered" && covStatus != "uncvrd" {
			*errors = append(*errors, NewErrCoverageStatusWord(filePath, lineNum, covStatus))
		}

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

func ParseCoverageFootnote(mctx *MarkdownContext, filePath string, line string, lineNum int, errs *[]ProcessingError) (footnote *CoverageFootnote) {

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
					parsedURL, err := url.Parse(covMatch[2])

					// Add NewErrURLSyntax to errs
					if err != nil {
						*errs = append(*errs, NewErrURLSyntax(filePath, lineNum, covMatch[2]))
						continue
					}

					coverer := Coverer{
						CoverageLabel: covMatch[1],
						CoverageURL:   covMatch[2],
					}
					fileURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
					coverer.FileHash = mctx.rfiles[fileURL]
					footnote.Coverers = append(footnote.Coverers, coverer)
				}
			}
		}
	}
	return footnote
}
