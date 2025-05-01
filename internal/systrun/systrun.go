// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systrun

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

const (
	RemoteOrigin = "https://github.com/voedger/example"
)

type T interface {
	Errorf(format string, args ...interface{})
	FailNow()
	TempDir() string
	Helper()
}

// ExecRootCmdFunc defines the signature for the main execRootCmd function
type ExecRootCmdFunc func(args []string, version string) error

// RunSysTest executes a system test with the given parameters
// If testId contains no subfoldersthen a single git repo is created and reqmd receives a single folder as an argument
// Otherwise, each subfolder is treated as a separate path and separate git repos are created for each subfolder
func RunSysTest(t T, testsDir string, testId string, rootCmd ExecRootCmdFunc, version string) {

	t.Helper()

	// Find sysTestData Dir using testId
	sysTestDataDir, err := findSysTestDataDir(testsDir, testId)
	require.NoError(t, err, "Failed to find sysTestData Dir for testId: %s", testId)

	testFolderAbsPaths := []string{}
	{
		// Check if sysTestDataDir contains subdirectories
		entries, err := os.ReadDir(sysTestDataDir)
		require.NoError(t, err, "Failed to read sysTestData Dir: %s", sysTestDataDir)

		for _, entry := range entries {
			if entry.IsDir() {
				absPath := filepathJoin(sysTestDataDir, entry.Name())
				testFolderAbsPaths = append(testFolderAbsPaths, absPath)
			}
		}
		if len(testFolderAbsPaths) == 0 {
			testFolderAbsPaths = append(testFolderAbsPaths, sysTestDataDir)
		}
	}

	// Parse golden data from all requirement directories
	goldenData, err := parseGoldenData(testFolderAbsPaths)
	require.NoError(t, err, "Failed to parse golden data")

	var allTempFolders []string
	var testArgs []string = []string{"reqmd", "trace", "--ignore-lines", "^// line:"}

	commitHashes := make(map[string]string)

	// tempTestFolder keeps all test data for the current testId
	tempTestBaseFolder, err := filepath.Abs(filepath.Join(".testdata", testId))
	require.NoError(t, err, "Failed to get absolute path for tempTestBaseFolder: %s", tempTestBaseFolder)
	{
		// Remove tempTestFolder if it exists
		err = os.RemoveAll(tempTestBaseFolder)
		require.NoError(t, err, "Failed to remove tempTestFolder: %s", tempTestBaseFolder)

		// Create tempTestFolder
		err = os.MkdirAll(tempTestBaseFolder, 0755)
		require.NoError(t, err, "Failed to create tempTestFolder: %s", tempTestBaseFolder)
	}

	// Copy all test data to tempTestFolder and create git repos
	for _, testFolderAbsPath := range testFolderAbsPaths {
		tempFolderAlias := ""
		if len(testFolderAbsPaths) > 1 {
			tempFolderAlias = filepath.Base(testFolderAbsPath)
		}
		tempFolder := filepath.Join(tempTestBaseFolder, tempFolderAlias)
		err = os.MkdirAll(tempFolder, 0755)
		require.NoError(t, err, "Failed to create tempFolder: %s", tempFolder)

		allTempFolders = append(allTempFolders, tempFolder)

		// Copy root req and src to temp dirs
		copyDir(t, testFolderAbsPath, tempFolder)

		createGitRepo(t, tempFolder)
		commitAllFiles(t, tempFolder)
		commitHash := getCommitHash(t, tempFolder)
		commitHashes[tempFolderAlias] = commitHash

		// Prepare args
		testArgs = append(testArgs, tempFolder)
	}

	// Replace placeholders in tempReqs
	replacePlaceholders(t, goldenData, commitHashes)

	// Run main.execRootCmd using args and version
	var stdout, stderr bytes.Buffer
	_ = execRootCmd(rootCmd, testArgs, version, &stdout, &stderr)

	// Check errors against all requirement directories
	validateErrors(t, &stderr, allTempFolders, goldenData)

	// Validate golden lines against all requirement directories
	validateGoldenLines(t, goldenData, allTempFolders)

	// Keep tempFolders but remove .git subfolders
	for _, testFolderAbsPath := range allTempFolders {
		gitFolder := filepath.Join(testFolderAbsPath, ".git")
		err = os.RemoveAll(gitFolder)
		require.NoError(t, err, "Failed to remove git folder: %s", gitFolder)
	}

}

