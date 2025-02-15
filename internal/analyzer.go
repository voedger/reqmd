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

	for _, file := range files {
		if file.Type != FileTypeMarkdown {
			continue
		}

		// Check if file has requirements but no PackageID
		if len(file.Requirements) > 0 && file.PackageID == "" {
			errors = append(errors, NewErrMissingPackageIDWithReqs(file.Path, file.Requirements[0].Line))
			continue
		}

		// Check for duplicate requirements
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

			// For bare requirements without coverage annotation
			if !req.IsAnnotated {
				action := Action{
					Type:       ActionAnnotate,
					FileStruct: &file,
					Line:       req.Line,
					Data:       reqID,
				}
				actions = append(actions, action)
				if IsVerbose {
					Verbose("ActionAnnotate: " + action.String())
				}
			}
		}
	}

	return actions, errors
}
