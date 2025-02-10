package internal

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMdParser_ParseMarkdownFile(t *testing.T) {
	testDataDir := filepath.Join("testdata", "mdparser-1.md")

	basicFile, _, err := ParseMarkdownFile(testDataDir)
	require.NoError(t, err)

	// Test package ID and requirements count
	assert.Equal(t, "com.example.basic", basicFile.PackageID, "incorrect package ID")
	assert.Len(t, basicFile.Requirements, 2, "incorrect number of requirements")

	// Find and verify REQ001
	req001 := findRequirement(basicFile.Requirements, "REQ001")
	assert.NotNil(t, req001, "REQ001 not found")
	if req001 != nil {
		assert.False(t, req001.IsAnnotated, "REQ001 should not be annotated")
		assert.Equal(t, 7, req001.Line, "REQ001 is on wrong line")
	}

	// Find and verify REQ002
	req002 := findRequirement(basicFile.Requirements, "REQ002")
	assert.NotNil(t, req002, "REQ002 not found")
	if req002 != nil {
		assert.True(t, req002.IsAnnotated, "REQ002 should be annotated")
	}

	// Test coverage footnote
	// [^~REQ002~]: `[~com.example.basic/REQ002~impl]` [folder1/filename1:line1:impl](https://example.com/pkg1/filename1), [folder2/filename2:line2:test](https://example.com/pkg2/filename2)
	{
		assert.Len(t, basicFile.CoverageFootnotes, 1, "should have 1 coverage footnote")
		if len(basicFile.CoverageFootnotes) > 0 {
			footnote := basicFile.CoverageFootnotes[0]
			assert.Equal(t, "REQ002", footnote.RequirementID, "incorrect requirement ID in footnote")
			assert.Equal(t, "com.example.basic", footnote.PackageID, "incorrect package ID in footnote")

			require.Len(t, footnote.Coverers, 2, "should have 2 coverage references")
			assert.Equal(t, "folder1/filename1:line1:impl", footnote.Coverers[0].CoverageLabel)
			assert.Equal(t, "https://example.com/pkg1/filename1#L11", footnote.Coverers[0].CoverageURL)
			assert.Equal(t, "folder2/filename2:line2:test", footnote.Coverers[1].CoverageLabel)
			assert.Equal(t, "https://example.com/pkg2/filename2#L22", footnote.Coverers[1].CoverageURL)
		}
	}
}

func TestMdParser_ParseMarkdownFileErr_reqsiteid(t *testing.T) {
	testDataDir := filepath.Join("testdata", "mdparser-errs.md")

	_, errors, err := ParseMarkdownFile(testDataDir)
	require.NoError(t, err)
	require.Len(t, errors, 1, "should have 1 error")
	assert.Equal(t, "reqsiteid", errors[0].Code)
	assert.Equal(t, 8, errors[0].Line)
}

// Helper function to find requirement by name
func findRequirement(reqs []RequirementSite, name string) *RequirementSite {
	for _, req := range reqs {
		if req.RequirementName == name {
			return &req
		}
	}
	return nil
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

	errors := []SyntaxError{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs := ParseRequirements("", tt.line, 1, &errors)

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

func TestMdParser_ParseCoverageFootnote(t *testing.T) {
	line := "[^~REQ002~]: `[~com.example.basic/REQ002~impl]`[folder1/filename1:line1:impl](https://example.com/pkg1/filename1), [folder2/filename2:line2:test](https://example.com/pkg2/filename2)"
	note := ParseCoverageFootnote("", line, 1, nil)
	require.NotNil(t, note)
}
