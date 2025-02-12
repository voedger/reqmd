package internal

// ITracer defines the high-level interface for tracing workflow.
// It orchestrates scanning, analyzing, and applying changes.
type ITracer interface {
	// Scan: parse files under given paths to build a list of FileStructures and any syntax errors.
	Scan(paths []string) ([]FileStructure, []SyntaxError)
	// Analyze: perform semantic checks and prepare a set of Actions.
	Analyze(files []FileStructure) ([]Action, []SemanticError)
	// Apply: execute the prepared Actions (e.g., modify files, update footnotes).
	Apply(actions []Action) error
}

// IScanner is responsible for scanning file paths and parsing them into FileStructures.
type IScanner interface {
	Scan(reqPath string, srcPaths []string) ([]FileStructure, []SyntaxError)
}

// IAnalyzer checks for semantic issues (e.g., unique RequirementIDs) and generates Actions.
type IAnalyzer interface {
	Analyze(files []FileStructure) ([]Action, []SemanticError)
}

// IApplier applies the Actions (file updates, footnote generation, etc.).
type IApplier interface {
	Apply(actions []Action) error
}

type IGit interface {
	PathToRoot() string
	CommitHash() string
	FileHash(filePath string) (string, error)
}

// Optional specialized parsers, if you want to keep them separate:
// IMarkdownParser could parse only Markdown files.
// ISourceCoverageParser could parse only source files.
//
// type IMarkdownParser interface {
// 	ParseMarkdown(path string) (FileStructure, []SyntaxError)
// }
//
// type ISourceCoverageParser interface {
// 	ParseSourceCoverage(path string) (FileStructure, []SyntaxError)
// }
