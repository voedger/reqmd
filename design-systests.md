# System Testing Plan for reqmd using Embedded Fixtures

## 1. Testing Infrastructure Setup

### Create Basic Test Structure

```go
package systest

// TestFixture represents a loaded test environment
type TestFixture struct {
	RootDir         string
	RequirementsDir string
	SourcesDir      string
	GoldenDir       string
}

// TestFile represents a file to be created or modified in test fixtures
type TestFile struct {
	Path     string
	Content  string
	IsSource bool // false means requirements file, true means source file
}
```

### Create Fixture Loader Using Embedded Folders

```go
package systest

import (
	"embed"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed fixtures/*
var fixturesFS embed.FS

// FixtureLoader loads fixtures from the embedded filesystem
type FixtureLoader struct {
	fsys fs.FS
}

// NewFixtureLoader creates a new fixture loader
func NewFixtureLoader() *FixtureLoader {
	return &FixtureLoader{fsys: fixturesFS}
}

// LoadFixture loads a specific fixture directory to a temporary location
func (l *FixtureLoader) LoadFixture(t *testing.T, fixtureName string) TestFixture {
	rootDir := t.TempDir()
	
	// Create standard directory structure
	reqDir := filepath.Join(rootDir, "requirements")
	srcDir := filepath.Join(rootDir, "sources")
	goldenDir := filepath.Join(rootDir, "golden")
	
	require.NoError(t, os.MkdirAll(reqDir, 0755))
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	require.NoError(t, os.MkdirAll(goldenDir, 0755))
	
	// Initialize git repositories
	initGitRepo(t, reqDir)
	initGitRepo(t, srcDir)
	
	// Copy files from embedded filesystem to temp directory
	copyFixtureFiles(t, l.fsys, fixtureName, reqDir, srcDir, goldenDir)
	
	return TestFixture{
		RootDir:         rootDir,
		RequirementsDir: reqDir,
		SourcesDir:      srcDir,
		GoldenDir:       goldenDir,
	}
}

// ModifyFixture loads a fixture and then applies additional files or modifications
func (l *FixtureLoader) ModifyFixture(t *testing.T, fixtureName string, modifications []TestFile) TestFixture {
	fixture := l.LoadFixture(t, fixtureName)
	
	// Apply modifications
	for _, mod := range modifications {
		dir := fixture.RequirementsDir
		if mod.IsSource {
			dir = fixture.SourcesDir
		}
		
		fullPath := filepath.Join(dir, mod.Path)
		dirPath := filepath.Dir(fullPath)
		
		require.NoError(t, os.MkdirAll(dirPath, 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(mod.Content), 0644))
		
		// Add to git for source files
		if mod.IsSource {
			addToGit(t, fullPath)
		}
	}
	
	return fixture
}

// Helper functions
func copyFixtureFiles(t *testing.T, fsys fs.FS, fixtureName string, reqDir, srcDir, goldenDir string) {
	// Implementation for copying files from embedded filesystem to the appropriate directories
}

func initGitRepo(t *testing.T, dir string) {
	// Implementation for git initialization
}

func addToGit(t *testing.T, filePath string) {
	// Implementation for adding files to git
}
```

## 2. Create Fixture Directory Structure

```text
c:\workspaces\work\reqmd\systest\fixtures\
├── basic\
│   ├── requirements\
│   │   └── requirements.md
│   ├── sources\
│   │   └── src\
│   │       └── impl.go
│   └── golden\
│       ├── requirements.md
│       └── reqmdfiles.json
├── extensions\
│   ├── requirements\
│   │   └── requirements.md
│   ├── sources\
│   │   └── src\
│   │       ├── impl.go
│   │       └── impl.js
│   └── golden\
│       └── requirements.md
├── dry-run\
│   ├── requirements\
│   │   └── requirements.md
│   ├── sources\
│   │   └── src\
│   │       └── impl.go
│   └── golden\
│       └── output.txt
├── verbose\
│   ├── requirements\
│   │   └── requirements.md
│   ├── sources\
│   │   └── src\
│   │       └── impl.go
│   └── golden\
│       └── output.txt
├── errors\
│   ├── duplicate\
│   │   ├── requirements\
│   │   │   ├── requirements.md
│   │   │   └── other\
│   │   │       └── requirements.md
│   │   └── golden\
│   │       └── error.txt
│   ├── missing-package\
│   │   ├── requirements\
│   │   │   └── requirements.md
│   │   └── golden\
│   │       └── error.txt
│   └── syntax\
│       ├── requirements\
│       │   └── requirements.md
│       └── golden\
│           └── error.txt
└── multi-source\
    ├── requirements\
    │   └── requirements.md
    ├── sources\
    │   ├── backend\
    │   │   └── server.go
    │   └── frontend\
    │       └── client.ts
    └── golden\
        └── requirements.md
```

## 3. Reference File Comparison Helpers

