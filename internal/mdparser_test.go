// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMdParser_ParseMarkdownFile(t *testing.T) {
	testDataFile := filepath.Join("testdata", "mdparser-1.md")

	basicFile, _, err := ParseMarkdownFile(newMdCtx(), testDataFile)
	require.NoError(t, err)

	// Test package ID and requirements count
	assert.Equal(t, "com.example.basic", basicFile.PackageID, "incorrect package ID")
	assert.Len(t, basicFile.Requirements, 2, "incorrect number of requirements")

	// Find and verify REQ001
	req001 := findRequirement(basicFile.Requirements, "REQ001")
	assert.NotNil(t, req001, "REQ001 not found")
	if req001 != nil {
		assert.False(t, req001.HasAnnotationRef, "REQ001 should not be annotated")
		assert.Equal(t, 7, req001.Line, "REQ001 is on wrong line")
	}

	// Find and verify REQ002
	req002 := findRequirement(basicFile.Requirements, "REQ002")
	assert.NotNil(t, req002, "REQ002 not found")
	if req002 != nil {
		assert.True(t, req002.HasAnnotationRef, "REQ002 should be annotated")
	}

	// Test coverage footnote
	// [^~REQ002~]: `[~com.example.basic/REQ002~impl]` [folder1/filename1:line1:impl](https://example.com/pkg1/filename1), [folder2/filename2:line2:test](https://example.com/pkg2/filename2)
	{
		assert.Len(t, basicFile.CoverageFootnotes, 1, "should have 1 coverage footnote")
		if len(basicFile.CoverageFootnotes) > 0 {
			footnote := basicFile.CoverageFootnotes[0]
			assert.Equal(t, CoverageFootnoteId("~REQ002~"), footnote.CoverageFootnoteId, "incorrect CoverageFootnoteID in footnote")
			assert.Equal(t, "com.example.basic", footnote.PackageID, "incorrect package ID in footnote")

			require.Len(t, footnote.Coverers, 2, "should have 2 coverage references")
			assert.Equal(t, "folder1/filename1:line1:impl", footnote.Coverers[0].CoverageLabel)
			assert.Equal(t, "https://example.com/pkg1/filename1#L11", footnote.Coverers[0].CoverageUrL)
			assert.Equal(t, "folder2/filename2:line2:test", footnote.Coverers[1].CoverageLabel)
			assert.Equal(t, "https://example.com/pkg2/filename2#L22", footnote.Coverers[1].CoverageUrL)
		}
	}
}

func TestMdParser_ParseMarkdownFile_Errors(t *testing.T) {
	testDataDir := filepath.Join("testdata", "mdparser-errs.md")

	_, errors, err := ParseMarkdownFile(newMdCtx(), testDataDir)
	require.NoError(t, err)

	// We expect 4 errors in the test file:
	// 1. Invalid package name (non-identifier)
	// 2. Invalid requirement name (non-identifier)
	// 3. Invalid coverage status
	// 4. Unmatched fence
	require.Len(t, errors, 4, "expected exactly 4 syntax errors")

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

	// Check coverage status error
	assert.Equal(t, "covstatus", errors[2].Code)
	assert.Equal(t, 10, errors[2].Line)

	// Check unmatched fence error
	assert.Equal(t, "unmatchedfence", errors[3].Code)
	assert.Equal(t, 14, errors[3].Line)
}

