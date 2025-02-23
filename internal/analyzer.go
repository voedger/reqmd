package internal

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
)

type analyzer struct {
	coverages        map[RequirementID]*requirementCoverage // RequirementID -> RequirementCoverage
	changedFootnotes map[RequirementID]bool
	maxFootnoteIds   map[FilePath]CoverageFootnoteId // Track max footnote ID per file
}

type requirementCoverage struct {
	Site            *RequirementSite
	FileStructure   *FileStructure
	CurrentCoverers []*Coverer // Is not nil if there are existing footnotes
	NewCoverers     []*Coverer
}

func NewAnalyzer() IAnalyzer {
	return &analyzer{
		coverages:        make(map[RequirementID]*requirementCoverage),
		changedFootnotes: make(map[RequirementID]bool),
		maxFootnoteIds:   make(map[FilePath]CoverageFootnoteId),
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

	a.analyzeMdActions(result)
	a.analyzeReqmdjsons(result)

	return result, nil
}

func (a *analyzer) analyzeMdActions(result *AnalyzerResult) {
	// Process each coverage to generate actions
	for requirementID, coverage := range a.coverages {
		// Sort both lists by FileHash for comparison
		sortCoverersByFileHash(coverage.CurrentCoverers)
		sortCoverersByFileHash(coverage.NewCoverers)

		// coverageStatus is "covered" if there are new coverers
		coverageStatus := CoverageStatusWordUncvrd
		if len(coverage.NewCoverers) > 0 {
			coverageStatus = CoverageStatusWordCovered
		}

		// Site action is needed if there's no annotation or coverage status changed
		if !coverage.Site.HasAnnotationRef || coverage.Site.CoverageStatusWord != coverageStatus {
			siteAction := MdAction{
				Type:            ActionSite,
				Path:            coverage.FileStructure.Path,
				Line:            coverage.Site.Line,
				RequirementName: coverage.Site.RequirementName,
				Data:            FormatRequirementSite(coverage.Site.RequirementName, coverageStatus),
			}
			result.MdActions[coverage.FileStructure.Path] = append(
				result.MdActions[coverage.FileStructure.Path],
				siteAction,
			)
		}

		// Footnote action is needed if coverers are different or site is not annotated
		if !areCoverersEqualByHashes(coverage.CurrentCoverers, coverage.NewCoverers) || coverage.CurrentCoverers == nil {
			a.changedFootnotes[requirementID] = true

			// Create footnote action
			newCf := &CoverageFootnote{
				PackageID:       coverage.FileStructure.PackageID,
				RequirementName: coverage.Site.RequirementName,
				ID:              footnoteid,
				Coverers:        make([]Coverer, len(coverage.NewCoverers)),
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
				if cf.RequirementName == coverage.Site.RequirementName {
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

// Helper function to get appropriate coverage status based on coverers
func getCoverageStatus(coverage *requirementCoverage) CoverageStatusWord {
	if len(coverage.NewCoverers) > 0 {
		return CoverageStatusWordCovered
	}
	return CoverageStatusWordUncvrd
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
					FileURL2FileHash: make(map[string]string),
				}
			}

			// Add FileURLs and hashes from current coverers
			for _, c := range coverage.CurrentCoverers {
				fileURL := FileUrl(c.CoverageUrL)
				allJsons[folder].FileURL2FileHash[fileURL] = c.FileHash
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
					FileURL2FileHash: make(map[string]string),
				}
			}

			// Mark folder as changed
			changedJsons[folder] = true

			// Add FileURLs and hashes from new coverers
			for _, c := range coverage.NewCoverers {
				fileURL := FileUrl(c.CoverageUrL)
				allJsons[folder].FileURL2FileHash[fileURL] = c.FileHash
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

			// Track max footnote ID in this file
			for _, cf := range file.CoverageFootnotes {
				if cf.ID > a.maxFootnoteIds[file.Path] {
					a.maxFootnoteIds[file.Path] = cf.ID
				}
			}

			for _, req := range file.Requirements {
				reqID := file.PackageID + "/" + req.RequirementName

				// Check for duplicates using coverages map
				if existing, exists := a.coverages[reqID]; exists {
					*errors = append(*errors, NewErrDuplicateRequirementID(
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
					if footnote.RequirementName == req.RequirementName {
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
				if coverage, exists := a.coverages[tag.RequirementID]; exists {
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

// Finds the next available footnote ID for a given file
func (a *analyzer) nextFootnoteId(filePath FilePath) CoverageFootnoteId {
	currentMax, ok := a.maxFootnoteIds[filePath]
	if !ok {
		currentMax = 0
	}
	nextId := currentMax + 1
	a.maxFootnoteIds[filePath] = nextId
	return nextId
}