```go
package systest

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var updateReference = flag.Bool("update", false, "Update reference files instead of comparing")

// CompareWithReference compares actual output with reference data
func CompareWithReference(t *testing.T, actual []byte, fixture TestFixture, relativePath string) {
	t.Helper()
	
	referencePath := filepath.Join(fixture.GoldenDir, relativePath)
	
	if *updateReference {
		// Update reference data
		err := os.MkdirAll(filepath.Dir(referencePath), 0755)
		require.NoError(t, err, "Failed to create reference directory")
		
		err = os.WriteFile(referencePath, actual, 0644)
		require.NoError(t, err, "Failed to write reference file")
		return
	}
	
	// Compare with reference data
	expected, err := os.ReadFile(referencePath)
	require.NoError(t, err, "Failed to read reference file: %s", referencePath)
	
	require.Equal(t, string(expected), string(actual), 
		"Output didn't match reference data in %s", referencePath)
}
```

## 4. Test Implementation Plan

### 4.1 Basic Functionality Tests

```go
package systest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicTracing(t *testing.T) {
	// Load fixture from embedded filesystem
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "basic")
	
	// Run reqmd command
	cmd := exec.Command("reqmd", "trace", 
		fixture.RequirementsDir,
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Compare command output with reference data
	CompareWithReference(t, output, fixture, "output.txt")
	
	// Verify markdown was updated correctly
	updatedMd, err := os.ReadFile(filepath.Join(fixture.RequirementsDir, "requirements.md"))
	require.NoError(t, err)
	
	// Compare the updated markdown with reference file
	CompareWithReference(t, updatedMd, fixture, "requirements.md")
	
	// Verify reqmdfiles.json exists and matches reference
	jsonData, err := os.ReadFile(filepath.Join(fixture.RequirementsDir, "reqmdfiles.json"))
	require.NoError(t, err)
	CompareWithReference(t, jsonData, fixture, "reqmdfiles.json")
	
	// Also perform basic file existence check for clarity
	assert.FileExists(t, filepath.Join(fixture.RequirementsDir, "reqmdfiles.json"))
}
```

### 4.2 Command Line Options Tests

```go
package systest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionsOption(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "extensions")
	
	// Run with only .go extension
	cmd := exec.Command("reqmd", "trace", "-e", ".go",
		fixture.RequirementsDir, 
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Verify results against reference file
	updatedMd, err := os.ReadFile(filepath.Join(fixture.RequirementsDir, "requirements.md"))
	require.NoError(t, err)
	CompareWithReference(t, updatedMd, fixture, "requirements.md")
}

func TestDryRunOption(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "dry-run")
	
	// Get original content
	originalMd, err := os.ReadFile(filepath.Join(fixture.RequirementsDir, "requirements.md"))
	require.NoError(t, err)
	
	// Run with dry-run flag
	cmd := exec.Command("reqmd", "trace", "--dry-run",
		fixture.RequirementsDir,
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Compare output with reference
	CompareWithReference(t, output, fixture, "output.txt")
	
	// Verify file was not modified
	afterMd, err := os.ReadFile(filepath.Join(fixture.RequirementsDir, "requirements.md"))
	require.NoError(t, err)
	
	assert.Equal(t, string(originalMd), string(afterMd), "File should not be modified in dry run")
	assert.NoFileExists(t, filepath.Join(fixture.RequirementsDir, "reqmdfiles.json"))
}

func TestVerboseOption(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "verbose")
	
	// Run with verbose flag
	cmd := exec.Command("reqmd", "trace", "-v",
		fixture.RequirementsDir,
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Compare verbose output with reference file
	CompareWithReference(t, output, fixture, "output.txt")
}
```

### 4.3 Error Handling Tests

```go
package systest

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuplicateRequirementError(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "errors/duplicate")
	
	// Run command expecting error
	cmd := exec.Command("reqmd", "trace", fixture.RequirementsDir)
	output, err := cmd.CombinedOutput()
	
	// Verify error behavior
	assert.Error(t, err, "Command should fail with duplicate requirements")
	
	// Compare error output with reference
	CompareWithReference(t, output, fixture, "error.txt")
}

func TestMissingPackageError(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "errors/missing-package")
	
	// Run command expecting error
	cmd := exec.Command("reqmd", "trace", fixture.RequirementsDir)
	output, err := cmd.CombinedOutput()
	
	assert.Error(t, err)
	CompareWithReference(t, output, fixture, "error.txt")
}

func TestSyntaxError(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "errors/syntax")
	
	cmd := exec.Command("reqmd", "trace", fixture.RequirementsDir)
	output, err := cmd.CombinedOutput()
	
	assert.Error(t, err)
	CompareWithReference(t, output, fixture, "error.txt")
}
```

### 4.4 Complex Scenario Tests