// Helper function to find requirement by name
func findRequirement(reqs []RequirementSite, name RequirementName) *RequirementSite {
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
	var errors []ProcessingError
	res := ParseRequirements("test.md", "`~Post.handler~`covrd[^~Post.handler~]", 1, &errors)
	require.Len(t, res, 0)
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
				RequirementName:  "Post.handler",
				Line:             1,
				HasAnnotationRef: false,
			}},
		},
		{
			name:  "Annotated requirement site with coverage status",
			input: "`~Post.handler~`covered[^~Post.handler~]✅",
			expected: []RequirementSite{{
				RequirementName:     "Post.handler",
				CoverageFootnoteID:  "~Post.handler~",
				CoverageStatusWord:  "covered",
				CoverageStatusEmoji: "✅",
				Line:                1,
				HasAnnotationRef:    true,
			}},
		},
		{
			name:  "Annotated requirement site with coverage status, no emoji",
			input: "`~Post.handler~`covered[^~Post.handler~]",
			expected: []RequirementSite{{
				RequirementName:     "Post.handler",
				CoverageFootnoteID:  "~Post.handler~",
				CoverageStatusWord:  "covered",
				CoverageStatusEmoji: "",
				Line:                1,
				HasAnnotationRef:    true,
			}},
		},
		{
			name:  "Uncovered requirement site with status",
			input: "`~Post.handler~`uncvrd[^~Post.handler~]❓",
			expected: []RequirementSite{{
				RequirementName:     "Post.handler",
				CoverageFootnoteID:  "~Post.handler~",
				CoverageStatusWord:  "uncvrd",
				CoverageStatusEmoji: "❓",
				Line:                1,
				HasAnnotationRef:    true,
			}},
		},
	}

	for tidx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errors []ProcessingError
			got := ParseRequirements("test.md", tt.input, 1, &errors)

			require.Equal(t, len(tt.expected), len(got), "number of requirements mismatch: %d: %s: %s", tidx, tt.name, tt.input)
			for i, exp := range tt.expected {
				// Set common fields that we don't need to specify in every test case
				exp.FilePath = "test.md"
				assert.Equal(t, exp, got[i], "requirement %d.%d mismatch, %s: %s", tidx, i, tt.name, tt.input)
			}
			assert.Empty(t, errors, "unexpected errors")
		})
	}
}

func Test_ParseCoverageFootnote(t *testing.T) {
	line := "[^~REQ002~]: `[~com.example.basic/REQ002~impl]`[folder1/filename1:line1:impl](https://example.com/pkg1/filename1#L11), [folder2/filename2:line2:test](https://example.com/pkg2/filename2#L22)"
	ctx := &MarkdownContext{
		rfiles: &Reqmdjson{
			FileUrl2FileHash: map[string]string{
				"https://example.com/pkg1/filename1": "hash1",
				"https://example.com/pkg2/filename2": "hash2",
			},
		},
	}
	note := ParseCoverageFootnote(ctx, "", line, 1, nil)
	require.NotNil(t, note)

	assert.Equal(t, CoverageFootnoteId("~REQ002~"), note.CoverageFootnoteId, "incorrect CoverageFootnoteID in footnote")
	assert.Equal(t, "com.example.basic", note.PackageID, "incorrect package ID in footnote")

	require.Len(t, note.Coverers, 2, "should have 2 coverage references")
	assert.Equal(t, "folder1/filename1:line1:impl", note.Coverers[0].CoverageLabel)
	assert.Equal(t, "https://example.com/pkg1/filename1#L11", note.Coverers[0].CoverageUrL)
	assert.Equal(t, "hash1", note.Coverers[0].FileHash)
	assert.Equal(t, "folder2/filename2:line2:test", note.Coverers[1].CoverageLabel)
	assert.Equal(t, "https://example.com/pkg2/filename2#L22", note.Coverers[1].CoverageUrL)
	assert.Equal(t, "hash2", note.Coverers[1].FileHash)
}

func Test_ParseCoverageFootnote2(t *testing.T) {
	line := "[^~VVMLeader.def~]: `[~server.design.orch/VVMLeader.def~]` [apps/app.go:80:impl](https://example.com/pkg1/filename1#L80)"
	ctx := &MarkdownContext{
		rfiles: &Reqmdjson{
			FileUrl2FileHash: map[string]string{
				"https://example.com/pkg1/filename1": "hash1",
			},
		},
	}
	note := ParseCoverageFootnote(ctx, "", line, 1, nil)
	require.NotNil(t, note)

	assert.Equal(t, CoverageFootnoteId("~VVMLeader.def~"), note.CoverageFootnoteId, "incorrect CoverageFootnoteID in footnote")
	assert.Equal(t, "server.design.orch", note.PackageID, "incorrect package ID in footnote")

	require.Len(t, note.Coverers, 1, "should have 1 coverer")
	assert.Equal(t, "apps/app.go:80:impl", note.Coverers[0].CoverageLabel)
}

