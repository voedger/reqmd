package internal

import (
	"fmt"
	"sort"
	"strings"
)

type analyzer struct {
	coverages map[string]*RequirementCoverage // RequirementID -> RequirementCoverage
}

func NewAnalyzer() IAnalyzer {
	return &analyzer{
		coverages: make(map[string]*RequirementCoverage),
	}
}

func (a *analyzer) Analyze(files []FileStructure) ([]Action, []ProcessingError) {
	var errors []ProcessingError

	// Track requirement IDs to check for duplicates
	seenReqs := make(map[string]struct {
		filePath string
		line     int
	})

	// First pass: Build RequirementCoverages from all FileStructures
	if err := a.buildRequirementCoverages(files, seenReqs, &errors); err != nil {
		return nil, errors
	}

	// Second pass: Generate actions based on coverage analysis
	return a.generateActions(), errors
}

func (a *analyzer) buildRequirementCoverages(files []FileStructure, seenReqs map[string]struct {
	filePath string
	line     int
}, errors *[]ProcessingError) error {
	// First collect all requirements and their locations
	for _, file := range files {
		if file.Type == FileTypeMarkdown {
			if len(file.Requirements) > 0 && file.PackageID == "" {
				*errors = append(*errors, NewErrMissingPackageIDWithReqs(file.Path, file.Requirements[0].Line))
				continue
			}

			for _, req := range file.Requirements {
				reqID := file.PackageID + "/" + req.RequirementName
				if existing, exists := seenReqs[reqID]; exists {
					*errors = append(*errors, NewErrDuplicateRequirementID(
						existing.filePath, existing.line,
						file.Path, req.Line,
						reqID))
					continue
				}
				seenReqs[reqID] = struct {
					filePath string
					line     int
				}{
					filePath: file.Path,
					line:     req.Line,
				}

				// Initialize or update RequirementCoverage
				coverage, exists := a.coverages[reqID]
				if !exists {
					coverage = &RequirementCoverage{
						Site:          &req,
						FileStructure: &file,
					}
					a.coverages[reqID] = coverage
				}

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
				Type:          ActionAddCoverer,
				FileStruct:    coverage.FileStructure,
				Line:          coverage.Site.Line,
				RequirementID: reqID,
				Data:          formatCoverageFootnote(&newCf),
			}
			actions = append(actions, *coverage.ActionFootnote)

			// Update coverage status
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
