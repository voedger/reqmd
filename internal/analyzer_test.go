package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