func Test_ParseCoverageFootnote_JustFootnote(t *testing.T) {
	line := "[^12]:"
	ctx := &MarkdownContext{
		rfiles: &Reqmdjson{
			FileUrl2FileHash: map[string]string{
				"https://example.com/pkg1/filename1": "hash1",
			},
		},
	}
	note := ParseCoverageFootnote(ctx, "", line, 1, nil)
	require.NotNil(t, note)
	assert.Equal(t, CoverageFootnoteId("12"), note.CoverageFootnoteId)
	assert.Equal(t, "", note.PackageID)

}

func Test_ParseCoverageFootnote_errors(t *testing.T) {
	line := "[^~REQ002~]: `[~com.example.basic/REQ002~impl]`[folder2/filename2:line2:test](://example.com/path)"

	var errors []ProcessingError
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

func TestParseRequirements_errors(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr ProcessingError
	}{
		{
			name: "multiple requirements on single line",
			line: "`~REQ001~` `~REQ002~`covered[^~REQ002~]✅",
			wantErr: ProcessingError{
				Code:     "multisites",
				FilePath: "test.md",
				Line:     1,
				Message:  "Only one RequirementSite is allowed per line: `~REQ001~`,  `~REQ002~`covered[^~REQ002~]✅",
			},
		},
		{
			name: "invalid coverage status",
			line: "`~REQ001~`covrd[^~REQ001~]✅",
			wantErr: ProcessingError{
				Code:     "covstatus",
				FilePath: "test.md",
				Line:     3,
				Message:  "CoverageStatusWord shall be 'covered' or 'uncvrd': covrd",
			},
		},
		{
			name: "empty coverage status",
			line: "`~REQ001~`[^~REQ001~]✅",
			wantErr: ProcessingError{
				Code:     "covstatus",
				FilePath: "test.md",
				Line:     4,
				Message:  "CoverageStatusWord shall be 'covered' or 'uncvrd': ",
			},
		},
	}

	for tidx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errors []ProcessingError
			line := tt.line
			requirements := ParseRequirements("test.md", line, tt.wantErr.Line, &errors)

			// Check that no requirements were returned
			assert.Nil(t, requirements, "%d: should return nil requirements", tidx)

			// Verify the error
			if assert.Len(t, errors, 1, "%d: should have exactly one error", tidx) {
				assert.Equal(t, tt.wantErr.Code, errors[0].Code)
				assert.Equal(t, tt.wantErr.Message, errors[0].Message)
				assert.Equal(t, tt.wantErr.FilePath, errors[0].FilePath)
				assert.Equal(t, tt.wantErr.Line, errors[0].Line)
			}
		})
	}
}

func TestMdParser_ParseMarkdownFile_IgnorePackage(t *testing.T) {
	// Create a temporary file for testing
	content := []byte(`---
reqmd.package: ignoreme
---
This content should be ignored. ` + "`~REQ001~`" + `
`)
	tmpfile := filepath.Join(t.TempDir(), "test.md")
	require.NoError(t, os.WriteFile(tmpfile, content, 0644))

	// Parse the file
	structure, errs, err := ParseMarkdownFile(newMdCtx(), tmpfile)
	require.NoError(t, err)
	assert.Empty(t, errs)
	assert.NotNil(t, structure)
	assert.Equal(t, "ignoreme", structure.PackageID)
	assert.Empty(t, structure.Requirements)
	assert.Empty(t, structure.CoverageFootnotes)
}
