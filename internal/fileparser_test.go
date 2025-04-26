// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileParser_md(t *testing.T) {
	testDataFile := filepath.Join("testdata", "mdparser-1.md")

	basicFile, _, err := parseFile(newMdCtx(), testDataFile)
	require.NoError(t, err)

	// Test package Id and requirements count
	assert.Equal(t, "com.example.basic", string(basicFile.PackageId), "incorrect package id")
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
			assert.Equal(t, CoverageFootnoteId("~REQ002~"), footnote.CoverageFootnoteId, "incorrect CoverageFootnoteId in footnote")
			assert.Equal(t, "com.example.basic", string(footnote.PackageId), "incorrect package Id in footnote")

			require.Len(t, footnote.Coverers, 2, "should have 2 coverage references")
			assert.Equal(t, "folder1/filename1:line1:impl", footnote.Coverers[0].CoverageLabel)
			assert.Equal(t, "https://example.com/pkg1/filename1#L11", footnote.Coverers[0].CoverageURL)
			assert.Equal(t, "folder2/filename2:line2:test", footnote.Coverers[1].CoverageLabel)
			assert.Equal(t, "https://example.com/pkg2/filename2#L22", footnote.Coverers[1].CoverageURL)
		}
	}
}

func TestFileParser_md_error_pkgident(t *testing.T) {
	testData := filepath.Join("testdata", "systest", "errors_syn", "err_pkgident.md")
	_, errors, err := parseFile(newMdCtx(), testData)
	require.NoError(t, err)
	require.Len(t, errors, 1, "expected exactly 1 syntax error")
}

func TestFileParser_md_IgnorePackage(t *testing.T) {
	// Create a temporary file for testing
	content := []byte(`---
reqmd.package: ignoreme
---
This content should be ignored. ` + "`~REQ001~`" + `
`)
	tmpfile := filepath.Join(t.TempDir(), "test.md")
	require.NoError(t, os.WriteFile(tmpfile, content, 0644))

	// Parse the file
	structure, errs, err := parseFile(newMdCtx(), tmpfile)
	require.NoError(t, err)
	assert.Empty(t, errs)
	assert.NotNil(t, structure)
	assert.Equal(t, "ignoreme", string(structure.PackageId))
	assert.Empty(t, structure.Requirements)
	assert.Empty(t, structure.CoverageFootnotes)
}

func TestFileParser_src(t *testing.T) {
	// Test data file contains a line with: [~server.api.v2/Post.handler~test]
	testDataFile := filepath.Join("testdata", "srccoverparser-1.go")
	srcFile, syntaxErrors, err := parseFile(newMdCtx(), testDataFile)
	require.NoError(t, err)
	assert.Len(t, syntaxErrors, 0)

	// Verify file type and that at least one coverage tag is found
	assert.Equal(t, FileTypeSource, srcFile.Type)
	require.NotEmpty(t, srcFile.CoverageTags)

	{
		tag := srcFile.CoverageTags[0]
		assert.Equal(t, StrToReqId("server.api.v2/Post.handler"), tag.RequirementId)
		assert.Equal(t, "impl", tag.CoverageType)
		// Adjust expected line number according to your test file content
		assert.Equal(t, 11, tag.Line)
	}

	{
		tag := srcFile.CoverageTags[1]
		assert.Equal(t, StrToReqId("server.api.v2/Post.handler"), tag.RequirementId)
		assert.Equal(t, "test", tag.CoverageType)
		// Adjust expected line number according to your test file content
		assert.Equal(t, 17, tag.Line)
	}
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
	res := parseRequirements("test.md", "`~Post.handler~`coovrd[^~Post.handler~]", 1, &errors)
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
				CoverageFootnoteId:  "~Post.handler~",
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
				CoverageFootnoteId:  "~Post.handler~",
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
				CoverageFootnoteId:  "~Post.handler~",
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
			got := parseRequirements("test.md", tt.input, 1, &errors)

			require.Equal(t, len(tt.expected), len(got), "number of requirements mismatch: %d: %s: %s", tidx, tt.name, tt.input)
			for i, exp := range tt.expected {
				// Set common fields that we don't need to specify in every test case
				assert.Equal(t, exp, got[i], "requirement %d.%d mismatch, %s: %s", tidx, i, tt.name, tt.input)
			}
			assert.Empty(t, errors, "unexpected errors")
		})
	}
}

func TestParseCoverageFootnote(t *testing.T) {
	line := "[^~REQ002~]: `[~com.example.basic/REQ002~impl]`[folder1/filename1:line1:impl](https://example.com/pkg1/filename1#L11), [folder2/filename2:line2:test](https://example.com/pkg2/filename2#L22)"
	ctx := &MarkdownContext{}
	note := ParseCoverageFootnote(ctx, "", line, 1, nil)
	require.NotNil(t, note)

	assert.Equal(t, CoverageFootnoteId("~REQ002~"), note.CoverageFootnoteId, "incorrect CoverageFootnoteId in footnote")
	assert.Equal(t, "com.example.basic", string(note.PackageId), "incorrect package Id in footnote")

	require.Len(t, note.Coverers, 2, "should have 2 coverage references")
	assert.Equal(t, "folder1/filename1:line1:impl", note.Coverers[0].CoverageLabel)
	assert.Equal(t, "https://example.com/pkg1/filename1#L11", note.Coverers[0].CoverageURL)
	assert.Equal(t, "folder2/filename2:line2:test", note.Coverers[1].CoverageLabel)
	assert.Equal(t, "https://example.com/pkg2/filename2#L22", note.Coverers[1].CoverageURL)
}

func TestParseCoverageFootnote2(t *testing.T) {
	line := "[^~VVMLeader.def~]: `[~server.design.orch/VVMLeader.def~]` [apps/app.go:80:impl](https://example.com/pkg1/filename1#L80)"
	ctx := &MarkdownContext{}
	note := ParseCoverageFootnote(ctx, "", line, 1, nil)
	require.NotNil(t, note)

	assert.Equal(t, CoverageFootnoteId("~VVMLeader.def~"), note.CoverageFootnoteId, "incorrect CoverageFootnoteId in footnote")
	assert.Equal(t, "server.design.orch", string(note.PackageId), "incorrect package Id in footnote")

	require.Len(t, note.Coverers, 1, "should have 1 coverer")
	assert.Equal(t, "apps/app.go:80:impl", note.Coverers[0].CoverageLabel)
}

func TestParseCoverageFootnote_JustFootnote(t *testing.T) {
	line := "[^12]:"
	ctx := &MarkdownContext{}
	note := ParseCoverageFootnote(ctx, "", line, 1, nil)
	require.NotNil(t, note)
	assert.Equal(t, CoverageFootnoteId("12"), note.CoverageFootnoteId)
	assert.Equal(t, "", string(note.PackageId))

}

func TestParseCoverageFootnote_errors(t *testing.T) {
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
				Message:  "only one RequirementSite is allowed per line: `~REQ001~`,  `~REQ002~`covered[^~REQ002~]✅",
			},
		},
		{
			name: "invalid coverage status",
			line: "`~REQ001~`coovrd[^~REQ001~]✅",
			wantErr: ProcessingError{
				Code:     "covstatus",
				FilePath: "test.md",
				Line:     3,
				Message:  "CoverageStatusWord shall be 'covered' or 'uncvrd': coovrd",
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
			requirements := parseRequirements("test.md", line, tt.wantErr.Line, &errors)

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
