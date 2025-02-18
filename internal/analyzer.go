package internal

import (
	"fmt"
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
	var result AnalyzerResult

	// First pass: Build RequirementCoverages from all FileStructures
	if err := a.buildRequirementCoverages(files, &result.ProcessingErrors); err != nil {
		return &result, err
	}

	// Second pass: Generate actions based on coverage analysis
	return &result, nil
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
					if footnote.RequirementID == reqID {
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

// // Helper function to sort coverers by FileURL
// func sortCoverersByFileURL(coverers []*Coverer) {
// 	sort.Slice(coverers, func(i, j int) bool {
// 		return coverers[i].CoverageURL < coverers[j].CoverageURL
// 	})
// }

// // Helper function to check if two coverer slices are equal
// func coverersEqual(a []*Coverer, b []*Coverer) bool {
// 	if len(a) != len(b) {
// 		return false
// 	}
// 	for i := range a {
// 		if a[i].CoverageURL != b[i].CoverageURL ||
// 			a[i].CoverageLabel != b[i].CoverageLabel ||
// 			a[i].FileHash != b[i].FileHash {
// 			return false
// 		}
// 	}
// 	return true
// }

// // Helper function to format a coverage footnote as a string
// func formatCoverageFootnote(cf *CoverageFootnote) string {
// 	var refs []string
// 	for _, coverer := range cf.Coverers {
// 		refs = append(refs, fmt.Sprintf("[%s](%s)", coverer.CoverageLabel, coverer.CoverageURL))
// 	}
// 	return fmt.Sprintf("[^~%s~]: %s", cf.RequirementID, strings.Join(refs, ", "))
// }
