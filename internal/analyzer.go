// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"path/filepath"
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
		MdActions:  make(map[FilePath][]MdAction),
		Reqmdjsons: make(map[FilePath]*Reqmdjson),
	}

	// Build RequirementCoverages from all FileStructures
	if err := a.buildRequirementCoverages(files, &result.ProcessingErrors); err != nil {
		return result, err
	}

	a.buildIdsSortedByPos()

	a.analyzeMdActions(result)
	a.analyzeReqmdjsons(result)

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
		sortCoverersByFileHash(coverage.CurrentCoverers)
		sortCoverersByFileHash(coverage.NewCoverers)

		// coverageStatus is "covered" if there are new coverers
		coverageStatus := CoverageStatusWordUncvrd
		if len(coverage.NewCoverers) > 0 {
			coverageStatus = CoverageStatusWordCovered
		}

		var footnoteId CoverageFootnoteId
		if !coverage.Site.HasAnnotationRef {
			footnoteId = a.nextFootnoteId(coverage.FileStructure.Path)
		} else {
			footnoteId = coverage.Site.CoverageFootnoteID
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
		if !areCoverersEqualByHashes(coverage.CurrentCoverers, coverage.NewCoverers) || coverage.CurrentCoverers == nil {

			a.changedFootnotes[requirementId] = true

			// Create footnote action
			newCf := &CoverageFootnote{
				PackageID:          coverage.FileStructure.PackageID,
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
				if cf.CoverageFootnoteId == coverage.Site.CoverageFootnoteID {
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

/*
Principles:

- If a folder has any requirement with changed footnotes, the whole folder's reqmd.json needs updating
- FileUrl() helper function is used to strip line numbers from CoverageURLs
*/
func (a *analyzer) analyzeReqmdjsons(result *AnalyzerResult) {
	// Map to track json files per folder
	allJsons := make(map[FolderPath]*Reqmdjson) // folder -> Reqmdjson
	changedJsons := make(map[FolderPath]bool)   // folder -> isChanged

	// Process coverages in two passes:
	// 1. Non-changed footnotes
	// 2. Changed footnotes

	// First pass: Process coverages with non-changed footnotes
	for requirementID, coverage := range a.coverages {
		if !a.changedFootnotes[requirementID] {
			folder := filepath.Dir(coverage.FileStructure.Path)

			// Initialize Reqmdjson for this folder if not exists
			if _, exists := allJsons[folder]; !exists {
				allJsons[folder] = &Reqmdjson{
					FileUrl2FileHash: make(map[string]string),
				}
			}

			// Add FileURLs and hashes from current coverers
			for _, c := range coverage.CurrentCoverers {
				fileURL := FileUrl(c.CoverageUrL)
				allJsons[folder].FileUrl2FileHash[fileURL] = c.FileHash
			}
		}
	}

	// Second pass: Process coverages with changed footnotes
	for requirementID, coverage := range a.coverages {
		if a.changedFootnotes[requirementID] {
			folder := filepath.Dir(coverage.FileStructure.Path)

			// Initialize Reqmdjson for this folder if not exists
			if _, exists := allJsons[folder]; !exists {
				allJsons[folder] = &Reqmdjson{
					FileUrl2FileHash: make(map[string]string),
				}
			}

			// Mark folder as changed
			changedJsons[folder] = true

			// Add FileURLs and hashes from new coverers
			for _, c := range coverage.NewCoverers {
				fileURL := FileUrl(c.CoverageUrL)
				allJsons[folder].FileUrl2FileHash[fileURL] = c.FileHash
			}
		}
	}

	// Add changed jsons to result
	for folder, json := range allJsons {
		if changedJsons[folder] {
			folder := filepath.ToSlash(folder)
			result.Reqmdjsons[folder] = json
		}
	}
}

func (a *analyzer) buildRequirementCoverages(files []FileStructure, errors *[]ProcessingError) error {
	// First collect all requirements and their locations
	for _, file := range files {
		if file.Type == FileTypeMarkdown {
			if len(file.Requirements) > 0 && file.PackageID == "" {
				*errors = append(*errors, NewErrMissingPackageIDWithReqs(file.Path, file.Requirements[0].Line))
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
					if req.CoverageFootnoteID != "" {
						updateMaxFootnoteId(req.CoverageFootnoteID)
					}
				}

				// Check CoverageFootnotes
				for _, cf := range file.CoverageFootnotes {
					updateMaxFootnoteId(cf.CoverageFootnoteId)
				}
			}

			for _, req := range file.Requirements {
				var reqID RequirementId = RequirementId(file.PackageID + "/" + string(req.RequirementName)) // FIXME make NewRequirementId

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
					if footnote.CoverageFootnoteId == req.CoverageFootnoteID {
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
		if file.Type == FileTypeSource {
			for _, tag := range file.CoverageTags {
				if coverage, exists := a.coverages[tag.RequirementId]; exists {
					coverer := &Coverer{
						CoverageLabel: file.RelativePath + ":" + fmt.Sprint(tag.Line) + ":" + tag.CoverageType,
						CoverageUrL:   file.FileURL() + "#L" + strconv.Itoa(tag.Line),
						FileHash:      file.FileHash,
					}
					coverage.NewCoverers = append(coverage.NewCoverers, coverer)
				}
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

func sortCoverersByFileHash(coverers []*Coverer) {
	sort.Slice(coverers, func(i, j int) bool {
		return coverers[i].FileHash < coverers[j].FileHash
	})
}

func areCoverersEqualByHashes(a []*Coverer, b []*Coverer) bool {
	comparator := func(c1, c2 *Coverer) int {
		switch {
		case c1.FileHash < c2.FileHash:
			return -1
		case c1.FileHash > c2.FileHash:
			return 1
		default:
			return 0
		}
	}
	return 0 == slices.CompareFunc(a, b, comparator)
}
