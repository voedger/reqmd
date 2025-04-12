// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0
package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// testStructure represents a filesystem structure for testing
type testStructure struct {
	name     string
	isDir    bool
	content  string
	children []testStructure
}

// createTestStructure creates a directory structure in the given root path
func createTestStructure(t *testing.T, root string, structure testStructure) error {
	path := filepath.Join(root, structure.name)

	if structure.isDir {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
		for _, child := range structure.children {
			if err := createTestStructure(t, path, child); err != nil {
				return err
			}
		}
	} else {
		if err := os.WriteFile(path, []byte(structure.content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func TestFoldersScanner(t *testing.T) {

	// Create temporary root directory
	root := t.TempDir()

	// Convert root to absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Define test structure
	structure := testStructure{
		name:  "root",
		isDir: true,
		children: []testStructure{
			{
				name:  "dir1",
				isDir: true,
				children: []testStructure{
					{name: "file1.txt", content: "content1"},
					{name: "file2.txt", content: "content2"},
				},
			},
			{
				name:  "dir2",
				isDir: true,
				children: []testStructure{
					{name: "file3.txt", content: "content3"},
					{
						name:  "subdir",
						isDir: true,
						children: []testStructure{
							{name: "file4.txt", content: "content4"},
						},
					},
				},
			},
			{name: "root-file.txt", content: "root content"},
		},
	}

	// Create test structure
	if err := createTestStructure(t, absRoot, structure); err != nil {
		t.Fatalf("Failed to create test structure: %v", err)
	}

	tests := []struct {
		name          string
		nroutines     int
		errorOnFile   string
		errorOnFolder string
		expectErrors  bool
		skipFolder    string
	}{
		{
			name:      "successful processing",
			nroutines: 2,
		},
		{
			name:      "single routine",
			nroutines: 1,
		},
		{
			name:         "error on specific file",
			nroutines:    2,
			errorOnFile:  "file2.txt",
			expectErrors: true,
		},
		{
			name:          "error on specific folder",
			nroutines:     2,
			errorOnFolder: "dir2",
			expectErrors:  true,
		},
		{
			name:         "invalid routine count",
			nroutines:    0,
			expectErrors: true,
		},
		{
			name:      "verify absolute paths",
			nroutines: 1,
		},
		{
			name:       "skip specific folder",
			nroutines:  2,
			skipFolder: "dir2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var processedFiles []string
			var processedFolders []string
			var mu sync.Mutex

			// Create folder processor
			fp := func(folder string) (FileProcessor, error) {
				// Verify folder path is absolute
				if tt.name == "verify absolute paths" && !filepath.IsAbs(folder) {
					t.Errorf("Expected absolute folder path, got: %s", folder)
				}

				relPath, _ := filepath.Rel(absRoot, folder)
				mu.Lock()
				processedFolders = append(processedFolders, relPath)
				mu.Unlock()

				if tt.errorOnFolder != "" && strings.HasSuffix(folder, tt.errorOnFolder) {
					return nil, errors.New("folder processor error")
				}

				if tt.skipFolder != "" && strings.HasSuffix(folder, tt.skipFolder) {
					return nil, nil
				}

				return func(filePath string) error {
					// Verify file path is absolute
					if tt.name == "verify absolute paths" && !filepath.IsAbs(filePath) {
						t.Errorf("Expected absolute file path, got: %s", filePath)
					}

					relPath, _ := filepath.Rel(absRoot, filePath)
					mu.Lock()
					processedFiles = append(processedFiles, relPath)
					mu.Unlock()

					if tt.errorOnFile != "" && strings.HasSuffix(filePath, tt.errorOnFile) {
						return errors.New("file processor error")
					}
					return nil
				}, nil
			}

			// Run scanner
			errs := FoldersScanner(tt.nroutines, 100, absRoot, fp)

			// Verify results
			if tt.expectErrors && len(errs) == 0 {
				t.Error("Expected errors but got none")
			}
			if !tt.expectErrors && len(errs) > 0 {
				t.Errorf("Expected no errors but got: %v", errs)
			}

			// Sort processed files and folders for deterministic comparison
			sort.Strings(processedFiles)
			sort.Strings(processedFolders)

			// Verify processed files if no folder errors
			if tt.errorOnFolder == "" && tt.nroutines > 0 {
				expectedFiles := []string{
					filepath.Join("root", "dir1", "file1.txt"),
					filepath.Join("root", "dir1", "file2.txt"),
					filepath.Join("root", "dir2", "file3.txt"),
					filepath.Join("root", "dir2", "subdir", "file4.txt"),
					filepath.Join("root", "root-file.txt"),
				}

				// Adjust expected files if a folder is being skipped
				if tt.skipFolder != "" {
					filteredFiles := []string{}
					for _, file := range expectedFiles {
						if !strings.Contains(file, tt.skipFolder) {
							filteredFiles = append(filteredFiles, file)
						}
					}
					expectedFiles = filteredFiles
				}

				sort.Strings(expectedFiles)

				if !tt.expectErrors {
					if len(processedFiles) != len(expectedFiles) {
						t.Errorf("Processed files count mismatch: got %d, want %d",
							len(processedFiles), len(expectedFiles))
					}
					for i := range expectedFiles {
						if i >= len(processedFiles) {
							break
						}
						if processedFiles[i] != expectedFiles[i] {
							t.Errorf("Processed file mismatch at %d: got %s, want %s",
								i, processedFiles[i], expectedFiles[i])
						}
					}
				}
			}

			// Verify breadth-first order of folder processing
			if tt.nroutines > 0 && len(processedFolders) > 0 {
				var lastDepth int
				for i, folder := range processedFolders {
					depth := strings.Count(folder, string(os.PathSeparator))
					if i > 0 && depth < lastDepth-1 {
						t.Error("Folders not processed in breadth-first order")
					}
					lastDepth = depth
				}
			}
		})
	}
}

// TestFoldersScanner_ALotOfErrors tests the FoldersScanner function with a large number of errors that more than the error channel capacity
func TestFoldersScanner_ALotOfErrors(t *testing.T) {
	// Create temporary root directory
	root := t.TempDir()

	// Convert root to absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create test structure with many files
	structure := testStructure{
		name:  "root",
		isDir: true,
		children: []testStructure{
			{
				name:     "dir1",
				isDir:    true,
				children: make([]testStructure, 50), // 50 files in dir1
			},
			{
				name:     "dir2",
				isDir:    true,
				children: make([]testStructure, 50), // 50 files in dir2
			},
		},
	}

	// Initialize files in dir1 and dir2
	for i := range structure.children[0].children {
		structure.children[0].children[i] = testStructure{
			name:    fmt.Sprintf("file%d.txt", i),
			content: fmt.Sprintf("content%d", i),
		}
	}
	for i := range structure.children[1].children {
		structure.children[1].children[i] = testStructure{
			name:    fmt.Sprintf("file%d.txt", i),
			content: fmt.Sprintf("content%d", i),
		}
	}

	// Create test structure
	if err := createTestStructure(t, absRoot, structure); err != nil {
		t.Fatalf("Failed to create test structure: %v", err)
	}

	var processedCount int32
	var mu sync.Mutex

	// Create folder processor that generates an error for every file
	fp := func(folder string) (FileProcessor, error) {
		return func(filePath string) error {
			mu.Lock()
			processedCount++
			mu.Unlock()
			return fmt.Errorf("error processing file: %s", filePath)
		}, nil
	}

	// Run scanner with small error channel capacity
	errs := FoldersScanner(4, 10, absRoot, fp)

	// Verify that we processed all files despite error channel being full
	mu.Lock()
	if processedCount != 100 { // 50 files in each directory
		t.Errorf("Expected 100 processed files, got %d", processedCount)
	}
	mu.Unlock()

	// Verify that we got some errors but not necessarily all of them
	if len(errs) == 0 {
		t.Error("Expected some errors but got none")
	}
	if len(errs) > 10 {
		t.Errorf("Expected at most 10 errors (channel capacity), got %d", len(errs))
	}
}
