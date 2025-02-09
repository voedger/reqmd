package internal

// MarkdownFile represents a markdown file that contains a header and a body.
// The header specifies the PackageID and the body includes requirement references
// and coverage footnotes.
type MarkdownFile struct {
	FilePath     string             // Absolute or relative file path
	PackageID    string             // Extracted from the header (e.g., "server.api.v2")
	Header       string             // The raw header content
	Body         string             // The markdown content excluding the header
	Requirements []Requirement      // List of requirement references in the file
	Footnotes    []CoverageFootnote // List of coverage footnotes generated or present in the file
}

// SourceFile represents a source file that may contain multiple coverage tags.
type SourceFile struct {
	FilePath     string        // Absolute or relative file path
	Content      string        // Raw source code content
	CoverageTags []CoverageTag // Coverage tags extracted from the file (e.g., [~server.api.v2/Post.handler~test])
}

// Requirement represents a requirement extracted from a markdown file.
// It may appear as a bare requirement or as a requirement site (with coverage annotation).
type Requirement struct {
	PackageID string // Inherited from the markdown file header (e.g., "server.api.v2")
	Name      string // The requirement name extracted from the text (e.g., "Post.handler")
	IsSite    bool   // Indicates if this requirement has a coverage annotation (RequirementSite)
}

// FullID returns the computed RequirementID, formed as PackageID/RequirementName.
func (r Requirement) FullID() string {
	return r.PackageID + "/" + r.Name
}

// CoverageTag represents a tag found in source files that is used to indicate
// that the source code covers a particular requirement.
type CoverageTag struct {
	// For example, given the tag [~server.api.v2/Post.handler~impl]:
	// RequirementID would be "server.api.v2/Post.handler"
	// CoverageType would be "impl" (implementation) or "test"
	RequirementID string
	CoverageType  string
}

// CoverageFootnote represents a footnote in a markdown file that links a requirement
// to its corresponding coverage tags. It contains a hint and a list of coverers.
type CoverageFootnote struct {
	// Hint is typically the CoverageFootnoteHint (e.g., "[~server.api.v2~impl]")
	Hint string
	// Coverers is the list of coverers linked in the footnote.
	Coverers []Coverer
}

// Coverer represents a coverage mapping element that appears in a footnote.
// It binds a coverage label to a CoverageURL.
type Coverer struct {
	// CoverageLabel is the text label indicating file location and context (e.g., "folder1/filename1:line1:impl")
	CoverageLabel string
	// CoverageURL is the URL linking to the file (e.g., a GitHub or GitLab URL with a CoverageArea anchor)
	CoverageURL string
}

// ActionType defines the type of action that can be applied to modify files.
type ActionType int

const (
	// ActionAdd indicates that new data should be added.
	ActionAdd ActionType = iota
	// ActionUpdate indicates that existing data should be modified.
	ActionUpdate
	// ActionDelete indicates that data should be removed.
	ActionDelete
)

// Action represents an update operation to be performed on an input file.
// It captures the type of change, the location, and the new data involved.
type Action struct {
	Type     ActionType // The type of action (Add, Update, Delete)
	FilePath string     // Path of the file where the action will be applied
	Line     int        // The line number in the file where the action applies
	Data     string     // The new data to insert or update with
}

// SyntaxError represents an error detected during the parsing (scanning) phase.
type SyntaxError struct {
	FilePath string // The file in which the error occurred
	Line     int    // The line number where the error was detected
	Message  string // A description of the syntax error
}

// SemanticError represents an error related to the semantics of the requirements,
// such as duplicate RequirementIDs across markdown files.
type SemanticError struct {
	FilePath string // The file in which the error was detected
	Line     int    // The line number where the error was detected
	Message  string // A description of the semantic error
}

// ReqmdFiles represents a mapping of FileURLs to their corresponding FileHashes.
// It is used to generate or update the reqmdfiles.json file.
type ReqmdFiles map[string]string
