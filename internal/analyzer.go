package internal

type analyzer struct{}

func NewAnalyzer() IAnalyzer {
	return &analyzer{}
}

func (a *analyzer) Analyze(files []FileStructure) ([]Action, []ProcessingError) {
	var actions []Action
	var errors []ProcessingError

	// Track requirement IDs to check for duplicates
	type reqLocation struct {
		filePath string
		line     int
	}
	seenReqs := make(map[string]reqLocation) // reqID -> location

	// Track coverage data
	type coverageInfo struct {
		count    int      // number of coverers for this requirement
		coverers []string // coverer labels to detect removals
	}
	reqCoverage := make(map[string]coverageInfo) // requirementID -> coverage info

	// Track file URLs and hashes
	fileURLHashes := make(map[string]string) // fileURL -> hash from reqmdjson

	for _, file := range files {
		// Skip source files at this stage
		if file.Type != FileTypeMarkdown {
			continue
		}

		// Check if file has requirements but no PackageID
		if len(file.Requirements) > 0 && file.PackageID == "" {
			errors = append(errors, NewErrMissingPackageIDWithReqs(file.Path, file.Requirements[0].Line))
			continue
		}

		// Process the requirements in the file
		for _, req := range file.Requirements {
			reqID := file.PackageID + "/" + req.RequirementName
			if existing, exists := seenReqs[reqID]; exists {
				errors = append(errors, NewErrDuplicateRequirementID(
					existing.filePath, existing.line,
					file.Path, req.Line,
					reqID))
				continue
			}
			seenReqs[reqID] = reqLocation{
				filePath: file.Path,
				line:     req.Line,
			}

			// Track bare requirements for annotation
			if !req.IsAnnotated {
				action := Action{
					Type:          ActionAnnotate,
					FileStruct:    &file,
					Line:          req.Line,
					RequirementID: reqID,
					Data:          reqID,
				}
				actions = append(actions, action)
				if IsVerbose {
					Verbose("ActionAnnotate: " + action.String())
				}
			}

			// Track coverage status changes
			coverers := 0
			if len(req.CoverageStatusWord) > 0 {
				for _, footnote := range file.CoverageFootnotes {
					if footnote.RequirementID == reqID {
						coverers = len(footnote.Coverers)
						break
					}
				}

				// Update coverage status if needed
				needsUpdate := (coverers > 0 && req.CoverageStatusWord == "uncvrd") ||
					(coverers == 0 && req.CoverageStatusWord == "covered")

				if needsUpdate {
					newStatus := "uncvrd"
					if coverers > 0 {
						newStatus = "covered"
					}
					action := Action{
						Type:          ActionUpdateStatus,
						FileStruct:    &file,
						Line:          req.Line,
						RequirementID: reqID,
						Data:          newStatus,
					}
					actions = append(actions, action)
					if IsVerbose {
						Verbose("ActionUpdateStatus: " + action.String())
					}
				}
			}

			// Store coverage info for later comparison
			reqCoverage[reqID] = coverageInfo{
				count:    coverers,
				coverers: make([]string, 0),
			}
		}

		// Process the coverage footnotes
		for _, footnote := range file.CoverageFootnotes {
			if coverage, exists := reqCoverage[footnote.RequirementID]; exists {
				for _, coverer := range footnote.Coverers {
					coverage.coverers = append(coverage.coverers, coverer.CoverageLabel)
					fileURLHashes[coverer.CoverageURL] = coverer.FileHash
				}
				reqCoverage[footnote.RequirementID] = coverage
			}
		}
	}

	// Now process source files for coverage tags
	for _, file := range files {
		if file.Type != FileTypeSource {
			continue
		}

		fileURL := file.FileURL()

		// Check if file URL needs to be added to reqmd.json
		if _, exists := fileURLHashes[fileURL]; !exists {
			action := Action{
				Type:       ActionAddFileURL,
				FileStruct: &file,
				Data:       file.FileHash,
			}
			actions = append(actions, action)
			if IsVerbose {
				Verbose("ActionAddFileURL: " + action.String())
			}
		} else if fileURLHashes[fileURL] != file.FileHash {
			// File hash has changed
			action := Action{
				Type:       ActionUpdateHash,
				FileStruct: &file,
				Data:       file.FileHash,
			}
			actions = append(actions, action)
			if IsVerbose {
				Verbose("ActionUpdateHash: " + action.String())
			}
		}
	}

	return actions, errors
}
