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

func (a *analyzer) Analyze(files []FileStructure) ([]Action, []ProcessingError) {
	var errors []ProcessingError

	// First pass: Build RequirementCoverages from all FileStructures
	if err := a.buildRequirementCoverages(files, &errors); err != nil {
		return nil, errors
	}

	// Second pass: Generate actions based on coverage analysis
	return a.generateActions(), errors
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

// Helper function to sort coverers by FileURL
func sortCoverersByFileURL(coverers []*Coverer) {
	sort.Slice(coverers, func(i, j int) bool {
		return coverers[i].CoverageURL < coverers[j].CoverageURL
	})
}

func (a *analyzer) generateActions() []Action {
	var actions []Action

	// Process each requirement coverage
	for reqID, coverage := range a.coverages {
		// Sort both current and new coverers for comparison
		sortCoverersByFileURL(coverage.CurrentCoverers)
		sortCoverersByFileURL(coverage.NewCoverers)

		// Check if coverage has changed
		if !coverersEqual(coverage.CurrentCoverers, coverage.NewCoverers) {
			// Convert []*Coverer to []Coverer for the footnote
			coverers := make([]Coverer, len(coverage.NewCoverers))
			for i, c := range coverage.NewCoverers {
				coverers[i] = *c
			}

			// Construct new footnote
			newCf := CoverageFootnote{
				RequirementID: reqID,
				PackageID:     coverage.FileStructure.PackageID,
				Coverers:      coverers,
			}

			// Create footnote action
			coverage.ActionFootnote = &Action{
				Type:          ActionFootnote,
				FileStruct:    coverage.FileStructure,
				Line:          coverage.Site.Line,
				RequirementID: reqID,
				Data:          formatCoverageFootnote(&newCf),
			}
			actions = append(actions, *coverage.ActionFootnote)

			// Update coverage status based on number of new coverers
			coverageStatus := CoverageStatusWordUncvrd
			if len(coverage.NewCoverers) > 0 {
				coverageStatus = CoverageStatusWordCovered
			}

			if coverage.Site.CoverageStatusWord != coverageStatus {
				coverage.ActionUpdateStatus = &Action{
					Type:          ActionUpdateStatus,
					FileStruct:    coverage.FileStructure,
					Line:          coverage.Site.Line,
					RequirementID: reqID,
					Data:          string(coverageStatus),
				}
				actions = append(actions, *coverage.ActionUpdateStatus)
			}
		}

		// Handle bare requirement sites
		if coverage.ActionFootnote == nil && !coverage.Site.IsAnnotated {
			actions = append(actions, Action{
				Type:          ActionAnnotate,
				FileStruct:    coverage.FileStructure,
				Line:          coverage.Site.Line,
				RequirementID: reqID,
				Data:          reqID,
			})
		}

		// Add actions for reqmd.json updates
		if len(coverage.NewCoverers) > 0 {
			for _, coverer := range coverage.NewCoverers {
				fileURL := coverage.FileStructure.FileURL()
				// For new files, add FileURL action
				if fileURL == "" {
					actions = append(actions, Action{
						Type:       ActionAddFileURL,
						FileStruct: coverage.FileStructure,
						Data:       coverer.FileHash,
					})
				} else if coverer.FileHash != coverage.FileStructure.FileHash {
					// For existing files with changed hash
					actions = append(actions, Action{
						Type:       ActionUpdateHash,
						FileStruct: coverage.FileStructure,
						Data:       coverer.FileHash,
					})
				}
			}
		}
	}

	return actions
}

// Helper function to check if two coverer slices are equal
func coverersEqual(a []*Coverer, b []*Coverer) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].CoverageURL != b[i].CoverageURL ||
			a[i].CoverageLabel != b[i].CoverageLabel ||
			a[i].FileHash != b[i].FileHash {
			return false
		}
	}
	return true
}

// Helper function to format a coverage footnote as a string
func formatCoverageFootnote(cf *CoverageFootnote) string {
	var refs []string
	for _, coverer := range cf.Coverers {
		refs = append(refs, fmt.Sprintf("[%s](%s)", coverer.CoverageLabel, coverer.CoverageURL))
	}
	return fmt.Sprintf("[^~%s~]: %s", cf.RequirementID, strings.Join(refs, ", "))
}
