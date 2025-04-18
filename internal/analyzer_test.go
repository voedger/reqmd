// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzer_error_Duplicates(t *testing.T) {
	analyzer := NewAnalyzer()

	files := []FileStructure{
		{
			Path:      "file1.md",
			Type:      FileTypeMarkdown,
			PackageID: "pkg1",
			Requirements: []RequirementSite{
				{
					FilePath:        "file1.md",
					RequirementName: "REQ001",
					Line:            10,
				},
			},
		},
		{
			Path:      "file2.md",
			Type:      FileTypeMarkdown,
			PackageID: "pkg1",
			Requirements: []RequirementSite{
				{
					FilePath:        "file2.md",
					RequirementName: "REQ001",
					Line:            20,
				},
			},
		},
	}

	result, err := analyzer.Analyze(files)
	assert.NoError(t, err)
	assert.Len(t, result.ProcessingErrors, 1)

	if assert.NotEmpty(t, result.ProcessingErrors) {
		err := result.ProcessingErrors[0]
		assert.Equal(t, "dupreqid", err.Code)
		assert.Equal(t, "file1.md", err.FilePath)
		assert.Equal(t, 10, err.Line)
		assert.Contains(t, err.Message, "pkg1/REQ001")
		assert.NotContains(t, err.Message, "file1.md:10")
		assert.Contains(t, err.Message, "file2.md:20")
	}
}

func TestAnalyzer_error_MissingPackageID(t *testing.T) {
	analyzer := NewAnalyzer()

	files := []FileStructure{
		{
			Path: "file1.md",
			Type: FileTypeMarkdown,
			Requirements: []RequirementSite{
				{
					FilePath:        "file1.md",
					RequirementName: "REQ001",
					Line:            10,
				},
			},
		},
	}

	result, err := analyzer.Analyze(files)
	assert.NoError(t, err)
	assert.Len(t, result.ProcessingErrors, 1)

	if assert.NotEmpty(t, result.ProcessingErrors) {
		err := result.ProcessingErrors[0]
		assert.Equal(t, "nopkgidreqs", err.Code)
		assert.Equal(t, "file1.md", err.FilePath)
		assert.Equal(t, 10, err.Line)
		assert.Contains(t, err.Message, "shall define reqmd.package")
	}
}

