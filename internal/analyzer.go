package internal

type analyzer struct{}

func NewAnalyzer() IAnalyzer {
	return &analyzer{}
}

func (a *analyzer) Analyze(files []FileStructure) ([]Action, []ProcessingError) {
	var actions []Action
	var errors []ProcessingError

	// Track requirement IDs to check for duplicates
	seenReqs := make(map[string]string) // reqID -> filePath

	for _, file := range files {
		if file.Type != FileTypeMarkdown {
			continue
		}

		// Check for duplicate requirements
		for _, req := range file.Requirements {
			if existingFile, exists := seenReqs[req.RequirementName]; exists {
				errors = append(errors, ProcessingError{
					FilePath: file.Path,
					Message:  "Duplicate requirement ID '" + req.RequirementName + "' also found in " + existingFile,
				})
				continue
			}
			seenReqs[req.RequirementName] = file.Path

			// For bare requirements without coverage annotation
			if !req.IsAnnotated {
				// Generate ActionAnnotate
				action := Action{
					Type:       ActionAnnotate,
					FileStruct: &file,
					Line:       req.Line,
					Data:       req.RequirementName, // Add uncovered status
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
