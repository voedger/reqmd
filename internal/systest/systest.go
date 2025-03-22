// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systest

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/voedger/reqmd/internal"
)

// SysTestFixture represents a loaded test environment for SysTests
type SysTestFixture struct {
	TestID   string
	TempReqs string
	TempSrc  string
}

// ExecRootCmdFunc defines the signature for the main execRootCmd function
type ExecRootCmdFunc func(args []string, version string) error

// RunSysTest executes a system test with the given parameters
func RunSysTest(t *testing.T, fs embed.FS, testID string, args []string, version string) {
	// Find sysTestData folder using fs and testID
	sysTestDataFolder, err := findSysTestDataFolder(fs, testID)
	require.NoError(t, err, "Failed to find sysTestData folder for testID: %s", testID)

	// Validate sysTestData folder (MUST contain reqs and src folders)
	validateSysTestDataFolder(t, fs, sysTestDataFolder)

	// Create temporary directories for reqs (tempReqs) and src (tempSrc)
	tempReqs := t.TempDir()
	tempSrc := t.TempDir()

	// Create git repos for tempReqs and tempSrc
	createGitRepo(t, tempReqs)
	createGitRepo(t, tempSrc)

	// Copy sysTestData.reqs to tempReqs and sysTestData.src to tempSrc
	copyEmbeddedFolder(t, fs, filepathJoin(sysTestDataFolder, "reqs"), tempReqs)
	copyEmbeddedFolder(t, fs, filepathJoin(sysTestDataFolder, "src"), tempSrc)

	// Commit all files in tempSrc
	commitAllFiles(t, tempSrc)

	// Find commitHash for tempSrc
	commitHash := getCommitHash(t, tempSrc)

	// Replace placeholders in all files in the tempReqs folder with commitHash
	replacePlaceholders(t, tempReqs, commitHash)

	// Prepare args to include tempReqs and tempSrc
	testArgs := append([]string{"reqmd"}, args...)
	testArgs = append(testArgs, tempReqs, tempSrc)

	// Run main.execRootCmd using args and version
	// Using a buffer to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	err = execRootCmd(testArgs, version, &stdout, &stderr)

	// Check errors if stderr is not empty
	if stderr.Len() > 0 {
		validateErrors(t, &stderr, tempReqs)
	}

	// Validate the tempReqs against GoldenData
	validateResults(t, fs, sysTestDataFolder, tempReqs)

	// If execRootCmd returned an error and stderr was empty, it's an unexpected error
	if err != nil && stderr.Len() == 0 {
		t.Fatalf("Unexpected error in execRootCmd: %v", err)
	}
}

// findSysTestDataFolder locates the test data folder for the given testID
func findSysTestDataFolder(_ embed.FS, testID string) (string, error) {
	return filepathJoin("testdata", testID), nil
}

// validateSysTestDataFolder ensures the test data folder has the required structure
func validateSysTestDataFolder(t *testing.T, fs embed.FS, folder string) {
	reqsDir := filepath.ToSlash(filepathJoin(folder, "reqs"))
	_, err := fs.ReadDir(reqsDir)
	require.NoError(t, err, "Failed to read reqs folder")

	srcDir := filepath.ToSlash(filepathJoin(folder, "src"))
	_, err = fs.ReadDir(srcDir)
	require.NoError(t, err, "Failed to read src folder")
}

// createGitRepo initializes a git repository in the given directory
func createGitRepo(t *testing.T, dir string) {
	// Initialize repository
	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err, "Failed to initialize git repo in %s", dir)

	// Configure git user for commit
	config, err := repo.Config()
	require.NoError(t, err, "Failed to get git config")

	config.User.Name = "Test User"
	config.User.Email = "test@example.com"

	err = repo.SetConfig(config)
	require.NoError(t, err, "Failed to set git config")
}

