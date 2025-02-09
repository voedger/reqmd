package internal

import (
	"path/filepath"
	"testing"
)

func TestMdParser_Scan(t *testing.T) {
	parser := NewMarkdownParser()
	testDataPath := filepath.Join("testdata")

	files, errors := parser.Scan([]string{testDataPath})

	// Verify basic.md parsing
	var basicFile *FileStructure
	for i := range files {
		if filepath.Base(files[i].Path) == "basic.md" {
			basicFile = &files[i]
			break
		}
	}

	if basicFile == nil {
		t.Fatal("basic.md not found in parsed files")
	}

	// Test package ID
	if basicFile.PackageID != "com.example.basic" {
		t.Errorf("expected package ID 'com.example.basic', got '%s'", basicFile.PackageID)
	}

	// Test requirements
	if len(basicFile.Requirements) != 3 {
		t.Errorf("expected 3 requirements, got %d", len(basicFile.Requirements))
	}

	// Verify REQ001
	found := false
	for _, req := range basicFile.Requirements {
		if req.ID == "REQ001" {
			found = true
			if req.IsAnnotated {
				t.Error("REQ001 should not be annotated")
			}
			if req.Line != 8 {
				t.Errorf("REQ001 should be on line 8, got %d", req.Line)
			}
		}
	}
	if !found {
		t.Error("REQ001 not found")
	}

	// Verify REQ002 (appears twice, one annotated)
	annotatedFound := false
	unannotatedFound := false
	for _, req := range basicFile.Requirements {
		if req.ID == "REQ002" {
			if req.IsAnnotated {
				annotatedFound = true
			} else {
				unannotatedFound = true
			}
		}
	}
	if !annotatedFound || !unannotatedFound {
		t.Error("REQ002 should appear as both annotated and unannotated")
	}

	// Verify errors from invalid.md
	hasInvalidFileError := false
	for _, err := range errors {
		if filepath.Base(err.FilePath) == "invalid.md" {
			hasInvalidFileError = true
			break
		}
	}
	if !hasInvalidFileError {
		t.Error("expected error for invalid.md")
	}
}

func TestMdParser_parseRequirements(t *testing.T) {
	parser := &mdParser{}

	tests := []struct {
		name            string
		line            string
		expectReqIDs    []string
		expectAnnotated []bool
	}{
		{
			name:            "simple requirement",
			line:            "This contains ~REQ001~",
			expectReqIDs:    []string{"REQ001"},
			expectAnnotated: []bool{false},
		},
		{
			name:            "multiple requirements",
			line:            "Contains ~REQ001~ and ~REQ002~",
			expectReqIDs:    []string{"REQ001", "REQ002"},
			expectAnnotated: []bool{false, false},
		},
		{
			name:            "annotated requirement",
			line:            "~REQ001~coverage[^1]",
			expectReqIDs:    []string{"REQ001"},
			expectAnnotated: []bool{true},
		},
		{
			name:            "mixed requirements",
			line:            "~REQ001~ normal and ~REQ002~coverage[^1] annotated",
			expectReqIDs:    []string{"REQ001", "REQ002"},
			expectAnnotated: []bool{false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs := parser.parseRequirements(tt.line, 1)

			if len(reqs) != len(tt.expectReqIDs) {
				t.Errorf("expected %d requirements, got %d", len(tt.expectReqIDs), len(reqs))
				return
			}

			for i, req := range reqs {
				if req.ID != tt.expectReqIDs[i] {
					t.Errorf("expected requirement ID %s, got %s", tt.expectReqIDs[i], req.ID)
				}
				if req.IsAnnotated != tt.expectAnnotated[i] {
					t.Errorf("expected IsAnnotated=%v for %s, got %v", tt.expectAnnotated[i], req.ID, req.IsAnnotated)
				}
			}
		})
	}
}
