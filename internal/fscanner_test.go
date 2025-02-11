package internal

import (
	"errors"
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
	root, err := os.MkdirTemp("", "folders-scanner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(root)

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
	if err := createTestStructure(t, root, structure); err != nil {
		t.Fatalf("Failed to create test structure: %v", err)
	}

	tests := []struct {
		name          string
		nroutines     int
		errorOnFile   string
		errorOnFolder string
		expectErrors  bool
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var processedFiles []string
			var processedFolders []string
			var mu sync.Mutex

			// Create folder processor
			fp := func(folder string) (FileProcessor, error) {
				relPath, _ := filepath.Rel(root, folder)
				mu.Lock()
				processedFolders = append(processedFolders, relPath)
				mu.Unlock()

				if tt.errorOnFolder != "" && strings.HasSuffix(folder, tt.errorOnFolder) {
					return nil, errors.New("folder processor error")
				}

				return func(filePath string) error {
					relPath, _ := filepath.Rel(root, filePath)
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
			errs := FoldersScanner(tt.nroutines, root, fp)

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
