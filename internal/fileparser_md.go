// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"net/url"
	"regexp"
)

// Regular expressions for parsing markdown elements
var (
	headerRegex          = regexp.MustCompile(`^reqmd\.package:\s*(.+)$`)
	identifierRegex      = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(?:\.[a-zA-Z][a-zA-Z0-9_]*)*$`)
	codeBlockMarkerRegex = regexp.MustCompile(`^\s*` + "```")
)

type MarkdownContext struct {
}

// isCodeBlockMarker checks if a line is a code block marker, handling indentation
func isCodeBlockMarker(line string) bool {
	return codeBlockMarkerRegex.MatchString(line)
}

func parseRequirements(filePath string, line string, lineNum int, errors *[]ProcessingError) []RequirementSite {
	var requirements []RequirementSite

	matches := RequirementSiteRegex.FindAllStringSubmatch(line, -1)
	if len(matches) > 1 {
		*errors = append(*errors, NewErrMultiSites(filePath, lineNum, matches[0][0], matches[1][0]))
		return nil
	}

	for _, match := range matches {
		if IsVerbose {
			Verbose("parseRequirements: RequirementSite", "site", match[0], "line", lineNum, "file", filePath)
		}
		reqName := match[1]
		if !identifierRegex.MatchString(reqName) {
			*errors = append(*errors, NewErrReqIdent(filePath, lineNum))
		}

		covStatus := match[2]
		if covStatus != "" &&
			covStatus != string(CoverageStatusWordCovered) &&
			covStatus != string(CoverageStatusWordUncvrd) &&
			covStatus != string(CoverageStatusWordCovrd) {
			*errors = append(*errors, NewErrCoverageStatusWord(filePath, lineNum, covStatus))
			return requirements
		}

		req := RequirementSite{
			RequirementName:     RequirementName(match[1]),
			CoverageStatusWord:  CoverageStatusWord(match[2]),
			CoverageFootnoteId:  CoverageFootnoteId(match[3]),
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
			CoverageFootnoteId: CoverageFootnoteId(matches[1]),
			PackageId:          PackageId(matches[2]),
			Line:               lineNum,
		}

		// Parse coverers if present
		if len(matches) > 5 && matches[5] != "" {
			covererMatches := CovererRegex.FindAllStringSubmatch(matches[5], -1)
			for _, covMatch := range covererMatches {
				if len(covMatch) > 2 {
					_, err := url.Parse(covMatch[2])
					// Add NewErrURLSyntax to errs
					if err != nil {
						*errs = append(*errs, NewErrURLSyntax(filePath, lineNum, covMatch[2], err))
						continue
					}

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
