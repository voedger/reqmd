// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	cfg "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
	"github.com/voedger/reqmd/internal"
)

// Uses .testdata/Test_gogit as a TestFolder
// Removes TestFolder before running this test
// Creates a git repo TestFolder
// Creates a subfolder TestFolder/subfolder1/subfolder2/
// Creates 4 text files in TestFolder, named 1.txt, 2.txt, 3.txt, 4.txt
// Creates 4 text files in subfolder1, named 11.txt, 12.txt, 13.txt, 14.txt
// Creates 4 text files in subfolder2, named 21.txt, 22.txt, 23.txt, 24.txt
// Opens repo using NewIGit
// Checks that Hash can be obtained for all files in TestFolder and its subfolders
// Checks that TestFolder can be obtained as a root folder of the Repository
func Test_IGit(t *testing.T) {
	// Use .testdata/gogit_test as TestFolder
	testFolder := ".testdata/Test_IGit"

	// Create path for subfolders
	subfolder1 := filepath.Join(testFolder, "subfolder1")
	subfolder2 := filepath.Join(subfolder1, "subfolder2")

	// Removes TestFolder before running this test
	_ = os.RemoveAll(testFolder)

	// Create directories
	if err := os.MkdirAll(subfolder2, 0755); err != nil {
		t.Fatalf("failed to create subfolders: %v", err)
	}

	// Creates 4 text files in TestFolder, named 1.txt, 2.txt, 3.txt, 4.txt
	baseFiles := []string{"1.txt", "2.txt", "3.txt", "4.txt"}
	for _, f := range baseFiles {
		if err := os.WriteFile(filepath.Join(testFolder, f), []byte(f+" content"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", f, err)
		}
	}

	// Creates 4 text files in subfolder1, named 11.txt, 12.txt, 13.txt, 14.txt
	sub1Files := []string{"11.txt", "12.txt", "13.txt", "14.txt"}
	for _, f := range sub1Files {
		if err := os.WriteFile(filepath.Join(subfolder1, f), []byte(f+" content"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", f, err)
		}
	}

	// Creates 4 text files in subfolder2, named 21.txt, 22.txt, 23.txt, 24.txt
	sub2Files := []string{"21.txt", "22.txt", "23.txt", "24.txt"}
	for _, f := range sub2Files {
		if err := os.WriteFile(filepath.Join(subfolder2, f), []byte(f+" content"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", f, err)
		}
	}

	// Creates a git repo TestFolder
	repo, err := gogit.PlainInit(testFolder, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Add and commit files
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	if _, err := wt.Add("."); err != nil {
		t.Fatalf("failed to add files: %v", err)
	}

	_, err = wt.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Add a remote origin for test purposes
	_, err = repo.CreateRemote(&cfg.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/voedger/example"},
	})
	require.NoError(t, err)

	// Opens repo using NewIGit
	var igit internal.IGit
	igit, err = internal.NewIGit(testFolder)
	require.NoError(t, err)

	// Checks that Hash can be obtained for all files in TestFolder and its subfolders
	// Check files in the root folder
	for _, f := range baseFiles {
		filePath := filepath.Join(testFolder, f)
		filePath, err = filepath.Abs(filePath)
		require.NoError(t, err)
		relPath, hash, err := igit.FileHash(filePath)
		require.NoError(t, err, "Should get hash for file %s", f)
		require.NotEmpty(t, hash, "Hash should not be empty for file %s", f)
		require.Equal(t, f, relPath, "Relative path should match the file name for %s", f)
	}

	// Check files in subfolder1
	for _, f := range sub1Files {
		filePath := filepath.Join(subfolder1, f)
		filePath, err = filepath.Abs(filePath)
		require.NoError(t, err)
		relPath, hash, err := igit.FileHash(filePath)
		require.NoError(t, err, "Should get hash for file %s", filePath)
		require.NotEmpty(t, hash, "Hash should not be empty for file %s", f)
		require.Equal(t, "subfolder1/"+f, relPath, "Relative path should match for %s", f)
	}

	// Check files in subfolder2
	for _, f := range sub2Files {
		filePath := filepath.Join(subfolder2, f)
		filePath, err = filepath.Abs(filePath)
		require.NoError(t, err)
		relPath, hash, err := igit.FileHash(filePath)
		require.NoError(t, err, "Should get hash for file %s", f)
		require.NotEmpty(t, hash, "Hash should not be empty for file %s", f)
		require.Equal(t, "subfolder1/subfolder2/"+f, relPath, "Relative path should match for %s", f)
	}

	// Checks that TestFolder can be obtained as a root folder of the Repository
	rootPath := igit.PathToRoot()
	absTestFolder, err := filepath.Abs(testFolder)
	require.NoError(t, err)
	require.Equal(t, filepath.ToSlash(absTestFolder), rootPath, "Root path should match test folder")

	// Also verify the repo URL construction
	repoURL := igit.RepoRootFolderURL()
	require.Contains(t, repoURL, "https://github.com/voedger/example/blob/", "Repo URL should be properly constructed")
}
