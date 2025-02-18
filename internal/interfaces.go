package internal

// ITracer defines the high-level interface for tracing workflow.
// It orchestrates scanning, analyzing, and applying changes.
type ITracer interface {
	Trace() error
}

// IScanner is responsible for scanning file paths and parsing them into FileStructures.
type IScanner interface {
	Scan(reqPath string, srcPaths []string) (*ScannerResult, error)
}

// IAnalyzer checks for semantic issues (e.g., unique RequirementIDs) and generates Actions.
type IAnalyzer interface {
	Analyze(files []FileStructure) (*AnalyzerResult, error)
}

// IApplier applies the Actions (file updates, footnote generation, etc.).
type IApplier interface {
	Apply(*AnalyzerResult) error
}

type IGit interface {
	PathToRoot() string
	CommitHash() string
	FileHash(filePath string) (string, error)
	RepoRootFolderURL() string
}

// Optional specialized parsers, if you want to keep them separate:
// IMarkdownParser could parse only Markdown files.
// ISourceCoverageParser could parse only source files.
//
// type IMarkdownParser interface {
// 	ParseMarkdown(path string) (FileStructure, []ProcessingError)
// }
//
// type ISourceCoverageParser interface {
// 	ParseSourceCoverage(path string) (FileStructure, []ProcessingError)
// }
