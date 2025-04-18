// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type CoverageStatusWord string

const (
	CoverageStatusWordEmpty   CoverageStatusWord = ""
	CoverageStatusWordCovered CoverageStatusWord = "covered"
	CoverageStatusWordUncvrd  CoverageStatusWord = "uncvrd"
)

type RequirementName string
type RequirementId string
type CoverageFootnoteId string
type FilePath = string
type FolderPath = string

// SyntaxError captures syntax and semantic errors.

// FileType distinguishes between different file categories (Markdown vs source, etc.).
type FileType int

const (
	FileTypeMarkdown FileType = iota
	FileTypeSource
)

type CoverageStatusEmoji string

const (
	CoverageStatusEmojiEmpty   CoverageStatusEmoji = ""
	CoverageStatusEmojiCovered CoverageStatusEmoji = "✅"
	CoverageStatusEmojiUncvrd  CoverageStatusEmoji = "❓"
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
	Line                int                // line number where the requirement is defined/referenced
	RequirementName     RequirementName    // e.g., "Post.handler"
	CoverageFootnoteID  CoverageFootnoteId // Other.handler for "`~Post.handler~`cov[^~Other.handler~]"
	CoverageStatusWord  CoverageStatusWord // "covered", "uncvrd", or empty
	CoverageStatusEmoji CoverageStatusEmoji
	HasAnnotationRef    bool // true if it already has coverage annotation reference, false if it’s bare
}

var RequirementSiteRegex = regexp.MustCompile(
	"`~([^~]+)~`" + // RequirementSiteLabel = "`" "~" RequirementName "~" "`"
		"(?:" + // Optional group for coverage status and footnote
		"\\s*([a-zA-Z]+)?" + // Optional CoverageStatusWord
		"\\s*\\[\\^([^\\]]+)\\]" + // CoverageFootnoteReference
		"\\s*(✅|❓)?" + // Optional CoverageStatusEmoji
		")?")

// Build a string representation of the RequirementSite according to the requirements
// CoverageStatusEmoji is ✅ for "covered", and ❓ for "uncvrd"
func FormatRequirementSite(requirementName RequirementName, coverageStatusWord CoverageStatusWord, footnoteId CoverageFootnoteId) string {
	lbl := fmt.Sprintf("`~%s~`", requirementName)

	emoji := CoverageStatusEmojiUncvrd
	if coverageStatusWord == CoverageStatusWordCovered {
		emoji = CoverageStatusEmojiCovered
	}

	return fmt.Sprintf("%s%s[^%s]%s", lbl, coverageStatusWord, footnoteId, emoji)
}

// CoverageTag represents a coverage marker found in source code.
type CoverageTag struct {
	RequirementId RequirementId // e.g., "server.api.v2/Post.handler"
	CoverageType  string        // e.g., "impl", "test"
	Line          int           // line number where the coverage tag was found
}

func (c *CoverageTag) String() string {
	return fmt.Sprintf("%s/%s:%d", c.RequirementId, c.CoverageType, c.Line)
}

// CoverageFootnote represents the footnote in Markdown that references coverage tags.
type CoverageFootnote struct {
	FilePath           string
	Line               int
	PackageID          string
	RequirementName    RequirementName
	CoverageFootnoteId CoverageFootnoteId
	Coverers           []Coverer
}