// copyEmbeddedFolder copies files from embedded FS to target directory
func copyEmbeddedFolder(t *testing.T, fs embed.FS, sourceDir, targetDir string) {
	// Read the source directory
	entries, err := fs.ReadDir(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist in embedded FS, which is fine
			return
		}
		require.NoError(t, err, "Failed to read directory: %s", sourceDir)
	}

	// Copy each entry
	for _, entry := range entries {
		sourcePath := filepathJoin(sourceDir, entry.Name())
		targetPath := filepathJoin(targetDir, entry.Name())

		if entry.IsDir() {
			// Create the target directory
			err := os.MkdirAll(targetPath, 0755)
			require.NoError(t, err, "Failed to create directory: %s", targetPath)

			// Recursively copy the subdirectory
			copyEmbeddedFolder(t, fs, sourcePath, targetPath)
		} else {
			// Read the file content
			content, err := fs.ReadFile(sourcePath)
			require.NoError(t, err, "Failed to read file: %s", sourcePath)

			// Write the file content to the target path
			err = os.WriteFile(targetPath, content, 0644)
			require.NoError(t, err, "Failed to write file: %s", targetPath)
		}
	}
}

// commitAllFiles adds and commits all files in the given directory
func commitAllFiles(t *testing.T, dir string) {
	repo, err := git.PlainOpen(dir)
	require.NoError(t, err, "Failed to open git repository in %s", dir)

	// Get the worktree
	wt, err := repo.Worktree()
	require.NoError(t, err, "Failed to get worktree in %s", dir)

	// Add all files
	_, err = wt.Add(".")
	require.NoError(t, err, "Failed to add files in %s", dir)

	// Commit files
	_, err = wt.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			// When:  object.TimeNow(),
		},
	})
	require.NoError(t, err, "Failed to commit files in %s", dir)
}

// getCommitHash returns the current commit hash for the repository
func getCommitHash(t *testing.T, dir string) string {
	repo, err := git.PlainOpen(dir)
	require.NoError(t, err, "Failed to open git repository in %s", dir)

	// Get HEAD reference
	ref, err := repo.Head()
	require.NoError(t, err, "Failed to get HEAD reference in %s", dir)

	return ref.Hash().String()
}

// replacePlaceholders replaces {{.CommitHash}} in all files with the actual commitHash
func replacePlaceholders(t *testing.T, dir string, commitHash string) {
	// Walk through all files in the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Check if the file contains the placeholder
		if !bytes.Contains(content, []byte("{{.CommitHash}}")) {
			return nil
		}

		// Create a template from the file content
		tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
		if err != nil {
			return err
		}

		// Apply the template with the commitHash
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]string{"CommitHash": commitHash})
		if err != nil {
			return err
		}

		// Write the result back to the file
		return os.WriteFile(path, buf.Bytes(), info.Mode())
	})

	require.NoError(t, err, "Failed to replace placeholders in %s", dir)
}

// execRootCmd redirects stdout and stderr to capture output and call the main package's execRootCmd
func execRootCmd(args []string, version string, stdout, stderr io.Writer) error {
	// Save the original stdout and stderr
	oldStdout, oldStderr := os.Stdout, os.Stderr

	// Create pipe readers and writers for stdout and stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		return err
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		return err
	}

	// Set the writers as stdout and stderr
	os.Stdout, os.Stderr = wOut, wErr

	// Create channels to prevent deadlocks
	outC := make(chan struct{})
	errC := make(chan struct{})

	// Copy the output in separate goroutines
	go func() {
		_, _ = io.Copy(stdout, rOut)
		close(outC)
	}()
	go func() {
		_, _ = io.Copy(stderr, rErr)
		close(errC)
	}()

	// Call the main function
	err = internal.ExecRootCmd(args, version)

	// Close the writers and wait for copying to complete
	wOut.Close()
	wErr.Close()
	<-outC
	<-errC

	// Restore stdout and stderr
	os.Stdout, os.Stderr = oldStdout, oldStderr

	return err
}