```go
package systest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultipleSourceDirectories(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "multi-source")
	
	// Extract backend and frontend paths
	backendDir := filepath.Join(fixture.SourcesDir, "backend")
	frontendDir := filepath.Join(fixture.SourcesDir, "frontend")
	
	// Run reqmd with multiple source paths
	cmd := exec.Command("reqmd", "trace",
		fixture.RequirementsDir,
		backendDir, frontendDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Compare results with reference data
	updatedMd, err := os.ReadFile(filepath.Join(fixture.RequirementsDir, "requirements.md"))
	require.NoError(t, err)
	CompareWithReference(t, updatedMd, fixture, "requirements.md")
}

func TestDynamicFixtureModification(t *testing.T) {
	// Test using the ModifyFixture method to demonstrate hybrid approach
	loader := NewFixtureLoader()
	fixture := loader.ModifyFixture(t, "basic", []TestFile{
		{
			Path: "modified.md",
			Content: `# Modified Requirements
reqmd.package: com.example.modified

## Feature One
\`~MOD001~\`

## Feature Two
\`~MOD002~\`
`,
			IsSource: false,
		},
		{
			Path: "src/modified_impl.go",
			Content: `package impl

// [~com.example.modified/MOD001~impl]
func FeatureOne() {}

// [~com.example.modified/MOD002~test]
func TestFeatureTwo() {}
`,
			IsSource: true,
		},
	})
	
	// Run reqmd on the modified fixture
	cmd := exec.Command("reqmd", "trace",
		fixture.RequirementsDir,
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Verify results on the dynamically created file
	updatedMd, err := os.ReadFile(filepath.Join(fixture.RequirementsDir, "modified.md"))
	require.NoError(t, err)
	
	// For dynamic tests without a pre-defined reference file,
	// we can still make specific assertions
	assert.Contains(t, string(updatedMd), "`~MOD001~`covered")
	assert.Contains(t, string(updatedMd), "`~MOD002~`covered")
}
```

### 4.5 Performance Tests

```go
package systest

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLargeScaleProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	// Create a large fixture programmatically
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "basic")
	
	// Generate many additional requirements and implementations
	generateLargeTestData(t, fixture, 50, 200)
	
	startTime := time.Now()
	
	// Run reqmd on the large dataset
	cmd := exec.Command("reqmd", "trace",
		fixture.RequirementsDir,
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	duration := time.Since(startTime)
	
	// Log performance metrics
	t.Logf("Processing time: %v", duration)
	assert.Less(t, duration, 30*time.Second, "Processing should complete within reasonable time")
}

// Helper to generate large test data
func generateLargeTestData(t *testing.T, fixture TestFixture, reqCount, fileCount int) {
	// Implementation for generating large test datasets
}
```

## 5. Main Test Setup

```go
package systest

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var reqmdBinary string

func TestMain(m *testing.M) {
	flag.StringVar(&reqmdBinary, "binary", "", "Path to reqmd binary to test")
	flag.BoolVar(&updateReference, "update", false, "Update reference files instead of comparing")
	flag.Parse()

	// Build binary if not specified
	if reqmdBinary == "" {
		tempBin, err := buildReqmdBinary()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to build reqmd binary: %v\n", err)
			os.Exit(1)
		}
		reqmdBinary = tempBin
		defer os.Remove(reqmdBinary)
	}

	// Replace exec.Command for reqmd calls
	origExecCommand := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "reqmd" {
			return origExecCommand(reqmdBinary, args...)
		}
		return origExecCommand(name, args...)
	}
	defer func() { execCommand = origExecCommand }()

	// Run tests
	os.Exit(m.Run())
}

// Helper to build reqmd binary for testing
func buildReqmdBinary() (string, error) {
	tempFile := filepath.Join(os.TempDir(), "reqmd")
	if runtime.GOOS == "windows" {
		tempFile += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", tempFile)
	cmd.Dir = ".." // Adjust if needed to point to project root
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return tempFile, nil
}

// Allow mocking exec.Command
var execCommand = exec.Command
```

## 6. Test Execution Strategy

1. **Development Process**:
   - Create all fixture files first with proper directory structure
   - Create reference output files (can be empty initially)
   - Implement fixture loading infrastructure
   - Run tests with `-update` flag to populate reference files
   - Review reference files for correctness
   - Add more complex tests once basic functionality is verified

2. **Running Tests**:
   - Standard tests: `go test ./systest`
   - Update reference files: `go test ./systest -update`
   - Verbose mode: `go test -v ./systest`
   - Skip performance tests: `go test -short ./systest`
   - Test specific category: `go test ./systest -run TestBasic`

3. **CI Integration**:
   - Add system tests to CI pipeline after unit tests
   - Configure CI to build binary before running system tests
   - Archive test artifacts on CI for debugging failures

This approach provides:

- Clear separation between test data, reference outputs, and test logic
- Realistic test scenarios using real input files
- Good maintainability with the ability to easily update reference files
- Comprehensive coverage of the reqmd tool's functionality
- Clear organization with three distinct directory types for each test case