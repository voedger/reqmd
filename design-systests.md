# System Testing Plan for reqmd using Embedded Fixtures

## 1. Testing Infrastructure Setup

### Create Basic Test Structure

```go
package systest

// TestFixture represents a loaded test environment
type TestFixture struct {
	RootDir      string
	MarkdownsDir string
	SourcesDir   string
}

// TestFile represents a file to be created or modified in test fixtures
type TestFile struct {
	Path     string
	Content  string
	IsSource bool
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
	mdDir := filepath.Join(rootDir, "markdowns")
	srcDir := filepath.Join(rootDir, "sources")
	
	require.NoError(t, os.MkdirAll(mdDir, 0755))
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	
	// Initialize git repositories
	initGitRepo(t, mdDir)
	initGitRepo(t, srcDir)
	
	// Copy files from embedded filesystem to temp directory
	copyFixtureFiles(t, l.fsys, fixtureName, mdDir, srcDir)
	
	return TestFixture{
		RootDir:      rootDir,
		MarkdownsDir: mdDir,
		SourcesDir:   srcDir,
	}
}

// ModifyFixture loads a fixture and then applies additional files or modifications
func (l *FixtureLoader) ModifyFixture(t *testing.T, fixtureName string, modifications []TestFile) TestFixture {
	fixture := l.LoadFixture(t, fixtureName)
	
	// Apply modifications
	for _, mod := range modifications {
		dir := fixture.MarkdownsDir
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
func copyFixtureFiles(t *testing.T, fsys fs.FS, fixtureName string, mdDir, srcDir string) {
	// Implementation for copying files from embedded filesystem
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
│   ├── markdowns\
│   │   └── requirements.md
│   └── sources\
│       └── src\
│           └── impl.go
├── extensions\
│   ├── markdowns\
│   │   └── requirements.md
│   └── sources\
│       └── src\
│           ├── impl.go
│           └── impl.js
├── dry-run\
│   ├── markdowns\
│   │   └── requirements.md
│   └── sources\
│       └── src\
│           └── impl.go
├── verbose\
│   ├── markdowns\
│   │   └── requirements.md
│   └── sources\
│       └── src\
│           └── impl.go
├── errors\
│   ├── duplicate\
│   │   └── markdowns\
│   │       ├── requirements.md
│   │       └── other\
│   │           └── requirements.md
│   ├── missing-package\
│   │   └── markdowns\
│   │       └── requirements.md
│   └── syntax\
│       └── markdowns\
│           └── requirements.md
└── multi-source\
    ├── markdowns\
    │   └── requirements.md
    ├── backend\
    │   └── server.go
    └── frontend\
        └── client.ts
```

## 3. Test Implementation Plan

### 3.1 Basic Functionality Tests

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
		fixture.MarkdownsDir,
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Verify markdown was updated correctly
	updatedMd, err := os.ReadFile(filepath.Join(fixture.MarkdownsDir, "requirements.md"))
	require.NoError(t, err)
	
	// Perform assertions on the updated content
	mdContent := string(updatedMd)
	assert.Contains(t, mdContent, "`~REQ001~`covered[^1]✅")
	assert.Contains(t, mdContent, "`~REQ002~`covered[^2]✅")
	assert.Contains(t, mdContent, "[^1]: `[~com.example.basic/REQ001~impl]`")
	assert.Contains(t, mdContent, "[^2]: `[~com.example.basic/REQ002~test]`")
	
	// Verify reqmdfiles.json exists
	assert.FileExists(t, filepath.Join(fixture.MarkdownsDir, "reqmdfiles.json"))
}
```

### 3.2 Command Line Options Tests

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
		fixture.MarkdownsDir, 
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Verify only .go implementation was included
	updatedMd, err := os.ReadFile(filepath.Join(fixture.MarkdownsDir, "requirements.md"))
	require.NoError(t, err)
	
	mdContent := string(updatedMd)
	assert.Contains(t, mdContent, "impl.go")
	assert.NotContains(t, mdContent, "impl.js")
}

func TestDryRunOption(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "dry-run")
	
	// Get original content
	originalMd, err := os.ReadFile(filepath.Join(fixture.MarkdownsDir, "requirements.md"))
	require.NoError(t, err)
	
	// Run with dry-run flag
	cmd := exec.Command("reqmd", "trace", "--dry-run",
		fixture.MarkdownsDir,
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Verify file was not modified
	afterMd, err := os.ReadFile(filepath.Join(fixture.MarkdownsDir, "requirements.md"))
	require.NoError(t, err)
	
	assert.Equal(t, string(originalMd), string(afterMd), "File should not be modified in dry run")
	assert.NoFileExists(t, filepath.Join(fixture.MarkdownsDir, "reqmdfiles.json"))
}

func TestVerboseOption(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "verbose")
	
	// Run with verbose flag
	cmd := exec.Command("reqmd", "trace", "-v",
		fixture.MarkdownsDir,
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Check verbose output
	outputStr := string(output)
	assert.Contains(t, outputStr, "Scanning")
	assert.Contains(t, outputStr, "Processing")
	// Check for other expected verbose messages
}
```