// validateErrors checks if the stderr output matches expected error patterns
// Lines are formatted as : `fmt.Sprintf("%s:%d: %s", err.FilePath, err.Line, err.Message)`
// All lines in serr must match at least one GoldenErrors
// All GoldenErrors must match at least one line in serr
func validateErrors(t *testing.T, stderr *bytes.Buffer, tempReqs string) {
	// Split stderr into lines
	stderrLines := strings.Split(strings.TrimSpace(stderr.String()), "\n")

	// Create a map to track which stderr lines have been matched
	stderrMatched := make(map[string]bool)
	for _, line := range stderrLines {
		stderrMatched[line] = false
	}

	// Map to collect all golden error patterns
	goldenErrors := make(map[string]bool)

	// Read all markdown files in tempReqs to find GoldenErrors
	err := filepath.Walk(tempReqs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse content line by line to find GoldenErrors
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			// Skip if not a line with expected errors
			if !strings.HasPrefix(line, "// error: ") {
				continue
			}

			// Extract error regexes from the line
			errLine := strings.TrimPrefix(line, "// error: ")
			errRegexes := extractErrorRegexes(errLine)

			// The line number to check is i (0-based) for previous line
			lineNum := i
			if lineNum > 0 {
				lineNum-- // Reference the line before the error comment
			}

			// Create patterns for matching stderr lines
			for _, regex := range errRegexes {
				pattern := fmt.Sprintf(`%s:%d: .*%s`, regexp.QuoteMeta(filepath.Base(path)), lineNum+1, regex)
				goldenErrors[pattern] = false

				// Try to match this pattern against stderr lines
				for stderrLine := range stderrMatched {
					matched, err := regexp.MatchString(pattern, stderrLine)
					require.NoError(t, err, "Invalid error regex pattern: %s", pattern)
					if matched {
						stderrMatched[stderrLine] = true
						goldenErrors[pattern] = true
					}
				}
			}
		}

		return nil
	})

	require.NoError(t, err, "Failed to validate errors")

	// Check that all stderr lines were matched
	for line, matched := range stderrMatched {
		assert.True(t, matched, "Unexpected error in stderr: %s", line)
	}

	// Check that all golden errors were matched
	for pattern, matched := range goldenErrors {
		assert.True(t, matched, "Expected error not found in stderr: %s", pattern)
	}
}

// extractErrorRegexes extracts error regexes from a line like 'error: "regex1" "regex2"'
func extractErrorRegexes(line string) []string {
	var regexes []string
	regex := regexp.MustCompile(`"([^"]*)"`)

	matches := regex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) > 1 {
			regexes = append(regexes, match[1])
		}
	}
	return regexes
}

// validateResults checks if the files in tempReqs match the expected GoldenData
func validateResults(t *testing.T, fs embed.FS, sysTestDataFolder, tempReqs string) {
	// Walk through all markdown files in tempReqs
	err := filepath.Walk(tempReqs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse content line by line
		lines := strings.Split(string(content), "\n")

		// For each line, check if the next line contains GoldenData
		for i := 0; i < len(lines)-1; i++ {
			nextLine := lines[i+1]

			// If next line starts with "// reqsite" or "// footnote", it contains GoldenData
			if strings.HasPrefix(nextLine, "// reqsite") || strings.HasPrefix(nextLine, "// footnote") {
				// Extract the expected pattern from the GoldenData line
				goldenData := strings.TrimPrefix(nextLine, "// ")

				// Replace backticks with double quotes in GoldenData
				goldenData = strings.ReplaceAll(goldenData, "`", "\"")

				// Check if the current line matches the expected pattern
				currentLine := lines[i]
				assert.Contains(t, currentLine, goldenData,
					"Line content doesn't match GoldenData at line %d in %s", i+1, path)
			}
		}

		return nil
	})

	require.NoError(t, err, "Failed to validate results")

	// Check for GoldenReqmd
	validateGoldenReqmd(t, fs, sysTestDataFolder, tempReqs)
}

// validateGoldenReqmd checks if reqmd.json files match their golden counterparts
func validateGoldenReqmd(t *testing.T, fs embed.FS, sysTestDataFolder, tempReqs string) {
	// Find all reqmd.json files in tempReqs
	err := filepath.Walk(tempReqs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() != "reqmd.json" {
			return nil
		}

		// Get relative path from tempReqs
		relPath, err := filepath.Rel(tempReqs, path)
		if err != nil {
			return err
		}

		// Check if there's a corresponding reqmd-golden.json
		goldenPath := filepathJoin(filepath.Dir(relPath), "reqmd-golden.json")
		goldenFullPath := filepathJoin(sysTestDataFolder, "reqs", goldenPath)

		// Try to read the golden file
		goldenContent, err := fs.ReadFile(goldenFullPath)
		if os.IsNotExist(err) {
			// No golden file - skip validation
			return nil
		}
		if err != nil {
			return err
		}

		// Read the actual reqmd.json
		actualContent, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Compare contents
		assert.Equal(t, string(goldenContent), string(actualContent),
			"reqmd.json doesn't match golden file at %s", goldenFullPath)

		return nil
	})

	require.NoError(t, err, "Failed to validate reqmd.json files")
}

func filepathJoin(elem ...string) string {
	return filepath.ToSlash(filepath.Join(elem...))
}
