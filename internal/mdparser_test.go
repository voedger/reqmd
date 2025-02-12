package internal

import (
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMdParser_ParseMarkdownFile(t *testing.T) {
	testDataDir := filepath.Join("testdata", "mdparser-1.md")

	basicFile, _, err := ParseMarkdownFile(newMdCtx(), testDataDir)
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

func TestMdParser_ParseMarkdownFile_Errors(t *testing.T) {
	testDataDir := filepath.Join("testdata", "mdparser-errs.md")

	_, errors, err := ParseMarkdownFile(newMdCtx(), testDataDir)
	require.NoError(t, err)

	// We expect 3 errors in the test file:
	// 1. Invalid package name (non-identifier)
	// 2. Invalid requirement name (non-identifier)
	// 3. Mismatched requirement site IDs
	require.Len(t, errors, 4, "expected exactly 3 syntax errors")

	// Sort errors by line number for consistent testing
	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Line < errors[j].Line
	})

	// Check package ID error
	assert.Equal(t, "pkgident", errors[0].Code)
	assert.Equal(t, 2, errors[0].Line)

	// Check requirement name error
	assert.Equal(t, "reqident", errors[1].Code)
	assert.Equal(t, 8, errors[1].Line)

	// Check requirement site ID mismatch error
	assert.Equal(t, "reqsiteid", errors[2].Code)
	assert.Equal(t, 10, errors[2].Line)

	// Check coverage status error
	assert.Equal(t, "covstatus", errors[3].Code)
	assert.Equal(t, 12, errors[3].Line)
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

func newMdCtx() *MarkdownContext {
	return &MarkdownContext{}
}

func TestParseRequirements_invalid_coverage_status(t *testing.T) {
	var errors []SyntaxError
	res := ParseRequirements("test.md", "`~Post.handler~`covrd[^~Post.handler~]", 1, &errors)
	require.Len(t, res, 1)
}

func TestParseRequirements_table(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []RequirementSite
	}{
		{
			name:  "Bare requirement site",
			input: "`~Post.handler~`",
			expected: []RequirementSite{{
				RequirementName: "Post.handler",
				Line:            1,
				IsAnnotated:     false,
			}},
		},
		{
			name:  "Annotated requirement site with coverage status",
			input: "`~Post.handler~`covered[^~Post.handler~]✅",
			expected: []RequirementSite{{
				RequirementName:     "Post.handler",
				ReferenceName:       "Post.handler",
				CoverageStatusWord:  "covered",
				CoverageStatusEmoji: "✅",
				Line:                1,
				IsAnnotated:         true,
			}},
		},
		{
			name:  "Annotated requirement site with coverage status, no emoji",
			input: "`~Post.handler~`covered[^~Post.handler~]",
			expected: []RequirementSite{{
				RequirementName:     "Post.handler",
				ReferenceName:       "Post.handler",
				CoverageStatusWord:  "covered",
				CoverageStatusEmoji: "",
				Line:                1,
				IsAnnotated:         true,
			}},
		},
		{
			name:  "Uncovered requirement site with status",
			input: "`~Post.handler~`uncvrd[^~Post.handler~]❓",
			expected: []RequirementSite{{
				RequirementName:     "Post.handler",
				ReferenceName:       "Post.handler",
				CoverageStatusWord:  "uncvrd",
				CoverageStatusEmoji: "❓",
				Line:                1,
				IsAnnotated:         true,
			}},
		},
		{
			name:  "Multiple requirements in one line",
			input: "`~REQ001~` and `~REQ002~`covered[^~REQ002~]✅",
			expected: []RequirementSite{
				{
					RequirementName: "REQ001",
					Line:            1,
					IsAnnotated:     false,
				},
				{
					RequirementName:     "REQ002",
					ReferenceName:       "REQ002",
					CoverageStatusWord:  "covered",
					CoverageStatusEmoji: "✅",
					Line:                1,
					IsAnnotated:         true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errors []SyntaxError
			got := ParseRequirements("test.md", tt.input, 1, &errors)

			require.Equal(t, len(tt.expected), len(got), "number of requirements mismatch")
			for i, exp := range tt.expected {
				// Set common fields that we don't need to specify in every test case
				exp.FilePath = "test.md"
				assert.Equal(t, exp, got[i], "requirement %d mismatch", i)
			}
			assert.Empty(t, errors, "unexpected errors")
		})
	}
}

func Test_ParseCoverageFootnote(t *testing.T) {
	line := "[^~REQ002~]: `[~com.example.basic/REQ002~impl]`[folder1/filename1:line1:impl](https://example.com/pkg1/filename1#L11), [folder2/filename2:line2:test](https://example.com/pkg2/filename2#L22)"
	ctx := &MarkdownContext{
		rfiles: map[string]string{
			"https://example.com/pkg1/filename1": "hash1",
			"https://example.com/pkg2/filename2": "hash2",
		},
	}
	note := ParseCoverageFootnote(ctx, "", line, 1, nil)
	require.NotNil(t, note)

	assert.Equal(t, "REQ002", note.RequirementID, "incorrect requirement ID in footnote")
	assert.Equal(t, "com.example.basic", note.PackageID, "incorrect package ID in footnote")

	require.Len(t, note.Coverers, 2, "should have 2 coverage references")
	assert.Equal(t, "folder1/filename1:line1:impl", note.Coverers[0].CoverageLabel)
	assert.Equal(t, "https://example.com/pkg1/filename1#L11", note.Coverers[0].CoverageURL)
	assert.Equal(t, "hash1", note.Coverers[0].FileHash)
	assert.Equal(t, "folder2/filename2:line2:test", note.Coverers[1].CoverageLabel)
	assert.Equal(t, "https://example.com/pkg2/filename2#L22", note.Coverers[1].CoverageURL)
	assert.Equal(t, "hash2", note.Coverers[1].FileHash)
}

func Test_ParseCoverageFootnote_errors(t *testing.T) {
	line := "[^~REQ002~]: `[~com.example.basic/REQ002~impl]`[folder2/filename2:line2:test](://example.com/path)"

	var errors []SyntaxError
	note := ParseCoverageFootnote(newMdCtx(), "", line, 1, &errors)
	require.NotNil(t, note)

	require.Len(t, errors, 1, "should have 1 error")
	assert.Equal(t, "urlsyntax", errors[0].Code)
	assert.Contains(t, errors[0].Message, "://example.com/path")
}

// Test that golang can recognize patterns like "`~Post.handler~`covered[^~Post.handler~]✅"
func TestRegexpEmojis(t *testing.T) {
	text1 := "Some text1 `~Post.handler~`covered[^~Post.handler~]✅ Some text2"
	text2 := "Some text1 `~Post.handler~`covered[^~Post.handler~] Some text2"

	pattern := regexp.MustCompile("`.*✅")

	if !pattern.MatchString(text1) {
		t.Errorf("Pattern did not match, but was expected to match.\nText: %s", text1)
	}
	if pattern.MatchString(text2) {
		t.Errorf("Pattern matches, but was not expected to match.\nText: %s", text2)
	}
}
