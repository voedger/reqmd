package internal

// ActionType represents the type of transformation needed.
type ActionType string

const (
	ActionAdd    ActionType = "Add"
	ActionUpdate ActionType = "Update"
	ActionDelete ActionType = "Delete"
)

// Action describes a single transformation (add/update/delete) to be applied in a file.
type Action struct {
	Type     ActionType // e.g., Add, Update, Delete
	FilePath string     // which file is changed
	Line     int        // the line number where the change is applied
	Data     string     // new data (if any)
}

// SyntaxError captures syntax-level errors found while parsing a file.
type SyntaxError struct {
	FilePath string // file that has a syntax error
	Line     int    // line number where the syntax error is detected
	Message  string // human-readable description
}

// SemanticError captures higher-level domain errors (e.g., duplicate RequirementIDs).
type SemanticError struct {
	FilePath string // file that triggered the semantic error (optional if more general)
	Message  string // human-readable description
}

// FileType distinguishes between different file categories (Markdown vs source, etc.).
type FileType int

const (
	FileTypeMarkdown FileType = iota
	FileTypeSource
)

// FileStructure merges the parsed data from an input file (Markdown or source).
type FileStructure struct {
	Path         string
	Type         FileType          // indicates if it's Markdown or source
	PackageID    string            // parsed from Markdown header (if markdown)
	Requirements []RequirementSite // for Markdown: discovered requirements (bare or annotated)
	CoverageTags []CoverageTag     // for source: discovered coverage tags
	// ... Add more fields if needed for raw file content, line references, etc.
}

// RequirementSite represents a single requirement reference discovered in a Markdown file.
type RequirementSite struct {
	RequirementName string // e.g., "Post.handler"
	Line            int    // line number where the requirement is defined/referenced
	IsAnnotated     bool   // true if it already has coverage annotation, false if it’s bare
}

// CoverageTag represents a coverage marker found in source code.
type CoverageTag struct {
	RequirementID string // e.g., "server.api.v2/Post.handler"
	CoverageType  string // e.g., "impl", "test"
	Line          int    // line number where the coverage tag was found
}

// CoverageFootnote represents the footnote in Markdown that references coverage tags.
type CoverageFootnote struct {
	RequirementID string // e.g., "server.api.v2/Post.handler"
	Coverers      []Coverer
}

// Coverer represents one coverage reference within a footnote, e.g., [folder/file:line:impl](URL)
type Coverer struct {
	CoverageLabel string // e.g., "folder/file.go:42:impl"
	CoverageURL   string // e.g., full URL including commit hash
}

// ReqmdfilesMap corresponds to the structure in reqmdfiles.json, mapping file URLs to their current hash.
type ReqmdfilesMap map[string]string
