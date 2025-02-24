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
	identifierRegex      = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(?:\.[a-zA-Z][a-zA-Z0-9_]*)*$`)
	codeBlockMarkerRegex = regexp.MustCompile(`^\s*` + "```")
)

type MarkdownContext struct {
	rfiles *Reqmdjson
}

// isCodeBlockMarker checks if a line is a code block marker, handling indentation
func isCodeBlockMarker(line string) bool {
	return codeBlockMarkerRegex.MatchString(line)
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
	inCodeBlock := false
	var lastFenceLine int // Track the line number of the last opening fence

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for code block markers, handling indentation
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

	// Check for unmatched fence at end of file
	if inCodeBlock {
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

func ParseRequirements(filePath string, line string, lineNum int, errors *[]ProcessingError) []RequirementSite {
	var requirements []RequirementSite

	matches := RequirementSiteRegex.FindAllStringSubmatch(line, -1)
	if len(matches) > 1 {
		*errors = append(*errors, NewErrMultiSites(filePath, lineNum, matches[0][0], matches[1][0]))
		return nil
	}

	for _, match := range matches {
		reqName := match[1]
		if !identifierRegex.MatchString(reqName) {
			*errors = append(*errors, NewErrReqIdent(filePath, lineNum))
		}

		covStatus := match[2]
		if covStatus != "" && covStatus != "covered" && covStatus != "uncvrd" {
			*errors = append(*errors, NewErrCoverageStatusWord(filePath, lineNum, covStatus))
			return requirements
		}

		req := RequirementSite{
			FilePath:            filePath,
			RequirementName:     RequirementName(match[1]),
			CoverageStatusWord:  CoverageStatusWord(match[2]),
			CoverageFootnoteID:  CoverageFootnoteId(match[3]),
			CoverageStatusEmoji: CoverageStatusEmoji(match[4]),
			Line:                lineNum,
			HasAnnotationRef:    match[3] != "",
		}

		// TODO syntax error to match CoverageStatusEmoji and CoverageStatusWord

		if req.HasAnnotationRef && covStatus == "" {
			*errors = append(*errors, NewErrCoverageStatusWord(filePath, lineNum, covStatus))
			return requirements
		}

		requirements = append(requirements, req)
	}

	return requirements
}

func ParseCoverageFootnote(mctx *MarkdownContext, filePath string, line string, lineNum int, errs *[]ProcessingError) (footnote *CoverageFootnote) {

	matches := CoverageFootnoteRegex.FindStringSubmatch(line)
	if len(matches) > 0 {
		footnote = &CoverageFootnote{
			FilePath:           filePath,
			CoverageFootnoteID: CoverageFootnoteId(matches[1]),
			PackageID:          matches[2],
			Line:               lineNum,
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
						CoverageUrL:   covMatch[2],
					}
					fileURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
					if mctx != nil && mctx.rfiles != nil {
						coverer.FileHash = mctx.rfiles.FileUrl2FileHash[fileURL]
					}
					footnote.Coverers = append(footnote.Coverers, coverer)
				}
			}
		}
	}
	return footnote
}
