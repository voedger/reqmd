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
		assert.NotContains(t, err.Message, "file1.md:10")
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

// Non-annotated requirement
func TestAnalyzer_ActionFootnote_Nan(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with one requirement
	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ001", CoverageStatusWordUncvrd)
	mdFile.Requirements[0].HasAnnotationRef = false

	result, err := analyzer.Analyze([]FileStructure{mdFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate both a footnote and status update action
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify status update action
	assert.Equal(t, ActionSite, actions[0].Type)
	assert.Equal(t, "REQ001", actions[0].RequirementName)
	assert.Equal(t, 10, actions[0].Line)
	assert.Equal(t, FormatRequirementSite("REQ001", CoverageStatusWordUncvrd, 1), actions[0].Data)

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[1].Type)
	assert.Equal(t, "REQ001", actions[1].RequirementName)
	assert.Contains(t, actions[1].Data, "[^cf1]", "footnote should use cf1 notation")
}

// Non-annotated requirement with new coverer
func TestAnalyzer_ActionFootnote_Nan_NewCoverer(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with one requirement
	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ001", CoverageStatusWordUncvrd)
	mdFile.Requirements[0].HasAnnotationRef = false

	// Create a source file with coverage for that requirement
	srcFile := createSourceFileStructure(
		"src/impl.go",
		"https://github.com/org/repo/blob/main/src/impl.go",
		[]CoverageTag{
			createCoverageTag("pkg1/REQ001", "impl1", 20),
		},
	)

	result, err := analyzer.Analyze([]FileStructure{mdFile, srcFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate both a footnote and status update action
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify status update action
	assert.Equal(t, ActionSite, actions[0].Type)
	assert.Equal(t, "REQ001", actions[0].RequirementName)
	assert.Equal(t, 10, actions[0].Line)
	require.Contains(t, actions[0].Data, "`~REQ001~`covered[^cf1]✅", "site should use cf1 ID and include coverage status")

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[1].Type)
	assert.Equal(t, "REQ001", actions[1].RequirementName)
	assert.Contains(t, actions[1].Data, "[^cf1]:", "footnote should use cf1 ID")
	assert.Contains(t, actions[1].Data, "`[~pkg1/REQ001~impl]`", "footnote should have correct hint")
	assert.Contains(t, actions[1].Data, "[src/impl.go:20:impl1]", "footnote should have correct coverage")
	assert.Contains(t, actions[1].Data, "(https://github.com/org/repo/blob/main/src/impl.go/src/impl.go#L20)", "footnote should have correct URL")
}

// Annotated uncovered requirement with new coverer
func TestAnalyzer_ActionStatusUpdate_AnUncov_NewCoverer(t *testing.T) {
	analyzer := NewAnalyzer()

	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ002", CoverageStatusWordUncvrd)

	// Create source files with coverage
	srcFile := createSourceFileStructure(
		"src/impl.go",
		"https://github.com/org/repo/blob/main/src/impl.go",
		[]CoverageTag{
			createCoverageTag("pkg1/REQ002", "impl", 20),
		},
	)

	result, err := analyzer.Analyze([]FileStructure{mdFile, srcFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate both status update and footnote actions
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify site action
	assert.Equal(t, ActionSite, actions[0].Type)
	assert.Equal(t, "REQ002", actions[0].RequirementName)
	assert.Equal(t, 10, actions[0].Line)
	require.Contains(t, actions[0].Data, "`~REQ002~`covered[^cf1]✅", "site should use cf1 ID and correct status")

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[1].Type)
	assert.Equal(t, "REQ002", actions[1].RequirementName)
	assert.Contains(t, actions[1].Data, "[^cf1]:", "footnote should use cf1 ID")
	assert.Contains(t, actions[1].Data, "src/impl.go:20:impl", "footnote should have correct coverage")
	assert.Contains(t, actions[1].Data, "https://github.com/org/repo/blob/main/src/impl.go#L20", "footnote should have correct URL")
}

// Annotated covered requirement without coverers
func TestAnalyzer_ActionStatusUpdate_AnCov_NoCoverers(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a markdown file with covered requirement but no actual coverage
	mdFile := createMdStructureA("req.md", "pkg1", 11, "REQ001", CoverageStatusWordCovered)

	result, err := analyzer.Analyze([]FileStructure{mdFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	// Should generate status update action to uncovered
	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 2)

	// Verify site action
	assert.Equal(t, ActionSite, actions[0].Type)
	assert.Equal(t, "REQ001", actions[0].RequirementName)
	assert.Equal(t, 11, actions[0].Line)
	require.Contains(t, actions[0].Data, "`~REQ001~`uncvrd[^cf1]❓", "site should use cf1 ID and uncovered status")

	// Verify footnote action
	assert.Equal(t, ActionFootnote, actions[1].Type)
	assert.Equal(t, "REQ001", actions[1].RequirementName)
	assert.Contains(t, actions[1].Data, "[^cf1]:", "footnote should use cf1 ID")
	assert.Contains(t, actions[1].Data, "`[~pkg1/REQ001~impl]`", "footnote should have correct hint")
}

func TestAnalyzer_ActionFootnote_AnCov_NewHash(t *testing.T) {
	analyzer := NewAnalyzer()

	OldCoverageURL := "https://github.com/org/repo/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/old/file.go"
	OldFileHash := "oldhash"
	NewCoverageURL := "https://github.com/org/repo/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/new/file.go"
	NewFileHash := "newhash"

	// Create a markdown file with a requirement and existing footnote
	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ001", CoverageStatusWordCovered)
	mdFile.CoverageFootnotes = []CoverageFootnote{
		{
			RequirementName: "REQ001",
			Line:            20,
			ID:              1,
			Coverers: []Coverer{
				{
					CoverageLabel: "old/file.go:15:impl",
					CoverageUrL:   OldCoverageURL,
					FileHash:      OldFileHash,
				},
			},
		},
	}

	// Source file with the same Url but new hash
	srcFile := createSourceFileStructure(
		"src/impl.go",
		NewCoverageURL,
		[]CoverageTag{
			createCoverageTag("pkg1/REQ001", "impl", 20),
		},
	)
	srcFile.FileHash = NewFileHash

	result, err := analyzer.Analyze([]FileStructure{mdFile, srcFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 1)

	// Verify footnote update action uses existing line number
	assert.Equal(t, ActionFootnote, actions[0].Type)
	assert.Equal(t, 20, actions[0].Line)
	assert.Contains(t, actions[0].Data, "[^cf1]:", "footnote should use cf1 ID")
	assert.NotContains(t, actions[0].Data, OldCoverageURL)
	assert.Contains(t, actions[0].Data, NewCoverageURL)
}

func TestAnalyzer_ActionFootnote_AnCov_SameHash(t *testing.T) {
	analyzer := NewAnalyzer()

	OldCoverageURL := "https://github.com/org/repo/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/old/file.go"
	OldFileHash := "oldhash"
	NewCoverageURL := "https://github.com/org/repo/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/new/file.go"
	NewFileHash := OldFileHash

	// Create a markdown file with a requirement and existing footnote
	mdFile := createMdStructureA("req.md", "pkg1", 10, "REQ001", CoverageStatusWordCovered)
	mdFile.CoverageFootnotes = []CoverageFootnote{
		{
			RequirementName: "REQ001",
			Line:            20,
			ID:              1,
			Coverers: []Coverer{
				{
					CoverageLabel: "old/file.go:15:impl",
					CoverageUrL:   OldCoverageURL,
					FileHash:      OldFileHash,
				},
			},
		},
	}

	// Source file with the same URL but same hash
	srcFile := createSourceFileStructure(
		"src/impl.go",
		NewCoverageURL,
		[]CoverageTag{
			createCoverageTag("pkg1/REQ001", "impl", 20),
		},
	)
	srcFile.FileHash = NewFileHash

	result, err := analyzer.Analyze([]FileStructure{mdFile, srcFile})
	require.NoError(t, err)
	require.Empty(t, result.ProcessingErrors)

	actions := result.MdActions[mdFile.Path]
	require.Len(t, actions, 0)
}

// func TestAnalyzer_buildReqmdjsons(t *testing.T) {
// 	tests := []struct {
// 		name             string
// 		files            []FileStructure
// 		changedFootnotes map[RequirementID]bool
// 		want             map[FilePath]*Reqmdjson
// 	}{
// 		{
// 			name: "no changes, no jsons in result",
// 			files: []FileStructure{
// 				{
// 					Path:      "docs/req1.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg1",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ001"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ001",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "file.go:1:impl", CoverageUrL: "url1#L1", FileHash: "hash1"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			changedFootnotes: map[RequirementID]bool{},
// 			want:             map[FilePath]*Reqmdjson{},
// 		},
// 		{
// 			name: "changed footnotes, update json",
// 			files: []FileStructure{
// 				{
// 					Path:      "docs/req1.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg1",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ001"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ001",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "old.go:1:impl", CoverageUrL: "old_url#L1", FileHash: "old_hash"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			changedFootnotes: map[RequirementID]bool{
// 				"pkg1/REQ001": true,
// 			},
// 			want: map[FilePath]*Reqmdjson{
// 				"docs": {
// 					FileURL2FileHash: map[string]string{
// 						"new_url": "new_hash",
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "multiple files in same folder",
// 			files: []FileStructure{
// 				{
// 					Path:      "docs/req1.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg1",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ001"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ001",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "file1.go:1:impl", CoverageUrL: "url1#L1", FileHash: "hash1"},
// 							},
// 						},
// 					},
// 				},
// 				{
// 					Path:      "docs/req2.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg1",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ002"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ002",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "file2.go:1:impl", CoverageUrL: "url2#L1", FileHash: "hash2"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			changedFootnotes: map[RequirementID]bool{
// 				"pkg1/REQ001": true,
// 			},
// 			want: map[FilePath]*Reqmdjson{
// 				"docs": {
// 					FileURL2FileHash: map[string]string{
// 						"url1": "new_hash1",
// 						"url2": "hash2",
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "files in different folders",
// 			files: []FileStructure{
// 				{
// 					Path:      "docs/pkg1/req1.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg1",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ001"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ001",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "file1.go:1:impl", CoverageUrL: "url1#L1", FileHash: "hash1"},
// 							},
// 						},
// 					},
// 				},
// 				{
// 					Path:      "docs/pkg2/req2.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg2",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ001"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ001",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "file2.go:1:impl", CoverageUrL: "url2#L1", FileHash: "hash2"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			changedFootnotes: map[RequirementID]bool{
// 				"pkg1/REQ001": true,
// 			},
// 			want: map[FilePath]*Reqmdjson{
// 				"docs/pkg1": {
// 					FileURL2FileHash: map[string]string{
// 						"url1": "new_hash1",
// 					},
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Create analyzer and cast to concrete type
// 			a := NewAnalyzer().(*analyzer)

// 			// Initialize analyzer state
// 			result, err := a.Analyze(tt.files)
// 			require.NoError(t, err)

// 			// Set changed footnotes
// 			a.changedFootnotes = tt.changedFootnotes

// 			// Process coverages
// 			for requirementID, coverage := range a.coverages {
// 				if a.changedFootnotes[requirementID] {
// 					// Add new coverers
// 					coverage.NewCoverers = []*Coverer{
// 						{
// 							CoverageLabel: "new.go:1:impl",
// 							CoverageUrL:   "new_url#L1",
// 							FileHash:      "new_hash",
// 						},
// 					}
// 				}
// 			}

// 			// Clear result reqmdjsons before rebuilding
// 			result.Reqmdjsons = make(map[FilePath]*Reqmdjson)

// 			// Rebuild reqmdjsons
// 			a.buildReqmdjsons(result)

// 			// Verify result
// 			assert.Equal(t, tt.want, result.Reqmdjsons)
// 		})
// 	}
// }

// He
// func TestAnalyzer_buildReqmdjsons(t *testing.T) {
// 	tests := []struct {
// 		name             string
// 		files            []FileStructure
// 		changedFootnotes map[RequirementID]bool
// 		want             map[FilePath]*Reqmdjson
// 	}{
// 		{
// 			name: "no changes, no jsons in result",
// 			files: []FileStructure{
// 				{
// 					Path:      "docs/req1.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg1",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ001"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ001",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "file.go:1:impl", CoverageUrL: "url1#L1", FileHash: "hash1"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			changedFootnotes: map[RequirementID]bool{},
// 			want:             map[FilePath]*Reqmdjson{},
// 		},
// 		{
// 			name: "changed footnotes, update json",
// 			files: []FileStructure{
// 				{
// 					Path:      "docs/req1.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg1",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ001"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ001",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "old.go:1:impl", CoverageUrL: "old_url#L1", FileHash: "old_hash"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			changedFootnotes: map[RequirementID]bool{
// 				"pkg1/REQ001": true,
// 			},
// 			want: map[FilePath]*Reqmdjson{
// 				"docs": {
// 					FileURL2FileHash: map[string]string{
// 						"new_url": "new_hash",
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "multiple files in same folder",
// 			files: []FileStructure{
// 				{
// 					Path:      "docs/req1.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg1",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ001"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ001",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "file1.go:1:impl", CoverageUrL: "url1#L1", FileHash: "hash1"},
// 							},
// 						},
// 					},
// 				},
// 				{
// 					Path:      "docs/req2.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg1",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ002"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ002",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "file2.go:1:impl", CoverageUrL: "url2#L1", FileHash: "hash2"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			changedFootnotes: map[RequirementID]bool{
// 				"pkg1/REQ001": true,
// 			},
// 			want: map[FilePath]*Reqmdjson{
// 				"docs": {
// 					FileURL2FileHash: map[string]string{
// 						"url1": "new_hash1",
// 						"url2": "hash2",
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "files in different folders",
// 			files: []FileStructure{
// 				{
// 					Path:      "docs/pkg1/req1.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg1",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ001"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ001",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "file1.go:1:impl", CoverageUrL: "url1#L1", FileHash: "hash1"},
// 							},
// 						},
// 					},
// 				},
// 				{
// 					Path:      "docs/pkg2/req2.md",
// 					Type:      FileTypeMarkdown,
// 					PackageID: "pkg2",
// 					Requirements: []RequirementSite{
// 						{RequirementName: "REQ001"},
// 					},
// 					CoverageFootnotes: []CoverageFootnote{
// 						{
// 							RequirementName: "REQ001",
// 							Coverers: []Coverer{
// 								{CoverageLabel: "file2.go:1:impl", CoverageUrL: "url2#L1", FileHash: "hash2"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			changedFootnotes: map[RequirementID]bool{
// 				"pkg1/REQ001": true,
// 			},
// 			want: map[FilePath]*Reqmdjson{
// 				"docs/pkg1": {
// 					FileURL2FileHash: map[string]string{
// 						"url1": "new_hash1",
// 					},
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Create analyzer and cast to concrete type
// 			a := NewAnalyzer().(*analyzer)

// 			// Initialize analyzer state
// 			result, err := a.Analyze(tt.files)
// 			require.NoError(t, err)

// 			// Set changed footnotes
// 			a.changedFootnotes = tt.changedFootnotes

// 			// Process coverages
// 			for requirementID, coverage := range a.coverages {
// 				if a.changedFootnotes[requirementID] {
// 					// Add new coverers
// 					coverage.NewCoverers = []*Coverer{
// 						{
// 							CoverageLabel: "new.go:1:impl",
// 							CoverageUrL:   "new_url#L1",
// 							FileHash:      "new_hash",
// 						},
// 					}
// 				}
// 			}

// 			// Clear result reqmdjsons before rebuilding
// 			result.Reqmdjsons = make(map[FilePath]*Reqmdjson)

// 			// Rebuild reqmdjsons
// 			a.buildReqmdjsons(result)

//				// Verify result
//				assert.Equal(t, tt.want, result.Reqmdjsons)
//			})
//		}
//	}
//
// Helper function to create a simple FileStructure with one annotated requirement that has cw coverage
func createMdStructureA(path, pkgID string, line int, reqName string, cw CoverageStatusWord) FileStructure {

	emoji := CoverageStatusEmojiUncvrd
	if cw == CoverageStatusWordCovered {
		emoji = CoverageStatusEmojiCovered
	}

	fs := FileStructure{
		Path:      path,
		Type:      FileTypeMarkdown,
		PackageID: pkgID,
		Requirements: []RequirementSite{
			{
				FilePath:            path,
				RequirementName:     reqName,
				Line:                line,
				HasAnnotationRef:    true,
				CoverageStatusWord:  cw,
				CoverageStatusEmoji: emoji,
			},
		},
	}

	if cw == CoverageStatusWordCovered {
		fs.CoverageFootnotes = []CoverageFootnote{
			{
				RequirementName: reqName,
				Line:            line + 10,
				ID:              1,
				Coverers: []Coverer{
					{
						CoverageLabel: "somefolder/somefile.go:15:impl",
						CoverageUrL:   "someurl",
						FileHash:      "somehash",
					},
				},
			},
		}
	}

	return fs
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
