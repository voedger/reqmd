package internal

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ActionType represents the type of transformation needed.
type ActionType string

const (
	ActionAnnotate ActionType = "Annotate"
)

// Action describes a single transformation (add/update/delete) to be applied in a file.
type Action struct {
	Type     ActionType // e.g., Add, Update, Delete
	FilePath string     // which file is changed
	Line     int        // the line number where the change is applied
	Data     string     // new data (if any)
}

// String returns a human-readable representation of the Action
func (a *Action) String() string {
	switch a.Type {
	case ActionAnnotate:
		return fmt.Sprintf("%s at %s:%d: %s", a.Type, a.FilePath, a.Line, a.Data)
	default:
		return fmt.Sprintf("Unknown action at %s:%d", a.FilePath, a.Line)
	}
}

// SyntaxError captures syntax and semantic errors.
type ProcessingError struct {
	Code     string // error code (e.g., "pkgident")
	FilePath string // file that has a syntax error
	Line     int    // line number where the syntax error is detected
	Message  string // human-readable description
}

// Collection of ProcessingErrors
// Implements Error interface
type ProcessingErrors struct {
	Errors []ProcessingError
}

func (e *ProcessingErrors) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}

	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, fmt.Sprintf("%s:%d: %s", err.FilePath, err.Line, err.Message))
	}
	return strings.Join(msgs, "\n")
}

// FileType distinguishes between different file categories (Markdown vs source, etc.).
type FileType int

const (
	FileTypeMarkdown FileType = iota
	FileTypeSource
)

// FileStructure merges the parsed data from an input file (Markdown or source).
type FileStructure struct {
	Path              string
	Type              FileType           // indicates if it's Markdown or source
	PackageID         string             // parsed from Markdown header (if markdown)
	Requirements      []RequirementSite  // for Markdown: discovered requirements (bare or annotated)
	CoverageFootnotes []CoverageFootnote // for Markdown: discovered coverage footnotes
	CoverageTags      []CoverageTag      // for source: discovered coverage tags
	FileHash          string             // git hash of the file
	RepoRootFolderURL string
	RelativePath      string
}

func (f *FileStructure) FileURL() string {
	return f.RepoRootFolderURL + "/" + filepath.ToSlash(f.RelativePath)
}

// RequirementSite represents a single requirement reference discovered in a Markdown file.
type RequirementSite struct {
	FilePath            string
	Line                int    // line number where the requirement is defined/referenced
	RequirementName     string // e.g., "Post.handler"
	ReferenceName       string // Other.handler for "`~Post.handler~`cov[^~Other.handler~]"
	CoverageStatusWord  string // "covered", "uncvrd", or empty
	CoverageStatusEmoji string // "✅", "❓", or empty
	IsAnnotated         bool   // true if it already has coverage annotation, false if it’s bare
}

// CoverageTag represents a coverage marker found in source code.
type CoverageTag struct {
	RequirementID string // e.g., "server.api.v2/Post.handler"
	CoverageType  string // e.g., "impl", "test"
	Line          int    // line number where the coverage tag was found
}

// CoverageFootnote represents the footnote in Markdown that references coverage tags.
type CoverageFootnote struct {
	FilePath      string
	Line          int
	PackageID     string
	RequirementID string
	Coverers      []Coverer
}

// Coverer represents one coverage reference within a footnote, e.g., [folder/file:line:impl](URL)
type Coverer struct {
	CoverageLabel string // e.g., "folder/file.go:42:impl"
	CoverageURL   string // full URL including commit hash
	FileHash      string // git hash of the file specified in CoverageURL
}

// ReqmdfilesMap corresponds to the structure in reqmdfiles.json, mapping file URLs to their current hash.
type ReqmdfilesMap map[string]string