### 3.3 Error Handling Tests

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
	cmd := exec.Command("reqmd", "trace", fixture.MarkdownsDir)
	output, err := cmd.CombinedOutput()
	
	// Verify error behavior
	assert.Error(t, err, "Command should fail with duplicate requirements")
	assert.Contains(t, string(output), "duplicate")
	assert.Contains(t, string(output), "REQ001")
}

func TestMissingPackageError(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "errors/missing-package")
	
	// Run command expecting error
	cmd := exec.Command("reqmd", "trace", fixture.MarkdownsDir)
	output, err := cmd.CombinedOutput()
	
	assert.Error(t, err)
	assert.Contains(t, string(output), "shall define reqmd.package")
}

func TestSyntaxError(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "errors/syntax")
	
	cmd := exec.Command("reqmd", "trace", fixture.MarkdownsDir)
	output, err := cmd.CombinedOutput()
	
	assert.Error(t, err)
	// Assert specific syntax error messages
}
```

### 3.4 Complex Scenario Tests

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

func TestMultipleSourceDirectories(t *testing.T) {
	loader := NewFixtureLoader()
	fixture := loader.LoadFixture(t, "multi-source")
	
	// Extract backend and frontend paths from the fixture
	backendDir := filepath.Join(fixture.RootDir, "backend")
	frontendDir := filepath.Join(fixture.RootDir, "frontend")
	
	// Run reqmd with multiple source paths
	cmd := exec.Command("reqmd", "trace",
		fixture.MarkdownsDir,
		backendDir, frontendDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Verify both implementations are referenced
	updatedMd, err := os.ReadFile(filepath.Join(fixture.MarkdownsDir, "requirements.md"))
	require.NoError(t, err)
	
	mdContent := string(updatedMd)
	assert.Contains(t, mdContent, "server.go")
	assert.Contains(t, mdContent, "client.ts")
}

func TestDynamicFixtureModification(t *testing.T) {
	// Test using the ModifyFixture method to demonstrate hybrid approach
	loader := NewFixtureLoader()
	fixture := loader.ModifyFixture(t, "basic", []TestFile{
		{
			Path: "requirements.md",
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
		fixture.MarkdownsDir,
		fixture.SourcesDir)
	
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "reqmd trace command failed: %s", output)
	
	// Verify results on the modified content
	updatedMd, err := os.ReadFile(filepath.Join(fixture.MarkdownsDir, "requirements.md"))
	require.NoError(t, err)
	
	mdContent := string(updatedMd)
	assert.Contains(t, mdContent, "`~MOD001~`covered")
	assert.Contains(t, mdContent, "`~MOD002~`covered")
}
```

### 3.5 Performance Tests

```go
package systest

import (
	"fmt"
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
	baseFixture := loader.LoadFixture(t, "basic")
	
	// Generate many additional requirements and implementations
	generateLargeTestData(t, baseFixture, 50, 200)
	
	startTime := time.Now()
	
	// Run reqmd on the large dataset
	cmd := exec.Command("reqmd", "trace",
		baseFixture.MarkdownsDir,
		baseFixture.SourcesDir)
	
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

## 4. Main Test Setup

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

## 5. Test Execution Strategy

1. **Development Process**:
   - Create all fixture files first
   - Implement fixture loading infrastructure
   - Add one test category at a time, starting with basic functionality
   - Add more complex tests once basic functionality is verified

2. **Running Tests**:
   - Standard tests: `go test ./systest`
   - Verbose mode: `go test -v ./systest`
   - Skip performance tests: `go test -short ./systest`
   - Test specific category: `go test ./systest -run TestBasic`

3. **CI Integration**:
   - Add system tests to CI pipeline after unit tests
   - Configure CI to build binary before running system tests
   - Archive test artifacts on CI for debugging failures

This approach provides:

- Clear separation between test data and test logic
- Realistic test scenarios using real input files
- Good maintainability with the ability to easily modify fixtures
- Comprehensive coverage of the reqmd tool's functionality