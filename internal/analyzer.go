// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
)

type analyzer struct {
	coverages map[RequirementId]*requirementCoverage // RequirementId -> RequirementCoverage

	// RequirementIds sorted by position in the file
	// Position is coverages[RequirementId]Site.FilePath + coverages[RequirementId]Site.Line
	idsSortedByPos []RequirementId

	changedFootnotes  map[RequirementId]bool
	maxFootnoteIntIds map[FilePath]int // Track max footnote int id per file
}

type requirementCoverage struct {
	Site            *RequirementSite
	FileStructure   *FileStructure
	CurrentCoverers []*Coverer // Is not nil if there are existing footnotes
	NewCoverers     []*Coverer
}

func NewAnalyzer() IAnalyzer {
	return &analyzer{
		coverages:         make(map[RequirementId]*requirementCoverage),
		changedFootnotes:  make(map[RequirementId]bool),
		maxFootnoteIntIds: make(map[FilePath]int),
	}
}

func (a *analyzer) Analyze(files []FileStructure) (*AnalyzerResult, error) {
	result := &AnalyzerResult{
		MdActions: make(map[FilePath][]MdAction),
	}

	// Build RequirementCoverages from all FileStructures
	if err := a.buildRequirementCoverages(files, &result.ProcessingErrors); err != nil {
		return result, err
	}

	a.buildIdsSortedByPos()

	a.analyzeMdActions(result)

	return result, nil
}

func (a *analyzer) buildIdsSortedByPos() {
	// Create slice with all requirement IDs
	a.idsSortedByPos = make([]RequirementId, 0, len(a.coverages))
	for reqId := range a.coverages {
		a.idsSortedByPos = append(a.idsSortedByPos, reqId)
	}

	// Sort by file path + line number
	sort.Slice(a.idsSortedByPos, func(i, j int) bool {
		reqI := a.coverages[a.idsSortedByPos[i]]
		reqJ := a.coverages[a.idsSortedByPos[j]]

		// Compare file paths first
		if reqI.Site.FilePath != reqJ.Site.FilePath {
			return reqI.Site.FilePath < reqJ.Site.FilePath
		}

		// If same file, compare line numbers
		return reqI.Site.Line < reqJ.Site.Line
	})
}

func (a *analyzer) analyzeMdActions(result *AnalyzerResult) {
	// Process each coverage to generate actions
	for _, requirementId := range a.idsSortedByPos {
		coverage := a.coverages[requirementId]
		// Sort both lists by FileHash for comparison
		sortCoverersByCoverageURL(coverage.CurrentCoverers)
		sortCoverersByCoverageURL(coverage.NewCoverers)

		// coverageStatus is "covered" if there are new coverers
		coverageStatus := CoverageStatusWordUncvrd
		if len(coverage.NewCoverers) > 0 {
			coverageStatus = CoverageStatusWordCovered
		}

		var footnoteId CoverageFootnoteId
		if !coverage.Site.HasAnnotationRef {
			footnoteId = a.nextFootnoteId(coverage.FileStructure.Path)
		} else {
			footnoteId = coverage.Site.CoverageFootnoteId
		}

		// Check if site action is needed
		if !coverage.Site.HasAnnotationRef || coverage.Site.CoverageStatusWord != coverageStatus {
			siteAction := MdAction{
				Type:            ActionSite,
				Path:            coverage.FileStructure.Path,
				Line:            coverage.Site.Line,
				RequirementName: coverage.Site.RequirementName,
				Data:            FormatRequirementSite(coverage.Site.RequirementName, coverageStatus, footnoteId),
			}
			// Add actions to result
			result.MdActions[coverage.FileStructure.Path] = append(
				result.MdActions[coverage.FileStructure.Path],
				siteAction,
			)
		}

		// Footnote action is needed if coverers are different or site is not annotated
		if !areCoverersEqualByURLs(coverage.CurrentCoverers, coverage.NewCoverers) || coverage.CurrentCoverers == nil {

			a.changedFootnotes[requirementId] = true

			// Create footnote action
			newCf := &CoverageFootnote{
				PackageId:          coverage.FileStructure.PackageId,
				CoverageFootnoteId: footnoteId,
				Coverers:           make([]Coverer, len(coverage.NewCoverers)),
				RequirementName:    coverage.Site.RequirementName,
			}
			for i, c := range coverage.NewCoverers {
				newCf.Coverers[i] = *c
			}

			footnoteAction := MdAction{
				Type:            ActionFootnote,
				Path:            coverage.FileStructure.Path,
				RequirementName: coverage.Site.RequirementName,
				Data:            FormatCoverageFootnote(newCf),
			}

			// Find annotation line, keep 0 if not found
			for _, cf := range coverage.FileStructure.CoverageFootnotes {
				if cf.CoverageFootnoteId == coverage.Site.CoverageFootnoteId {
					footnoteAction.Line = cf.Line
					break
				}
			}

			// Add actions to result
			result.MdActions[coverage.FileStructure.Path] = append(
				result.MdActions[coverage.FileStructure.Path],
				footnoteAction,
			)
		}
	}
}