func validateGoldenLines(t T, goldenData *goldenData, paths []string) {

	t.Helper()

	// Skip validation if no golden lines are defined
	if len(goldenData.lines) == 0 {
		return
	}

	// For each file path in goldenData.lines
	for goldenPath, expectedLines := range goldenData.lines {
		// Try each tempReq directory to find the file
		found := false

		for _, tempReq := range paths {
			// Construct full path to the actual file
			actualPath := filepathJoin(tempReq, goldenPath)

			// Read the actual file content using loadFileLines
			actualLines, err := loadFileLines(actualPath)
			if err != nil {
				// Try the next tempReq if file not found in this one
				continue
			}

			found = true

			// Compare number of lines
			if len(actualLines) != len(expectedLines) {
				t.Errorf("Line count mismatch in %s: expected %d lines, got %d lines\n%s",
					goldenPath, len(expectedLines), len(actualLines), strings.Join(actualLines, "\n"))
				break
			}

			// Compare each line
			for i, expectedLine := range expectedLines {
				if i >= len(actualLines) {
					t.Errorf("Missing line %d in %s", i+1, goldenPath)
					continue
				}

				if actualLines[i] != expectedLine {
					t.Errorf("Line mismatch in %s at line %d:\nexpected: %s\ngot: %s",
						goldenPath, i+1, expectedLine, actualLines[i])
				}
			}

			// File was found and processed, no need to check other tempReqs
			break
		}

		if !found {
			t.Errorf("Failed to find file %s in any of the provided directories", goldenPath)
		}
	}
}

// findSysTestDataDir locates the test data Dir for the given testId
func findSysTestDataDir(testsDir string, testId string) (string, error) {
	return filepathJoin(testsDir, testId), nil
}

// createGitRepo initializes a git repository in the given directory
func createGitRepo(t T, dir string) {
	// Initialize repository
	err := runGitCommand(dir, "init", "-b", "main")
	require.NoError(t, err, "Failed to initialize git repo in %s", dir)

	// Configure git user for commit
	err = runGitCommand(dir, "config", "user.name", "Test User")
	require.NoError(t, err, "Failed to set git user name")
	err = runGitCommand(dir, "config", "user.email", "test@example.com")
	require.NoError(t, err, "Failed to set git user Email")

	// Add a remote origin for test purposes
	{
		origin := RemoteOrigin + "/" + filepath.Base(dir)
		err = runGitCommand(dir, "remote", "add", "origin", origin)
		require.NoError(t, err, "Failed to create origin remote")
	}
}

// runGitCommand executes a git command with the given arguments in the specified directory
func runGitCommand(dir string, args ...string) error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git command failed: %v\nStderr: %s", err, stderr.String())
	}
	return nil
}

