// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/go-git/go-git/v5"
	cfg "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	RemoteOrigin = "https://github.com/voedger/example"
)

type T interface {
	Errorf(format string, args ...interface{})
	FailNow()
	TempDir() string
}

// ExecRootCmdFunc defines the signature for the main execRootCmd function
type ExecRootCmdFunc func(args []string, version string) error

// RunSysTest executes a system test with the given parameters
func RunSysTest(t T, testsDir string, testID string, rootCmd ExecRootCmdFunc, args []string, version string) {
	// Find sysTestData Dir using testID
	sysTestDataDir, err := findSysTestDataDir(testsDir, testID)
	require.NoError(t, err, "Failed to find sysTestData Dir for testID: %s", testID)

	// Validate sysTestData Dir (MUST contain req and src Dirs)
	validateSysTestDataDir(t, sysTestDataDir)

	// Create temporary directories for req (tempReqs) and src (tempSrc)
	tempReqs := t.TempDir()
	tempSrc := t.TempDir()

	// Create git repos for tempReqs and tempSrc
	createGitRepo(t, tempReqs)
	createGitRepo(t, tempSrc)

	// Copy sysTestData.req to tempReqs and sysTestData.src to tempSrc
	copyDir(t, filepathJoin(sysTestDataDir, "req"), tempReqs)
	copyDir(t, filepathJoin(sysTestDataDir, "src"), tempSrc)

	// parseReqGoldenData
	grd, err := parseReqGoldenData(tempReqs)
	require.NoError(t, err, "Failed to parse req golden data")

	// Commit all files in tempSrc
	commitAllFiles(t, tempSrc)

	// Find commitHash for tempSrc
	commitHash := getCommitHash(t, tempSrc)

	// Replace placeholders in all files in the tempReqs Dir with commitHash
	replacePlaceholders(t, tempReqs, commitHash)

	// Prepare args to include tempReqs and tempSrc
	testArgs := append([]string{"reqmd"}, args...)
	testArgs = append(testArgs, tempReqs, tempSrc)

	// Run main.execRootCmd using args and version
	// Using a buffer to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	_ = execRootCmd(rootCmd, testArgs, version, &stdout, &stderr)

	// Check errors
	validateErrors(t, &stderr, tempReqs, grd)

	// Validate the tempReqs against GoldenData
	validateTempReqs(t, grd, tempReqs)

	// Check for GoldenReqmd
	validateReqmd(t, sysTestDataDir, tempReqs)
}

// findSysTestDataDir locates the test data Dir for the given testID
func findSysTestDataDir(testsDir string, testID string) (string, error) {
	return filepathJoin(testsDir, testID), nil
}

// validateSysTestDataDir ensures the test data Dir has the required structure
func validateSysTestDataDir(t T, Dir string) {
	reqDir := filepath.ToSlash(filepathJoin(Dir, "req"))
	_, err := os.Stat(reqDir)
	require.NoError(t, err, "Failed to read `req` dir")

	srcDir := filepath.ToSlash(filepathJoin(Dir, "src"))
	_, err = os.Stat(srcDir)
	require.NoError(t, err, "Failed to read `src` dir")
}

// createGitRepo initializes a git repository in the given directory
func createGitRepo(t T, dir string) {
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

	// Add a remote origin for test purposes
	_, err = repo.CreateRemote(&cfg.RemoteConfig{
		Name: "origin",
		URLs: []string{RemoteOrigin},
	})
	require.NoError(t, err, "Failed to create origin remote")
}

// copyDir copies files from source directory to target directory
func copyDir(t T, sourceDir, targetDir string) {
	// Read the source directory
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, which is fine
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
			copyDir(t, sourcePath, targetPath)
		} else {
			// Read the file content
			content, err := os.ReadFile(sourcePath)
			require.NoError(t, err, "Failed to read file: %s", sourcePath)

			// Write the file content to the target path
			err = os.WriteFile(targetPath, content, 0644)
			require.NoError(t, err, "Failed to write file: %s", targetPath)
		}
	}
}

// commitAllFiles adds and commits all files in the given directory
func commitAllFiles(t T, dir string) {
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
		},
	})
	require.NoError(t, err, "Failed to commit files in %s", dir)
}

// getCommitHash returns the current commit hash for the repository
func getCommitHash(t T, dir string) string {
	repo, err := git.PlainOpen(dir)
	require.NoError(t, err, "Failed to open git repository in %s", dir)

	// Get HEAD reference
	ref, err := repo.Head()
	require.NoError(t, err, "Failed to get HEAD reference in %s", dir)

	return ref.Hash().String()
}

