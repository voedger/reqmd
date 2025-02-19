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
		assert.Contains(t, err.Message, "file1.md:10")
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

func TestAnalyzer_ActionFootnote_NewCoverage(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with one requirement
	mdFile := createFileStructure("req.md", "pkg1", "REQ001", 10)

	// Create a source file with coverage for that requirement
	srcFile := createSourceFileStructure(
		"src/impl.go",
		"https://github.com/org/repo/blob/main/src/impl.go",
		[]CoverageTag{
			createCoverageTag("pkg1/REQ001", "impl", 20),
		},
	)

	result, err := analyzer.Analyze([]FileStructure{mdFile, srcFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate both a footnote and status update action
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[0].Type)
	assert.Equal(t, "REQ001", actions[0].RequirementName)
	assert.Contains(t, actions[0].Data, "[^~REQ001~]")
	assert.Contains(t, actions[0].Data, "src/impl.go:20:impl")

	// Verify status update action
	assert.Equal(t, ActionSite, actions[1].Type)
	assert.Equal(t, "REQ001", actions[1].RequirementName)
	assert.Equal(t, 10, actions[1].Line)
	assert.Contains(t, actions[1].Data, "covered")
}

func TestAnalyzer_ActionStatusUpdate_ExistingCoverage(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with an annotated requirement but incorrect status
	mdFile := createFileStructure("req.md", "pkg1", "REQ001", 10)
	mdFile.Requirements[0].IsAnnotated = true
	mdFile.Requirements[0].CoverageStatusWord = CoverageStatusWordUncvrd

	// Create source files with coverage
	srcFile := createSourceFileStructure(
		"src/impl.go",
		"https://github.com/org/repo/blob/main/src/impl.go#L20",
		[]CoverageTag{
			createCoverageTag("pkg1/REQ001", "impl", 20),
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
	assert.Equal(t, "REQ001", actions[0].RequirementName)
	assert.Equal(t, 10, actions[0].Line)
	assert.Contains(t, actions[0].Data, "covered")

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[1].Type)
	assert.Equal(t, "REQ001", actions[1].RequirementName)
	assert.Contains(t, actions[1].Data, "[^~REQ001~]")
	assert.Contains(t, actions[1].Data, "src/impl.go:20:impl")
	assert.Contains(t, actions[1].Data, "https://github.com/org/repo/blob/main/src/impl.go#L20")
}

func TestAnalyzer_ActionStatusUpdate_RemoveCoverage(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with covered requirement but no actual coverage
	mdFile := createFileStructure("req.md", "pkg1", "REQ001", 10)
	mdFile.Requirements[0].IsAnnotated = true
	mdFile.Requirements[0].CoverageStatusWord = CoverageStatusWordCovered

	result, err := analyzer.Analyze([]FileStructure{mdFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate status update action to uncovered
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 1)

	assert.Equal(t, ActionSite, actions[0].Type)
	assert.Equal(t, "pkg1/REQ001", actions[0].RequirementName)
	assert.Equal(t, 10, actions[0].Line)
	assert.Contains(t, actions[0].Data, "uncvrd")
}

func TestAnalyzer_ActionFootnote_UpdateExisting(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with a requirement and existing footnote
	mdFile := createFileStructure("req.md", "pkg1", "REQ001", 10)
	mdFile.CoverageFootnotes = []CoverageFootnote{
		{
			RequirementName: "REQ001",
			Line:            20,
			Coverers: []Coverer{
				{
					CoverageLabel: "old/file.go:15:impl",
					CoverageURL:   "https://github.com/org/repo/blob/main/old/file.go",
					FileHash:      "oldhash",
				},
			},
		},
	}

	// Create a source file with new coverage
	srcFile := createSourceFileStructure(
		"src/impl.go",
		"https://github.com/org/repo/blob/main",
		[]CoverageTag{
			createCoverageTag("pkg1/REQ001", "impl", 20),
		},
	)

	result, err := analyzer.Analyze([]FileStructure{mdFile, srcFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify footnote update action uses existing line number
	assert.Equal(t, ActionFootnote, actions[0].Type)
	assert.Equal(t, "pkg1/REQ001", actions[0].RequirementName)
	assert.Equal(t, 20, actions[0].Line)
	assert.Contains(t, actions[0].Data, "src/impl.go:20:impl")
	assert.NotContains(t, actions[0].Data, "old/file.go")
}

// Helper function to create a simple FileStructure with one requirement
func createFileStructure(path, pkgID string, reqName string, line int) FileStructure {
	return FileStructure{
		Path:      path,
		Type:      FileTypeMarkdown,
		PackageID: pkgID,
		Requirements: []RequirementSite{
			{
				FilePath:        path,
				RequirementName: reqName,
				Line:            line,
			},
		},
	}
}

// Helper function to create a CoverageTag
func createCoverageTag(reqID, coverageType string, line int) CoverageTag {
	return CoverageTag{
		RequirementID: reqID,
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
