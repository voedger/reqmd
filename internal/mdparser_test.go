package internal

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMdParser_Scan(t *testing.T) {

	testDataDir := filepath.Join("testdata", "mdparser-1.md")

	basicFile, _, err := ParseMarkdownFile(testDataDir)
	require.NoError(t, err)

	// Test package ID
	if basicFile.PackageID != "com.example.basic" {
		t.Errorf("expected package ID 'com.example.basic', got '%s'", basicFile.PackageID)
	}

	// Test requirements
	if len(basicFile.Requirements) != 2 {
		t.Errorf("expected 2 requirements, got %d", len(basicFile.Requirements))
	}

	// Verify REQ001
	{
		found := false
		for _, req := range basicFile.Requirements {
			if req.RequirementName == "REQ001" {
				found = true
				if req.IsAnnotated {
					t.Error("REQ001 should not be annotated")
				}
				if req.Line != 7 {
					t.Errorf("REQ001 should be on line 7, got %d", req.Line)
				}
			}
		}
		if !found {
			t.Error("REQ001 not found")
		}
	}

	// Verify REQ002 (appears twice, one annotated)
	{
		cnt := 0
		for _, req := range basicFile.Requirements {
			if req.RequirementName == "REQ002" {
				cnt++
				if cnt > 1 {
					t.Error("REQ002 should only appear once")
				}
				if !req.IsAnnotated {
					t.Error("REQ002 should be annotated")
				}
			}
		}
		if cnt == 0 {
			t.Error("REQ002 not found")
		}
	}

}

func TestMdParser_parseRequirements(t *testing.T) {

	tests := []struct {
		name            string
		line            string
		expectReqIDs    []string
		expectAnnotated []bool
	}{
		{
			name:            "simple requirement",
			line:            "This contains `~REQ001~`",
			expectReqIDs:    []string{"REQ001"},
			expectAnnotated: []bool{false},
		},
		{
			name:            "multiple requirements",
			line:            "Contains `~REQ001~` and `~REQ002~`",
			expectReqIDs:    []string{"REQ001", "REQ002"},
			expectAnnotated: []bool{false, false},
		},
		{
			name:            "annotated requirement",
			line:            "`~REQ001~`cov[^~REQ001~]",
			expectReqIDs:    []string{"REQ001"},
			expectAnnotated: []bool{true},
		},
		{
			name:            "mixed requirements",
			line:            "`~REQ001~` normal and `~REQ002~`cov[^~REQ002~] annotated",
			expectReqIDs:    []string{"REQ001", "REQ002"},
			expectAnnotated: []bool{false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs := parseRequirements(tt.line, 1)

			if len(reqs) != len(tt.expectReqIDs) {
				t.Errorf("expected %d requirements, got %d (%s)", len(tt.expectReqIDs), len(reqs), tt.line)
				return
			}

			for i, req := range reqs {
				if req.RequirementName != tt.expectReqIDs[i] {
					t.Errorf("expected requirement ID %s, got %s (%s)", tt.expectReqIDs[i], req.RequirementName, tt.line)
				}
				if req.IsAnnotated != tt.expectAnnotated[i] {
					t.Errorf("expected IsAnnotated=%v for %s, got %v (%s)", tt.expectAnnotated[i], req.RequirementName, req.IsAnnotated, tt.line)
				}
			}
		})
	}
}

// requirementSiteRegex matches:
// - A backtick (`)
// - A RequirementSiteID: a tilde (~), an Identifier, then a tilde (~)
// - A backtick (`)
// - Optionally, the literal "cov" followed by a CoverageFootnoteReference:
//   - "[^"
//   - the same RequirementSiteID pattern
//   - "]"

func TestRequirementSiteRegex(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectMatch bool
		group1      string // The captured RequirementSiteID (e.g., "~Post.handler~")
		group2      string // The optional captured CoverageFootnoteReference (if any)
	}{
		{
			name:        "Bare requirement site",
			input:       "`~Post.handler~`",
			expectMatch: true,
			group1:      "Post.handler",
			group2:      "",
		},
		{
			name:        "Annotated requirement site",
			input:       "`~Post.handler~`cov[^~Post.handler~]",
			expectMatch: true,
			group1:      "Post.handler",
			group2:      "Post.handler",
		},
		{
			name:        "Annotated with different requirement id",
			input:       "`~Post.handler~`cov[^~Other.handler~]",
			expectMatch: true,
			group1:      "Post.handler",
			group2:      "Other.handler",
		},
		{
			name:        "Missing closing backtick",
			input:       "`~Post.handler~",
			expectMatch: false,
		},
		{
			name:        "Invalid identifier (starts with digit)",
			input:       "`~123InvalidIdentifier~`",
			group1:      "123InvalidIdentifier",
			expectMatch: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matches := RequirementSiteRegex.FindStringSubmatch(tc.input)
			if tc.expectMatch {
				if matches == nil {
					t.Fatalf("Expected match for input %q but got none", tc.input)
				}
				// matches[0] is the entire match, matches[1] is the RequirementSiteID,
				// and matches[2] is the optional coverage footnote requirementSiteID.
				if matches[1] != tc.group1 {
					t.Errorf("Expected group1 to be %q, got %q", tc.group1, matches[1])
				}
				// Check group2 only if it was expected.
				if len(matches) > 2 {
					if matches[2] != tc.group2 {
						t.Errorf("Expected group2 to be %q, got %q", tc.group2, matches[2])
					}
				} else if tc.group2 != "" {
					t.Errorf("Expected group2 to be %q but no group2 captured", tc.group2)
				}
			} else {
				if matches != nil {
					t.Errorf("Expected no match for input %q, but got %v", tc.input, matches)
				}
			}
		})
	}
}
