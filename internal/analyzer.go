package internal

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

type analyzer struct {
	coverages map[RequirementID]*RequirementCoverage // RequirementID -> RequirementCoverage
}

func NewAnalyzer() IAnalyzer {
	return &analyzer{
		coverages: make(map[RequirementID]*RequirementCoverage),
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

	a.buildMd(result)

	return result, nil
}

func (a *analyzer) buildMd(result *AnalyzerResult) {
	// Process each coverage to generate actions
	for _, coverage := range a.coverages {
		// Sort both lists by FileURL for comparison
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
				Data:            formatCoverageFootnote(newCf),
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
						CoverageURL:   file.FileURL(),
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

// Helper function to format a coverage footnote
func formatCoverageFootnote(cf *CoverageFootnote) string {
	var refs []string
	for _, coverer := range cf.Coverers {
		refs = append(refs, fmt.Sprintf("[%s](%s)", coverer.CoverageLabel, coverer.CoverageURL))
	}
	return fmt.Sprintf("[^~%s~]: %s", cf.RequirementName, strings.Join(refs, ", "))
}