// Bare requirement
func TestAnalyzer_ActionFootnote_Bare(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with one requirement
	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ001", CoverageStatusWordUncvrd)
	mdFile.Requirements[0].HasAnnotationRef = false

	result, err := analyzer.Analyze([]FileStructure{mdFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate both a footnote and status update action
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify status update action
	assert.Equal(t, ActionSite, actions[0].Type)
	assert.Equal(t, RequirementName("REQ001"), actions[0].RequirementName)
	assert.Equal(t, 10, actions[0].Line)
	assert.Equal(t, FormatRequirementSite("REQ001", CoverageStatusWordUncvrd, "1"), actions[0].Data)

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[1].Type)
	assert.Equal(t, RequirementName("REQ001"), actions[1].RequirementName)
	assert.Equal(t, "[^1]: `[~pkg1/REQ001~impl]`", actions[1].Data)
}

// Bare requirement, there is a footnote with CoverageFootnoteID == "19"
func TestAnalyzer_ActionFootnote_Bare_f19(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with one requirement
	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ001", CoverageStatusWordUncvrd)
	mdFile.Requirements[0].HasAnnotationRef = false

	mdFile.CoverageFootnotes = []CoverageFootnote{
		{
			CoverageFootnoteId: CoverageFootnoteId("19"),
			Line:               10 + 10,
		},
	}

	result, err := analyzer.Analyze([]FileStructure{mdFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate both a footnote and status update action
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify status update action
	assert.Equal(t, ActionSite, actions[0].Type)
	assert.Equal(t, RequirementName("REQ001"), actions[0].RequirementName)
	assert.Equal(t, 10, actions[0].Line)
	assert.Equal(t, FormatRequirementSite("REQ001", CoverageStatusWordUncvrd, "20"), actions[0].Data)

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[1].Type)
	assert.Equal(t, RequirementName("REQ001"), actions[1].RequirementName)
	assert.Equal(t, "[^20]: `[~pkg1/REQ001~impl]`", actions[1].Data)
}

// Bare requirement, there is a footnote with CoverageFootnoteID == "19"
// and a RequirementSite with CoverageFootnoteID == "21".
func TestAnalyzer_ActionFootnote_Bare_f19_r21(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with one requirement
	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ001", CoverageStatusWordUncvrd)
	mdFile.Requirements[0].HasAnnotationRef = false

	mdFile.CoverageFootnotes = []CoverageFootnote{
		{
			CoverageFootnoteId: CoverageFootnoteId("19"),
			Line:               10 + 10,
		},
	}
	// Add a RequirementSite with CoverageFootnoteID == "21"
	{
		req := RequirementSite{
			FilePath:            "req.md",
			RequirementName:     RequirementName("REQ002"),
			CoverageFootnoteID:  CoverageFootnoteId("21"),
			Line:                11,
			HasAnnotationRef:    true,
			CoverageStatusWord:  CoverageStatusWordUncvrd,
			CoverageStatusEmoji: CoverageStatusEmojiUncvrd,
		}
		mdFile.Requirements = append(mdFile.Requirements, req)
	}

	result, err := analyzer.Analyze([]FileStructure{mdFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate both a footnote and status update action
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 3)

	for _, action := range actions {
		if action.Type == ActionSite {
			// Verify status update action
			assert.Equal(t, ActionSite, action.Type)
			assert.Equal(t, RequirementName("REQ001"), action.RequirementName)
			assert.Equal(t, 10, action.Line)
			assert.Equal(t, FormatRequirementSite("REQ001", CoverageStatusWordUncvrd, "22"), action.Data)
			continue
		}
		if action.Type == ActionFootnote && action.RequirementName == "REQ001" {
			assert.Equal(t, ActionFootnote, action.Type)
			assert.Equal(t, RequirementName("REQ001"), action.RequirementName)
			assert.Equal(t, "[^22]: `[~pkg1/REQ001~impl]`", action.Data)
		}
		if action.Type == ActionFootnote && action.RequirementName == "REQ002" {
			assert.Equal(t, ActionFootnote, action.Type)
			assert.Equal(t, RequirementName("REQ002"), action.RequirementName)
			assert.Equal(t, "[^21]: `[~pkg1/REQ002~impl]`", action.Data)
		}
	}

}

// Bare requirement with new coverer
func TestAnalyzer_ActionFootnote_Bare_NewCoverer(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with one requirement
	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ001", CoverageStatusWordUncvrd)
	mdFile.Requirements[0].HasAnnotationRef = false

	// Create a source file with coverage for that requirement
	srcFile := createSourceFileStructure(
		"src/impl.go",
		"https://github.com/org/repo/blob/main/src/impl.go",
		[]CoverageTag{
			createCoverageTag("pkg1/REQ001", "impl1", 20),
		},
	)

	result, err := analyzer.Analyze([]FileStructure{mdFile, srcFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate both a footnote and status update action
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify status update action
	assert.Equal(t, ActionSite, actions[0].Type)
	assert.Equal(t, RequirementName("REQ001"), actions[0].RequirementName)
	assert.Equal(t, 10, actions[0].Line)
	assert.Equal(t, actions[0].Data, FormatRequirementSite("REQ001", CoverageStatusWordCovered, "1"))

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[1].Type)
	assert.Equal(t, RequirementName("REQ001"), actions[1].RequirementName)
	assert.Equal(t, "[^1]: `[~pkg1/REQ001~impl]` [src/impl.go:20:impl1](https://github.com/org/repo/blob/main/src/impl.go/src/impl.go#L20)", actions[1].Data)

}

// Annotated uncovered requirement
func TestAnalyzer_ActionStatusUpdate_AnUncov_NewCoverer(t *testing.T) {
	analyzer := NewAnalyzer()

	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ002", CoverageStatusWordUncvrd)

	// Create source files with coverage
	srcFile := createSourceFileStructure(
		"src/impl.go",
		"https://github.com/org/repo/blob/main/src/impl.go#L20",
		[]CoverageTag{
			createCoverageTag("pkg1/REQ002", "impl", 20),
		},
	)

	result, err := analyzer.Analyze([]FileStructure{mdFile, srcFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should only generate status update action since requirement is already annotated
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify site action
	assert.Equal(t, ActionSite, actions[0].Type)
	assert.Equal(t, RequirementName("REQ002"), actions[0].RequirementName)
	assert.Equal(t, 10, actions[0].Line)
	assert.Equal(t, actions[0].Data, FormatRequirementSite("REQ002", CoverageStatusWordCovered, "REQ002"))

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[1].Type)
	assert.Equal(t, RequirementName("REQ002"), actions[1].RequirementName)
	assert.Contains(t, actions[1].Data, "[^REQ002]:")
	assert.Contains(t, actions[1].Data, "src/impl.go:20:impl")
	assert.Contains(t, actions[1].Data, "https://github.com/org/repo/blob/main/src/impl.go#L20")
}

func TestAnalyzer_ActionStatusUpdate_AnCov_NoCoverers(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with covered requirement but no actual coverage
	mdFile := createMdStructureA("req.md", "pkg1", 11, "REQ001", CoverageStatusWordCovered)

	result, err := analyzer.Analyze([]FileStructure{mdFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate status update action to uncovered
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify site action
	assert.Equal(t, ActionSite, actions[0].Type)
	assert.Equal(t, RequirementName("REQ001"), actions[0].RequirementName)
	assert.Equal(t, 11, actions[0].Line)
	assert.Contains(t, actions[0].Data, FormatRequirementSite("REQ001", CoverageStatusWordUncvrd, "REQ001"))

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[1].Type)
	assert.Equal(t, actions[1].Data, "[^REQ001]: `[~pkg1/REQ001~impl]`")
	assert.Equal(t, 11+10, actions[1].Line)
}

func TestAnalyzer_ActionFootnote_AnCov_NewHash(t *testing.T) {
	analyzer := NewAnalyzer()

	OldCoverageURL := "https://github.com/org/repo/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/old/file.go"
	NewCoverageURL := "https://github.com/org/repo/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/new/file.go"

	// Create a markdown file with a requirement and existing footnote
	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ001", CoverageStatusWordCovered)
	mdFile.CoverageFootnotes = []CoverageFootnote{
		{
			CoverageFootnoteId: "REQ001",
			Line:               20,
			Coverers: []Coverer{
				{
					CoverageLabel: "old/file.go:15:impl",
					CoverageUrL:   OldCoverageURL,
				},
			},
		},
	}

	// Source file with the same Url but new hash
	srcFile := createSourceFileStructure(
		"src/impl.go",
		NewCoverageURL,
		[]CoverageTag{
			createCoverageTag("pkg1/REQ001", "impl", 20),
		},
	)

	result, err := analyzer.Analyze([]FileStructure{mdFile, srcFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 1)

	// Verify footnote update action uses existing line number
	assert.Equal(t, ActionFootnote, actions[0].Type)
	assert.NotContains(t, actions[0].Data, OldCoverageURL)
	assert.Contains(t, actions[0].Data, NewCoverageURL)
}

// Helper function to create a simple FileStructure with one annotated requirement that has cw coverage
func createMdStructureA(path, pkgID string, line int, reqName_ string, cw CoverageStatusWord) FileStructure {

	var reqName RequirementName = RequirementName(reqName_)
	var footnoteID CoverageFootnoteId = CoverageFootnoteId(reqName_)

	emoji := CoverageStatusEmojiUncvrd
	if cw == CoverageStatusWordCovered {
		emoji = CoverageStatusEmojiCovered
	}

	fs := FileStructure{
		Path:      path,
		Type:      FileTypeMarkdown,
		PackageID: pkgID,
		Requirements: []RequirementSite{
			{
				FilePath:            path,
				RequirementName:     reqName,
				CoverageFootnoteID:  footnoteID,
				Line:                line,
				HasAnnotationRef:    true,
				CoverageStatusWord:  cw,
				CoverageStatusEmoji: emoji,
			},
		},
	}

	if cw == CoverageStatusWordCovered {
		fs.CoverageFootnotes = []CoverageFootnote{
			{
				CoverageFootnoteId: footnoteID,
				Line:               line + 10,
				Coverers: []Coverer{
					{
						CoverageLabel: "somefolder/somefile.go:15:impl",
						CoverageUrL:   "someurl",
					},
				},
			},
		}
	}

	return fs
}

// Helper function to create a CoverageTag
func createCoverageTag(reqID RequirementId, coverageType string, line int) CoverageTag {
	return CoverageTag{
		RequirementId: reqID,
		CoverageType:  coverageType,
		Line:          line,
	}
}

// Helper function to create a source FileStructure
func createSourceFileStructure(path string, repoRootURL string, tags []CoverageTag) FileStructure {
	return FileStructure{
		Path:              path,
		Type:              FileTypeSource,
		CoverageTags:      tags,
		FileHash:          "hash1",
		RepoRootFolderURL: repoRootURL,
		RelativePath:      path, // For simplicity in tests
	}
}
