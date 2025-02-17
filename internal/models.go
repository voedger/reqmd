package internal

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ActionType represents the type of transformation needed.
type ActionType string

const (
	ActionAnnotate ActionType = "Annotate"
)

// Action describes a single transformation (add/update/delete) to be applied in a file.
type Action struct {
	Type       ActionType     // e.g., Add, Update, Delete
	FileStruct *FileStructure // which file is changed
	Line       int            // the line number where the change is applied
	Data       string         // new data (if any)
}

// String returns a human-readable representation of the Action
func (a *Action) String() string {
	switch a.Type {
	case ActionAnnotate:
		return fmt.Sprintf("%s at %s:%d: %s", a.Type, a.FileStruct.Path, a.Line, a.Data)
	default:
		return fmt.Sprintf("Unknown action at %s:%d", a.FileStruct.Path, a.Line)
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

// Reqmdjson models the structure of the reqmd.json file.
type Reqmdjson struct {
	FileHashes map[string]string //
}

// MarshalJSON implements custom JSON serialization for Reqmdjson
// to ensure FileURLs are ordered lexically
func (r *Reqmdjson) MarshalJSON() ([]byte, error) {
	// Get all keys and sort them
	keys := make([]string, 0, len(r.FileHashes))
	for k := range r.FileHashes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build ordered map manually
	var b strings.Builder
	b.WriteString(`{"FileHashes":{`)
	for i, k := range keys {
		if i > 0 {
			b.WriteString(",")
		}
		// Marshal key and value properly to handle special characters
		keyJSON, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		valueJSON, err := json.Marshal(r.FileHashes[k])
		if err != nil {
			return nil, err
		}
		b.Write(keyJSON)
		b.WriteString(":")
		b.Write(valueJSON)
	}
	b.WriteString("}}")
	return []byte(b.String()), nil
}

// UnmarshalJSON implements custom JSON deserialization for Reqmdjson, in fact it would work without it.
// Reasons to keep the custom `UnmarshalJSON`:
// 1. **Symmetry** - We have a custom `MarshalJSON` that ensures lexical ordering of keys. It's good practice to have matching marshal/unmarshal methods for consistency.
// 2. **Future-proofing** - If we later add validation or transformation logic during unmarshaling, having the method already in place makes it easier.
// 3. **Explicit contract** - The custom method makes it clear how the JSON deserialization should behave, even if it currently matches the default behavior.
// So while removing it would work technically, we keep it for maintainability and clarity.
func (r *Reqmdjson) UnmarshalJSON(data []byte) error {
	// Use a temporary type to avoid infinite recursion
	type TempReqmdjson struct {
		FileHashes map[string]string `json:"FileHashes"`
	}
	var temp TempReqmdjson
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	r.FileHashes = temp.FileHashes
	return nil
}