var (
	// "[^12]: `[~com.example.basic/REQ002~impl]`[folder1/filename1:line1:impl](https://example.com/pkg1/filename1#L10), [folder2/filename2:line2:test](https://example.com/pkg2/filename2#l15)"
	CoverageFootnoteRegex = regexp.MustCompile(`^\s*\[\^([^\]]+)\]:\s*` + //Footnote reference
		"(?:`\\[~([^~/]+)/([^~]+)~([^\\]]+)?\\]`)?" + // Hint with package and coverage type
		`(?:\s*(.+))?\s*$`) // Optional coverer list
	CovererRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

// Sort Coverers according to requirements:
// - Coverers shall be sorted by CoverageType, then by FilePath, then by Number, then by CoverageURL
func sortCoverers(coverers []Coverer) {
	sort.Slice(coverers, func(i, j int) bool {
		// Split CoverageLabel to get FilePath, Number and CoverageType
		// Format is filepath:number:coveragetype
		iParts := strings.Split(coverers[i].CoverageLabel, ":")
		jParts := strings.Split(coverers[j].CoverageLabel, ":")

		if len(iParts) != 3 || len(jParts) != 3 {
			return coverers[i].CoverageLabel < coverers[j].CoverageLabel
		}

		iFilePath, iNumStr, iType := iParts[0], iParts[1], iParts[2]
		jFilePath, jNumStr, jType := jParts[0], jParts[1], jParts[2]

		// Compare CoverageType first
		if iType != jType {
			return iType < jType
		}

		// Then compare FilePath
		if iFilePath != jFilePath {
			return iFilePath < jFilePath
		}

		// Convert number strings to integers for comparison
		iNum, iErr := strconv.Atoi(iNumStr)
		jNum, jErr := strconv.Atoi(jNumStr)
		if iErr == nil && jErr == nil && iNum != jNum {
			return iNum < jNum
		}

		// Finally compare by CoverageURL
		return coverers[i].CoverageURL < coverers[j].CoverageURL
	})
}

// Helper function to format a coverage footnote
func FormatCoverageFootnote(cf *CoverageFootnote) string {
	// Sort coverers before formatting
	sortCoverers(cf.Coverers)

	var refs []string
	for _, coverer := range cf.Coverers {
		refs = append(refs, fmt.Sprintf("[%s](%s)", coverer.CoverageLabel, coverer.CoverageURL))
	}
	hint := fmt.Sprintf("`[~%s/%s~impl]`", cf.PackageID, cf.RequirementName)
	if len(refs) > 0 {
		coverersStr := strings.Join(refs, ", ")
		res := fmt.Sprintf("[^%s]: %s %s", cf.CoverageFootnoteId, hint, coverersStr)
		return res
	}
	return fmt.Sprintf("[^%s]: %s", cf.CoverageFootnoteId, hint)
}

// Coverer represents one coverage reference within a footnote, e.g., [folder/file:line:impl](URL)
type Coverer struct {
	CoverageLabel string // e.g., "folder/file.go:42:impl"
	CoverageURL   string // full URL including commit hash
	fileHash      string // git hash of the file specified in CoverageURL, not used currently
}

func FileURL(coverageURL string) string {
	idx := strings.LastIndex(coverageURL, "#L")
	if idx == -1 {
		return coverageURL
	}
	return coverageURL[:idx]
}

const ReqmdjsonFileName = "reqmd.json"

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
		msgs = append(msgs, fmt.Sprintf("%s:%d: %s: %s", err.FilePath, err.Line, err.Code, err.Message))
	}
	return strings.Join(msgs, "\n")
}

// ScannerResult contains results from the scanning phase
type ScannerResult struct {
	Files            []FileStructure
	ProcessingErrors []ProcessingError
}

// MdActionType represents the type of markdown transformation needed.
type MdActionType string

const (
	ActionFootnote MdActionType = "Footnote" // Create/Update a CoverageFootnote
	ActionSite     MdActionType = "Site"     // Update RequirementSite
)

// MdAction describes a single transformation (add/update/delete) to be applied in a file.
type MdAction struct {
	Type            MdActionType // e.g., "Footnote", "Site"
	Path            string       // file path
	Line            int          // the line number where the change is applied. 0 means the
	Data            string       // new data (if any)
	RequirementName RequirementName
}

// String returns a human-readable representation of the Action
func (a *MdAction) String() string {
	return fmt.Sprintf("%s\n\t%s:%d\n\tRequirement: %s\n\tData: %s", a.Type, a.Path, a.Line, a.RequirementName, a.Data)
}

// AnalyzerResult contains results from the analysis phase
type AnalyzerResult struct {
	MdActions        map[FilePath][]MdAction
	ProcessingErrors []ProcessingError
}
