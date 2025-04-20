package hvgen_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/voedger/reqmd/internal/hvgen"
)

func TestHVGenerator(t *testing.T) {

	// NumReqSites        int
	// MaxSitesPerPackage int
	// MaxTagsPerSite     int
	// MaxSitesPerFile    int
	// MaxTagsPerFile     int
	// MaxTreeDepth       int
	// SrcToMdRatio       int

	testDir := filepath.Join(".testdata", "TestHVGenerator")
	// Remove testDir if it exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		err = os.RemoveAll(testDir)
		require.NoError(t, err)
	}

	cfg := hvgen.DefaultConfig(testDir)
	cfg.NumReqSites = 20
	cfg.MaxSitesPerPackage = 5
	cfg.MaxTagsPerSite = 2
	cfg.MaxSitesPerFile = 3
	cfg.MaxTagsPerFile = 3
	cfg.MaxTreeDepth = 2
	cfg.SrcToMdRatio = 5
	err := hvgen.HVGenerator(cfg)
	if err != nil {
		t.Fatalf("HVGenerator returned error: %v", err)
	}
}
