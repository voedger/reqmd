package internal

// ITracer defines the contract for the tracer that orchestrates the entire tracing process.
// It is responsible for coordinating the scan, analysis, and application phases.
type ITracer interface {
	// Trace executes the full tracing workflow given a list of markdown directories and optional repository paths.
	Trace(markdownPaths []string, repoPaths []string) error
}

// IScanner defines the contract for scanning directories to locate and parse input files.
// It handles the discovery and initial parsing of both markdown and source files.
type IScanner interface {
	// Scan traverses the specified directory and returns a slice of parsed MarkdownFile models.
	Scan(directory string) ([]*MarkdownFile, error)
}

// IAnalyzer defines the contract for analyzing parsed file structures to generate actions and detect semantic errors.
// It validates the requirements (e.g., ensuring uniqueness) and prepares update actions.
type IAnalyzer interface {
	// Analyze validates the parsed MarkdownFiles and generates a list of Actions.
	// It returns both the actions to be applied and any semantic errors encountered during analysis.
	Analyze(files []*MarkdownFile) (actions []*Action, semanticErrors []*SemanticError, err error)
}

// IApplicator defines the contract for applying the actions generated during analysis to update files.
// This includes updating markdown files and the reqmdfiles.json mapping.
type IApplicator interface {
	// Apply executes the list of Actions to modify files as required.
	Apply(actions []*Action) error
}

// IFileHasher defines the contract for computing file hashes.
// It abstracts the functionality of invoking external commands (e.g., git hash-object) to compute file hashes.
type IFileHasher interface {
	// Hash computes and returns the hash of a file given its URL.
	Hash(fileURL string) (string, error)
}
