package internal

import (
	"testing"
)

func TestAnalyzer_SemanticErrors(t *testing.T) {
	t.Run("should detect duplicate requirement IDs", func(t *testing.T) {
		// Setup
		analyzer := NewAnalyzer()
		files := []FileStructure{
			_mdfile1("file1.md", "pkg1", "REQ1", 10),
			_mdfile1("file2.md", "pkg1", "REQ1", 5),
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
			_mdfile1("file1.md", "", "REQ1", 10),
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

// Helper function to create a FileStructure with one requirement
func _mdfile1(path, packageID, reqName string, line int) FileStructure {
	return FileStructure{
		Type:      FileTypeMarkdown,
		Path:      path,
		PackageID: packageID,
		Requirements: []RequirementSite{
			{
				RequirementName: reqName,
				Line:            line,
			},
		},
	}
}

// Helper function to create a FileStructure with one requirement and optional coverage footnote
func _mdfile2(path, packageID, reqName string, line int, statusWord CoverageStatusWord, isAnnotated bool, coverFootnotes ...CoverageFootnote) FileStructure {
	return FileStructure{
		Type:      FileTypeMarkdown,
		Path:      path,
		PackageID: packageID,
		Requirements: []RequirementSite{
			{
				RequirementName:    reqName,
				Line:               line,
				IsAnnotated:        isAnnotated,
				CoverageStatusWord: statusWord,
			},
		},
		CoverageFootnotes: coverFootnotes,
	}
}

// Helper function to create a source FileStructure with coverage tags
func _srcfile(path string, relativePath string, fileHash string, tags ...CoverageTag) FileStructure {
	return FileStructure{
		Type:         FileTypeSource,
		Path:         path,
		RelativePath: relativePath,
		FileHash:     fileHash,
		CoverageTags: tags,
	}
}

func TestAnalyzer_Actions(t *testing.T) {
	t.Run("should generate footnote action for new coverage", func(t *testing.T) {
		analyzer := NewAnalyzer()
		files := []FileStructure{
			_mdfile2("file1.md", "pkg1", "REQ1", 10, "covered", true, CoverageFootnote{
				RequirementID: "pkg1/REQ1",
				PackageID:     "pkg1",
				Coverers: []Coverer{
					{CoverageLabel: "old/path:10:test", CoverageURL: "http://old/url", FileHash: "oldhash"},
				},
			}),
			_srcfile("src/new.go", "src/new.go", "newhash", CoverageTag{
				RequirementID: "pkg1/REQ1",
				CoverageType:  "test",
				Line:          20,
			}),
		}

		actions, errors := analyzer.Analyze(files)
		if len(errors) > 0 {
			t.Fatalf("Expected no errors, got %v", errors)
		}

		// Should generate ActionFootnote and possibly ActionAddFileURL/ActionUpdateHash
		var foundFootnote bool
		var foundFileAction bool
		for _, action := range actions {
			switch action.Type {
			case ActionFootnote:
				foundFootnote = true
			case ActionAddFileURL, ActionUpdateHash:
				foundFileAction = true
			}
		}

		if !foundFootnote {
			t.Error("Expected to find ActionFootnote")
		}
		if !foundFileAction {
			t.Error("Expected to find ActionAddFileURL or ActionUpdateHash")
		}
	})

	t.Run("should generate status update action for uncovered requirement", func(t *testing.T) {
		analyzer := NewAnalyzer()
		files := []FileStructure{
			_mdfile2("file1.md", "pkg1", "REQ1", 10, "covered", true, CoverageFootnote{
				RequirementID: "pkg1/REQ1",
				PackageID:     "pkg1",
				Coverers: []Coverer{
					{CoverageLabel: "old/path:10:test", CoverageURL: "http://old/url", FileHash: "oldhash"},
				},
			}),
		}

		actions, errors := analyzer.Analyze(files)
		if len(errors) > 0 {
			t.Fatalf("Expected no errors, got %v", errors)
		}

		// Should generate ActionFootnote and ActionUpdateStatus
		var foundFootnote bool
		var foundStatus bool
		for _, action := range actions {
			switch action.Type {
			case ActionFootnote:
				foundFootnote = true
			case ActionUpdateStatus:
				foundStatus = true
				if action.Data != string(CoverageStatusWordUncvrd) {
					t.Errorf("Expected status update to uncvrd, got %s", action.Data)
				}
			}
		}

		if !foundFootnote {
			t.Error("Expected to find ActionFootnote")
		}
		if !foundStatus {
			t.Error("Expected to find ActionUpdateStatus")
		}
	})

	t.Run("should generate annotate action for bare requirement", func(t *testing.T) {
		analyzer := NewAnalyzer()
		files := []FileStructure{
			_mdfile2("file1.md", "pkg1", "REQ1", 10, "", false),
		}

		actions, errors := analyzer.Analyze(files)
		if len(errors) > 0 {
			t.Fatalf("Expected no errors, got %v", errors)
		}
		if len(actions) != 1 {
			t.Fatalf("Expected 1 action, got %d", len(actions))
		}
		if actions[0].Type != ActionAnnotate {
			t.Errorf("Expected ActionAnnotate, got %v", actions[0].Type)
		}
	})

	// t.Run("should generate file URL action for new source file", func(t *testing.T) {
	// 	analyzer := NewAnalyzer()
	// 	files := []FileStructure{
	// 		_mdfile2("file1.md", "pkg1", "REQ1", 10, "", true),
	// 		_srcfile("src/new.go", "src/new.go", "newhash", CoverageTag{
	// 			RequirementID: "pkg1/REQ1",
	// 			CoverageType:  "test",
	// 			Line:          20,
	// 		}),
	// 	}

	// 	actions, errors := analyzer.Analyze(files)
	// 	if len(errors) > 0 {
	// 		t.Fatalf("Expected no errors, got %v", errors)
	// 	}

	// 	var foundAddFileURL bool
	// 	for _, action := range actions {
	// 		if action.Type == ActionAddFileURL {
	// 			foundAddFileURL = true
	// 			if action.Data != "newhash" {
	// 				t.Errorf("Expected file hash newhash, got %s", action.Data)
	// 			}
	// 			break
	// 		}
	// 	}

	// 	if !foundAddFileURL {
	// 		t.Error("Expected to find ActionAddFileURL")
	// 	}
	// })

	t.Run("should generate hash update action for changed file", func(t *testing.T) {
		analyzer := NewAnalyzer()
		files := []FileStructure{
			_mdfile2("file1.md", "pkg1", "REQ1", 10, "", true),
			_srcfile("src/existing.go", "src/existing.go", "newhash", CoverageTag{
				RequirementID: "pkg1/REQ1",
				CoverageType:  "test",
				Line:          20,
			}),
		}
		// Set RepoRootFolderURL to simulate existing file
		files[1].RepoRootFolderURL = "http://example.com"

		actions, errors := analyzer.Analyze(files)
		if len(errors) > 0 {
			t.Fatalf("Expected no errors, got %v", errors)
		}

		var foundUpdateHash bool
		for _, action := range actions {
			if action.Type == ActionUpdateHash {
				foundUpdateHash = true
				if action.Data != "newhash" {
					t.Errorf("Expected file hash newhash, got %s", action.Data)
				}
				break
			}
		}

		if !foundUpdateHash {
			t.Error("Expected to find ActionUpdateHash")
		}
	})
}
