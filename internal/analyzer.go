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
	seenReqs := make(map[string]reqLocation)

	// Track coverage data
	type coverageInfo struct {
		covererCount int      // number of coverers for this requirement
		coverers     []string // coverer labels to detect removals
	}
	reqCoverage := make(map[string]coverageInfo) // requirementID -> coverage info

	// Track file URLs and hashes
	fileURLHashes := make(map[string]string) // fileURL -> hash from reqmdjson

	// First pass: Process source files to gather coverage information
	for _, file := range files {
		if file.Type == FileTypeSource {
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
	}

	// Second pass: Process markdown files
	for _, file := range files {
		if file.Type != FileTypeMarkdown {
			continue
		}

		// Check if file has requirements but no PackageID
		if len(file.Requirements) > 0 && file.PackageID == "" {
			errors = append(errors, NewErrMissingPackageIDWithReqs(file.Path, file.Requirements[0].Line))
			continue
		}

		// First gather coverage information from footnotes
		for _, footnote := range file.CoverageFootnotes {
			coverage := reqCoverage[footnote.RequirementID]
			coverage.covererCount = len(footnote.Coverers)
			coverage.coverers = make([]string, 0, coverage.covererCount)
			for _, coverer := range footnote.Coverers {
				coverage.coverers = append(coverage.coverers, coverer.CoverageLabel)
				fileURLHashes[coverer.CoverageURL] = coverer.FileHash
			}
			reqCoverage[footnote.RequirementID] = coverage
		}

		// Process requirements and check for status changes
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

			// Handle coverage status changes
			coverage := reqCoverage[reqID]
			if len(req.CoverageStatusWord) > 0 {
				switch {
				case coverage.covererCount > 0 && req.CoverageStatusWord == "uncvrd":
					// Has coverers but marked as uncovered - change to covered
					action := Action{
						Type:          ActionUpdateStatus,
						FileStruct:    &file,
						Line:          req.Line,
						RequirementID: reqID,
						Data:          "covered",
					}
					actions = append(actions, action)
					if IsVerbose {
						Verbose("ActionUpdateStatus: " + action.String())
					}
				case coverage.covererCount == 0 && req.CoverageStatusWord == "covered":
					// No coverers but marked as covered - change to uncovered
					action := Action{
						Type:          ActionUpdateStatus,
						FileStruct:    &file,
						Line:          req.Line,
						RequirementID: reqID,
						Data:          "uncvrd",
					}
					actions = append(actions, action)
					if IsVerbose {
						Verbose("ActionUpdateStatus: " + action.String())
					}
				}
			}
		}
	}

	return actions, errors
}