// copyDir copies files from source directory to target directory
func copyDir(t T, sourceDir, targetDir string) {
	// Read the source directory
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, create target directory with empty .gitkeep file
			err = os.MkdirAll(targetDir, 0755)
			require.NoError(t, err, "Failed to create directory: %s", targetDir)

			// Create empty .gitkeep file
			gitkeepPath := filepathJoin(targetDir, ".gitkeep")
			err = os.WriteFile(gitkeepPath, []byte{}, 0644)
			require.NoError(t, err, "Failed to create .gitkeep file: %s", gitkeepPath)
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

// replacePlaceholders replaces {{.CommitHash}} in all goldenData.lines with the actual commit hash
func replacePlaceholders(_ T, goldenData *goldenData, commitHashes map[string]string) {
	// Replace in goldenData.lines
	for filePath, lines := range goldenData.lines {
		for i, line := range lines {
			if strings.Contains(line, "CommitHash") {
				for k, v := range commitHashes {
					placeholder := fmt.Sprintf("{{.CommitHash.%s}}", k)
					line = strings.ReplaceAll(line, placeholder, v)
					goldenData.lines[filePath][i] = line
				}
			}
		}
	}
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
// Is a stderr line starts with `\t` it is appended to the previous line
// All lines in stderr must match at least one item in grd.errors
// All grd.errors items must match at least one line in stderr
// stderr lines and grd.errors items are matched using all parts of the stderr lines: `path`, `line` and `message`
func validateErrors(t T, stderr *bytes.Buffer, tempReqs []string, grd *goldenData) {

	t.Helper()

	// If no errors are expected and none occurred, return successfully
	if len(grd.errors) == 0 && stderr.Len() == 0 {
		return
	}

	// Parse stderr lines into structured format
	stderrLines := strings.Split(stderr.String(), "\n")
	parsedErrors := make(map[string]bool)

	// Regular expression to match error lines in format "path:line: message"
	errRegex := regexp.MustCompile(`^(.+):(\d+): (.+)$`)

	// Process lines, handling indented lines (starting with \t)
	var processedLines []string
	var currentLine string

	for _, line := range stderrLines {
		if line == "" {
			if currentLine != "" {
				processedLines = append(processedLines, currentLine)
				currentLine = ""
			}
			continue
		}

		if strings.HasPrefix(line, "\t") {
			// If line starts with \t, append it to the previous line
			if currentLine != "" {
				currentLine += " " + strings.TrimSpace(line)
			}
		} else {
			// If it's a new line, add the previous completed line to processedLines
			if currentLine != "" {
				processedLines = append(processedLines, currentLine)
			}
			currentLine = line
		}
	}

	// Add the last line if it exists
	if currentLine != "" {
		processedLines = append(processedLines, currentLine)
	}

	for _, line := range processedLines {
		matches := errRegex.FindStringSubmatch(line)
		if len(matches) != 4 {
			// This line doesn't match our expected format
			t.Errorf("Unexpected error format: %s", line)
			continue
		}

		filePath := matches[1]
		lineNum, _ := strconv.Atoi(matches[2])
		message := matches[3]

		// Try to make path relative to each tempReq directory
		errorFound := false
		relPathFound := false

		for _, tempReq := range tempReqs {
			// Make path relative to tempReqs for comparison with golden data
			relPath, err := filepath.Rel(tempReq, filePath)
			if err != nil {
				continue
			}

			relPathFound = true

			// Normalize path for comparison
			relPath = filepath.ToSlash(relPath)

			// Check if this error matches any expected errors
			// Iterate through all expected errors in goldenReqData
			for goldFilePath, lineErrors := range grd.errors {
				// Get base filename for comparison
				goldFileName := filepath.Base(goldFilePath)
				errFileName := filepath.Base(filePath)

				// Check if filenames match
				if strings.EqualFold(goldFileName, errFileName) {
					// For each line number in the golden errors
					for goldLineNum, regexps := range lineErrors {
						// For each regex pattern for this line number
						for _, regexp := range regexps {
							// Create a test string that combines the elements for matching
							testString := fmt.Sprintf("%s:%d: %s", relPath, lineNum, message)

							// Check if the regex matches
							if regexp.MatchString(testString) {
								errorFound = true
								// Mark this expected error as found
								key := fmt.Sprintf("%s:%d:%s", goldFilePath, goldLineNum, regexp.String())
								parsedErrors[key] = true
								break
							}
						}
						if errorFound {
							break
						}
					}
				}
				if errorFound {
					break
				}
			}

			if errorFound {
				break
			}
		}

		// If we couldn't make a relative path from any tempReq directory
		if !relPathFound {
			t.Errorf("Failed to get relative path for %s from any tempReqs directory", filePath)
			continue
		}

		// If error doesn't match any expected errors, fail the test
		if !errorFound {
			t.Errorf("Unexpected error: %s", line)
		}
	}

	// Check that all expected errors were found
	for goldFilePath, lineErrors := range grd.errors {
		for goldLineNum, regexps := range lineErrors {
			for _, regexp := range regexps {
				key := fmt.Sprintf("%s:%d:%s", goldFilePath, goldLineNum, regexp.String())
				if !parsedErrors[key] {
					t.Errorf("Expected error not found in stderr: %s line %d: %s", goldFilePath, goldLineNum, regexp.String())
				}
			}
		}
	}
}

func filepathJoin(elem ...string) string {
	return filepath.ToSlash(filepath.Join(elem...))
}

// loadFileLines loads and splits the file content into lines
func loadFileLines(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Replace all Windows CRLF with LF and split content into lines
	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	// Split on LF
	lines := strings.Split(normalized, "\n")

	return lines, nil
}
