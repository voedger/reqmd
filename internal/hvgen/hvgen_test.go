package hvgen_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/voedger/reqmd/internal/hvgen"
)

func TestHVGenerator(t *testing.T) {

	testDir := filepath.Join(".testdata", "TestHVGenerator")

	err := os.RemoveAll(testDir)
	require.NoError(t, err)

	// NumReqSites        int
	// MaxSitesPerPackage int
	// MaxTagsPerSite     int
	// MaxSitesPerFile    int
	// MaxTagsPerFile     int
	// MaxTreeDepth       int
	// SrcToMdRatio       int

	cfg := hvgen.DefaultConfig(testDir)
	cfg.NumReqSites = 500
	cfg.MaxSitesPerPackage = 10
	cfg.MaxSitesPerFile = 4
	cfg.MaxTagsPerSite = 10
	cfg.MaxTagsPerFile = 4
	cfg.MaxTreeDepth = 2
	err = hvgen.HVGenerator(cfg)
	require.NoError(t, err)

	err = createGitRepo(testDir)
	require.NoError(t, err)

}

// Create a git repo in testDir and commit all files
// - branch: `main“
// - origin: `github.com/voedger/example`
func createGitRepo(testDir string) error {
	// Initialize git repository
	cmd := exec.Command("git", "init", testDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user for the repository
	cmd = exec.Command("git", "-C", testDir, "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git user.name: %w", err)
	}

	cmd = exec.Command("git", "-C", testDir, "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git user.email: %w", err)
	}

	// Set the origin URL
	cmd = exec.Command("git", "-C", testDir, "remote", "add", "origin", "https://github.com/voedger/example")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git origin: %w", err)
	}

	// Add all files
	cmd = exec.Command("git", "-C", testDir, "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// Commit all files
	cmd = exec.Command("git", "-C", testDir, "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit files to git: %w", err)
	}

	return nil
}
