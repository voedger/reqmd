package internal

import (
	"testing"
)

func TestAnalyzer_SemanticErrors(t *testing.T) {
	t.Run("should detect duplicate requirement IDs", func(t *testing.T) {
		// Setup
		analyzer := NewAnalyzer()
		files := []FileStructure{
			{
				Type:      FileTypeMarkdown,
				Path:      "file1.md",
				PackageID: "pkg1",
				Requirements: []RequirementSite{
					{
						RequirementName: "REQ1",
						Line:            10,
					},
				},
			},
			{
				Type:      FileTypeMarkdown,
				Path:      "file2.md",
				PackageID: "pkg1",
				Requirements: []RequirementSite{
					{
						RequirementName: "REQ1",
						Line:            5,
					},
				},
			},
		}

		// Act
		_, errors := analyzer.Analyze(files)

		if len(errors) != 1 {
			t.Fatalf("Expected 1 error, got %d", len(errors))
		}
		if errors[0].Code != "dupreqid" {
			t.Errorf("Expected error code 'dupreqid', got '%s'", errors[0].Code)
		}
	})

	t.Run("should detect missing package ID when requirements exist", func(t *testing.T) {
		// Setup
		analyzer := NewAnalyzer()
		files := []FileStructure{
			{
				Type:      FileTypeMarkdown,
				Path:      "file1.md",
				PackageID: "", // Missing package ID
				Requirements: []RequirementSite{
					{
						RequirementName: "REQ1",
						Line:            10,
					},
				},
			},
		}

		// Act
		_, errors := analyzer.Analyze(files)

		if len(errors) != 1 {
			t.Fatalf("Expected 1 error, got %d", len(errors))
		}
		if errors[0].Code != "nopkgidreqs" {
			t.Errorf("Expected error code 'nopkgidreqs', got '%s'", errors[0].Code)
		}
		if errors[0].Line != 10 {
			t.Errorf("Expected error on line 10, got line %d", errors[0].Line)
		}
	})
}