func (a *analyzer) buildRequirementCoverages(files []FileStructure, errors *[]ProcessingError) error {

	// Processes files to analyze and manage requirements and their coverage.
	// - Iterates through the provided files and processes only Markdown files.
	// - Validates that Markdown files with requirements have a valid PackageId.
	//   If not, it logs an error and skips further processing for that file.
	// - Tracks the maximum integer value of footnote IDs in c.maxFootnoteIntIds
	//   for each file by analyzing both RequirementSites and CoverageFootnotes.
	// - Builds `a.coverages`, map[RequirementId]*requirementCoverage:
	//   - For each requirement in the file:
	//     - Generates a unique RequirementId by combining the PackageId and the
	//       requirement name.
	//     - Checks for duplicate RequirementIds in the `coverages` map. If a
	//       duplicate is found, it logs an error and skips the requirement.
	//     - Initializes a `requirementCoverage` object and adds it to the
	//       `coverages` map.
	// - Processes existing coverage footnotes for each requirement and populates
	//   the `CurrentCoverers` field of the `requirementCoverage` object if a
	//   matching CoverageFootnoteId is found.
	// This code is part of a larger system that tracks and validates requirements
	// in Markdown files, ensuring they are uniquely identified and properly
	// annotated with coverage information.

	for _, file := range files {
		if file.Type == FileTypeMarkdown {
			if len(file.Requirements) > 0 && file.PackageId == "" {
				*errors = append(*errors, NewErrMissingPackageIdWithReqs(file.Path, file.Requirements[0].Line))
				continue
			}

			// Track max footnote int Id from both RequirementSites and CoverageFootnotes
			{
				updateMaxFootnoteId := func(id CoverageFootnoteId) {
					intId, err := strconv.Atoi(string(id))
					if err != nil {
						return
					}
					if intId > a.maxFootnoteIntIds[file.Path] {
						a.maxFootnoteIntIds[file.Path] = intId
					}
				}

				// Check RequirementSites
				for _, req := range file.Requirements {
					if req.CoverageFootnoteId != "" {
						updateMaxFootnoteId(req.CoverageFootnoteId)
					}
				}

				// Check CoverageFootnotes
				for _, cf := range file.CoverageFootnotes {
					updateMaxFootnoteId(cf.CoverageFootnoteId)
				}
			}

			for _, req := range file.Requirements {
				var reqID RequirementId = RequirementId(file.PackageId + "/" + string(req.RequirementName)) // FIXME make NewRequirementId

				// Check for duplicates using coverages map
				if existing, exists := a.coverages[reqID]; exists {
					*errors = append(*errors, NewErrDuplicateRequirementId(
						existing.FileStructure.Path, existing.Site.Line,
						file.Path, req.Line,
						reqID))
					continue
				}

				// Initialize RequirementCoverage
				coverage := &requirementCoverage{
					Site:          &req,
					FileStructure: &file,
				}
				a.coverages[reqID] = coverage

				// Process existing coverage footnotes
				for _, footnote := range file.CoverageFootnotes {
					if footnote.CoverageFootnoteId == req.CoverageFootnoteId {
						// Convert []Coverer to []*Coverer
						coverage.CurrentCoverers = make([]*Coverer, len(footnote.Coverers))
						for i := range footnote.Coverers {
							cov := footnote.Coverers[i]
							coverage.CurrentCoverers[i] = &cov
						}
						break
					}
				}
			}
		}
	}

	// Then collect all coverage tags
	for _, file := range files {
		for _, tag := range file.CoverageTags {
			if coverage, exists := a.coverages[tag.RequirementId]; exists {
				coverer := &Coverer{
					CoverageLabel: file.RelativePath + ":" + fmt.Sprint(tag.Line) + ":" + tag.CoverageType,
					CoverageURL:   file.FileURL() + "#L" + strconv.Itoa(tag.Line),
					fileHash:      file.FileHash,
				}
				coverage.NewCoverers = append(coverage.NewCoverers, coverer)
			}
		}
	}

	return nil
}

// Finds the next available footnote ID for a given file
// nolint
func (a *analyzer) nextFootnoteId(filePath FilePath) CoverageFootnoteId {
	currentMax, ok := a.maxFootnoteIntIds[filePath]
	if !ok {
		currentMax = 0
	}
	nextId := currentMax + 1
	a.maxFootnoteIntIds[filePath] = nextId
	return CoverageFootnoteId(strconv.Itoa(nextId))
}

func sortCoverersByCoverageURL(coverers []*Coverer) {
	sort.Slice(coverers, func(i, j int) bool {
		return coverers[i].CoverageURL < coverers[j].CoverageURL
	})
}

func areCoverersEqualByURLs(a []*Coverer, b []*Coverer) bool {
	comparator := func(c1, c2 *Coverer) int {
		switch {
		case c1.CoverageURL < c2.CoverageURL:
			return -1
		case c1.CoverageURL > c2.CoverageURL:
			return 1
		default:
			return 0
		}
	}
	return 0 == slices.CompareFunc(a, b, comparator)
}