// replacePlaceholders replaces {{.CommitHash}} in all files with the actual commitHash
func replacePlaceholders(t T, dir string, commitHash string) {
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
func execRootCmd(rootCmd ExecRootCmdFunc, args []string, version string, stdout, stderr io.Writer) error {
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
	err = rootCmd(args, version)

	// Close the writers and wait for copying to complete
	wOut.Close()
	wErr.Close()
	<-outC
	<-errC

	// Restore stdout and stderr
	os.Stdout, os.Stderr = oldStdout, oldStderr

	return err
}

// validateErrors checks if the stderr output matches expected error patterns from goldenReqData
// stderr lines are parsed into `path`, `line` and `message` parts according to the formatting: `fmt.Sprintf("%s:%d: %s", err.FilePath, err.Line, err.Message)`
// All lines in stderr must match at least one item in grd.errors
// All grd.errors items must match at least one line in stderr
// stderr lines and grd.errors items are matched using all parts of the stderr lines: `path`, `line` and `message`
func validateErrors(t T, stderr *bytes.Buffer, tempReqs string, grd *goldenReqData) {
	// If no errors are expected and none occurred, return successfully
	if len(grd.errors) == 0 && stderr.Len() == 0 {
		return
	}

	// Parse stderr lines into structured format
	stderrLines := strings.Split(stderr.String(), "\n")
	parsedErrors := make(map[string]bool)

	// Regular expression to match error lines in format "path:line: message"
	errRegex := regexp.MustCompile(`^(.+):(\d+): (.+)$`)

	for _, line := range stderrLines {
		if line == "" {
			continue
		}

		matches := errRegex.FindStringSubmatch(line)
		if len(matches) != 4 {
			// This line doesn't match our expected format
			t.Errorf("Unexpected error format: %s", line)
			continue
		}

		filePath := matches[1]
		lineNum, _ := strconv.Atoi(matches[2])
		message := matches[3]

		// Make path relative to tempReqs for comparison with golden data
		relPath, err := filepath.Rel(tempReqs, filePath)
		if err != nil {
			t.Errorf("Failed to get relative path for %s: %v", filePath, err)
			continue
		}

		// Normalize path for comparison
		relPath = filepath.ToSlash(relPath)

		// Check if this error matches any expected errors
		errorFound := false

		// Iterate through all expected errors in goldenReqData
		for goldFilePath, lineErrors := range grd.errors {
			// Get base filename for comparison
			goldFileName := filepath.Base(goldFilePath)
			errFileName := filepath.Base(filePath)

			// Check if filenames match
			if strings.EqualFold(goldFileName, errFileName) {
				// For each line number in the golden errors
				for goldLineNum, items := range lineErrors {
					// For each regex pattern for this line number
					for _, item := range items {
						// Create a test string that combines the elements for matching
						testString := fmt.Sprintf("%s:%d: %s", relPath, lineNum, message)

						// Check if the regex matches
						if item.regex.MatchString(testString) {
							errorFound = true
							// Mark this expected error as found
							key := fmt.Sprintf("%s:%d:%s", goldFilePath, goldLineNum, item.regex.String())
							parsedErrors[key] = true
							break
						}
					}
					if errorFound {
						break
					}
				}
			}
		}

		// If error doesn't match any expected errors, fail the test
		if !errorFound {
			t.Errorf("Unexpected error: %s", line)
		}
	}

	// Check that all expected errors were found
	for goldFilePath, lineErrors := range grd.errors {
		for goldLineNum, items := range lineErrors {
			for _, item := range items {
				key := fmt.Sprintf("%s:%d:%s", goldFilePath, goldLineNum, item.regex.String())
				if !parsedErrors[key] {
					t.Errorf("Expected error not found in stderr: %s line %d: %s", goldFilePath, goldLineNum, item.regex.String())
				}
			}
		}
	}
}

// validateTempReqs checks if the files in tempReqs match the expected goldenReqData
// - reqsites
// - footnotes
// - newfootnotes
func validateTempReqs(t T, grd *goldenReqData, tempReqsDir string) {

	if len(grd.reqsites) == 0 && len(grd.footnotes) == 0 && len(grd.newfootnotes) == 0 {
		// No validation data found, nothing to validate
		return
	}

	// Track which items have been found
	reqsitesFound := make(map[string]map[int]map[string]bool)
	footnotesFound := make(map[string]map[int]map[string]bool)
	newfootnotesFound := make(map[string]map[string]bool)

	// Initialize the tracking maps
	for filePath, lineItems := range grd.reqsites {
		if reqsitesFound[filePath] == nil {
			reqsitesFound[filePath] = make(map[int]map[string]bool)
		}
		for lineNum, items := range lineItems {
			if reqsitesFound[filePath][lineNum] == nil {
				reqsitesFound[filePath][lineNum] = make(map[string]bool)
			}
			for _, item := range items {
				reqsitesFound[filePath][lineNum][item.data] = false
			}
		}
	}

	for filePath, lineItems := range grd.footnotes {
		if footnotesFound[filePath] == nil {
			footnotesFound[filePath] = make(map[int]map[string]bool)
		}
		for lineNum, items := range lineItems {
			if footnotesFound[filePath][lineNum] == nil {
				footnotesFound[filePath][lineNum] = make(map[string]bool)
			}
			for _, item := range items {
				footnotesFound[filePath][lineNum][item.data] = false
			}
		}
	}

	for filePath, items := range grd.newfootnotes {
		if newfootnotesFound[filePath] == nil {
			newfootnotesFound[filePath] = make(map[string]bool)
		}
		for _, item := range items {
			newfootnotesFound[filePath][item.data] = false
		}
	}

	// Walk through all markdown files in tempReqs
	err := filepath.Walk(tempReqsDir, func(path string, info os.FileInfo, err error) error {
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

		// Get relative path to match with golden data
		relPath, err := filepath.Rel(tempReqsDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		// For each file in the golden data, check if this file matches (by base name)
		for goldFilePath := range grd.reqsites {
			goldBaseName := filepath.Base(goldFilePath)
			currentBaseName := filepath.Base(path)

			if strings.EqualFold(goldBaseName, currentBaseName) {
				validateFileContents(t, string(content), goldFilePath, relPath, reqsitesFound, "reqsite")
			}
		}

		for goldFilePath := range grd.footnotes {
			goldBaseName := filepath.Base(goldFilePath)
			currentBaseName := filepath.Base(path)

			if strings.EqualFold(goldBaseName, currentBaseName) {
				validateFileContents(t, string(content), goldFilePath, relPath, footnotesFound, "footnote")
			}
		}

		// Check for newfootnotes (these don't have line numbers, just check if they exist in the file)
		for goldFilePath, items := range grd.newfootnotes {
			goldBaseName := filepath.Base(goldFilePath)
			currentBaseName := filepath.Base(path)

			if strings.EqualFold(goldBaseName, currentBaseName) {
				for _, item := range items {
					if strings.Contains(string(content), item.data) {
						newfootnotesFound[goldFilePath][item.data] = true
					}
				}
			}
		}

		return nil
	})
	require.NoError(t, err, "Failed to walk through markdown files")

	// Verify all reqsites were found
	for filePath, lineItems := range reqsitesFound {
		for lineNum, items := range lineItems {
			for item, found := range items {
				assert.True(t, found, "Expected reqsite not found in %s at line %d: %s", filePath, lineNum, item)
			}
		}
	}

	// Verify all footnotes were found
	for filePath, lineItems := range footnotesFound {
		for lineNum, items := range lineItems {
			for item, found := range items {
				assert.True(t, found, "Expected footnote not found in %s at line %d: %s", filePath, lineNum, item)
			}
		}
	}

	// Verify all newfootnotes were found
	for filePath, items := range newfootnotesFound {
		for item, found := range items {
			assert.True(t, found, "Expected newfootnote not found in %s: %s", filePath, item)
		}
	}
}

// validateFileContents checks file contents for the existence of expected items
func validateFileContents(t T, content string, goldFilePath string, relPath string, itemsFound map[string]map[int]map[string]bool, itemType string) {
	lines := strings.Split(content, "\n")

	// For each line in the file
	for i, line := range lines {
		lineNum := i + 1 // Convert to 1-based line numbers

		// Check if this line number has expected items in the golden data
		if itemsFound[goldFilePath][lineNum] != nil {
			for expectedItem := range itemsFound[goldFilePath][lineNum] {
				if strings.Contains(line, expectedItem) {
					itemsFound[goldFilePath][lineNum][expectedItem] = true
				}
			}
		}
	}
}

// validateReqmd checks if reqmd.json files match their golden counterparts
func validateReqmd(t T, sysTestDataDir, tempReqs string) {
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
		goldenFullPath := filepathJoin(sysTestDataDir, "req", goldenPath)

		// Try to read the golden file
		goldenContent, err := os.ReadFile(goldenFullPath)
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
