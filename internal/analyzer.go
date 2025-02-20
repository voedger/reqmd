package internal

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
)

type analyzer struct {
	coverages        map[RequirementID]*RequirementCoverage // RequirementID -> RequirementCoverage
	changedFootnotes map[RequirementID]bool
}

func NewAnalyzer() IAnalyzer {
	return &analyzer{
		coverages:        make(map[RequirementID]*RequirementCoverage),
		changedFootnotes: make(map[RequirementID]bool),
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

	a.buildMdActions(result)
	a.buildReqmdjsons(result)

	return result, nil
}

func (a *analyzer) buildMdActions(result *AnalyzerResult) {
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

		// Check if site action is needed
		if !coverage.Site.IsAnnotated || coverage.Site.CoverageStatusWord != coverageStatus {
			siteAction := MdAction{
				Type:            ActionSite,
				Path:            coverage.FileStructure.Path,
				Line:            coverage.Site.Line,
				RequirementName: coverage.Site.RequirementName,
				Data:            FormatRequirementSite(coverage.Site.RequirementName, coverageStatus),
			}
			// Add actions to result
			result.MdActions[coverage.FileStructure.Path] = append(
				result.MdActions[coverage.FileStructure.Path],
				siteAction,
			)
		}

		// Footnote action is needed if coverers are different
		if !areCoverersEqualByHashes(coverage.CurrentCoverers, coverage.NewCoverers) {

			a.changedFootnotes[requirementID] = true

			// Create footnote action
			newCf := &CoverageFootnote{
				RequirementName: coverage.Site.RequirementName,
				Coverers:        make([]Coverer, len(coverage.NewCoverers)),
			}
			for i, c := range coverage.NewCoverers {
				newCf.Coverers[i] = *c // Convert *Coverer to Coverer
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

/*
Principles:

- If a folder has any requirement with changed footnotes, the whole folder's reqmd.json needs updating
- FileUrl() helper function is used to strip line numbers from CoverageURLs

Flow:

- allJsons is created, map[requirementFolder]*Reqmdjson
  - requirementFolder is determined as filepath.Dir(coverage.FileStructure.Path)

- changedJsons is created, map[requirementFolder]bool
- First allJsons is populated from coverages with non-changed footnotes
  - All FileURL(coverer.CoverageUrl) and FileHashes from CurrentCoverers

- Then allJsons and changedJsons are populated from coverages with changed footnotes
  - All FileUrL(coverer.CoverageUrl) and FileHashes from NewCoverers
  - changedJsons for a given folder is set to true

- Reqmdjson from allJsons which are mentioned in changedJsons are added to result
*/
func (a *analyzer) buildReqmdjsons(result *AnalyzerResult) {
	// Map to track json files per folder
	allJsons := make(map[string]*Reqmdjson) // folder -> Reqmdjson
	changedJsons := make(map[string]bool)   // folder -> isChanged

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
				coverage := &RequirementCoverage{
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
						CoverageUrL:   file.FileURL() + "#" + strconv.Itoa(tag.Line),
						FileHash:      file.FileHash,
					}
					coverage.NewCoverers = append(coverage.NewCoverers, coverer)
				}
			}
		}
	}

	return nil
}

// Helper function to sort coverers by FileURL
func sortCoverersByFileHash(coverers []*Coverer) {
	sort.Slice(coverers, func(i, j int) bool {
		return coverers[i].FileHash < coverers[j].FileHash
	})
}

// areCoverersEqualByHashes compares two slices of Coverer pointers for equality
// based on their FileHash values. Returns true if both slices contain the same
// FileHash values regardless of order, false otherwise.
//
// Preconditions:
// - Slices must be non-nil
// - All elements must be non-nil
// - Empty slices are considered equal
//
// Time complexity: O(n log n) where n is the length of the longer slice
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
