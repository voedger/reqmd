package internal

import (
	"fmt"
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

	a.buildMd1(result)
	a.buildMd2(result)

	return result, nil
}

func (a *analyzer) buildMd1(result *AnalyzerResult) {
	// Process each coverage to generate actions
	for _, coverage := range a.coverages {
		// Sort both lists by FileURL for comparison
		sortCoverersByFileURL(coverage.CurrentCoverers)
		sortCoverersByFileURL(coverage.NewCoverers)

		if !coverersEqual(coverage.CurrentCoverers, coverage.NewCoverers) {
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

			// Create status update action
			coverageStatus := CoverageStatusWordUncvrd
			if len(coverage.NewCoverers) > 0 {
				coverageStatus = CoverageStatusWordCovered
			}

			statusAction := MdAction{
				Type:            ActionSite,
				Path:            coverage.FileStructure.Path,
				Line:            coverage.Site.Line,
				RequirementName: coverage.Site.RequirementName,
				Data:            formatRequirementSite(coverage.Site.RequirementName, coverageStatus),
			}

			// Add actions to result
			result.MdActions[coverage.FileStructure.Path] = append(
				result.MdActions[coverage.FileStructure.Path],
				footnoteAction,
				statusAction,
			)
		}
	}
}

func (a *analyzer) buildMd2(result *AnalyzerResult) {
	for _, coverage := range a.coverages {
		// Only process if there's no existing footnote action and the site is not annotated
		if !coverage.Site.IsAnnotated {
			// Create annotation action
			annotateAction := MdAction{
				Type:            ActionSite,
				Path:            coverage.FileStructure.Path,
				Line:            coverage.Site.Line,
				RequirementName: coverage.Site.RequirementName,
				Data:            formatRequirementSite(coverage.Site.RequirementName, CoverageStatusWordUncvrd),
			}
			// Add action to result
			result.MdActions[coverage.FileStructure.Path] = append(
				result.MdActions[coverage.FileStructure.Path],
				annotateAction,
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
					if footnote.RequirementName == reqID {
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
func sortCoverersByFileURL(coverers []*Coverer) {
	sort.Slice(coverers, func(i, j int) bool {
		return coverers[i].CoverageURL < coverers[j].CoverageURL
	})
}

// Helper function to check if two coverer lists have matching URLs
func coverersEqual(a []*Coverer, b []*Coverer) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].CoverageURL != b[i].CoverageURL {
			return false
		}
	}
	return true
}

// Helper function to format a coverage footnote
func formatCoverageFootnote(cf *CoverageFootnote) string {
	var refs []string
	for _, coverer := range cf.Coverers {
		refs = append(refs, fmt.Sprintf("[%s](%s)", coverer.CoverageLabel, coverer.CoverageURL))
	}
	return fmt.Sprintf("[^~%s~]: %s", cf.RequirementName, strings.Join(refs, ", "))
}

// Build a string representation of the RequirementSite according to the requirements
// CoverageStatusEmoji is ✅ for "covered", and ❓ for "uncvrd"
func formatRequirementSite(requirementName string, coverageStatusWord CoverageStatusWord) string {
	result := fmt.Sprintf("`~%s~`", requirementName)

	emoji := "✅"
	if coverageStatusWord == CoverageStatusWordUncvrd {
		emoji = "❓"
	}

	return fmt.Sprintf("%s%s[^~%s~]%s", result, coverageStatusWord, requirementName, emoji)
}
