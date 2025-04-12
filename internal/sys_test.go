// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/voedger/reqmd/internal/hvgen"
	"github.com/voedger/reqmd/internal/systrun"
)

var sysTestsDir = filepath.Join("testdata", "systest")

func Test_systest_noreqs(t *testing.T) {
	runSysTest(t, "noreqs")
}

func Test_systest_errors(t *testing.T) {
	runSysTest(t, "errors")
}

func Test_systest_justreqs(t *testing.T) {
	runSysTest(t, "justreqs")
}

// Reqs and srcs in different folders
func Test_systest_req_src(t *testing.T) {
	runSysTest(t, "req_src")
}

// Requirements and sources in the same folder
func Test_systest_reqsrc(t *testing.T) {
	runSysTest(t, "reqsrc")
}

func runSysTest(t *testing.T, testID string) {
	systrun.RunSysTest(t, sysTestsDir, testID, ExecRootCmd, Version)
}

// Test with high-volume data generation
func Test_systest_highvolume(t *testing.T) {
	// Skip in normal tests due to volume
	if testing.Short() {
		t.Skip("Skipping high-volume test in short mode")
	}

	// Create high-volume test directory
	testDir := filepath.Join(sysTestsDir, "highvolume")

	// Remove test directory if it exists
	if _, err := os.Stat(testDir); err == nil {
		if err := os.RemoveAll(testDir); err != nil {
			t.Fatalf("Failed to remove existing test directory: %v", err)
		}
	}

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create high-volume test directory: %v", err)
	}

	// Generate test files
	config := hvgen.HVGeneratorConfig{
		// NumMarkdownFiles:    100,  // Reduced for test speed, use 1000+ for real load testing
		// NumSourceFiles:      1000, // Reduced for test speed, use 10000+ for real load testing
		NumMarkdownFiles:    2, // Reduced for test speed, use 1000+ for real load testing
		NumSourceFiles:      2, // Reduced for test speed, use 10000+ for real load testing
		ReqsPerMarkdownFile: 20,
		ImplsPerRequirement: 5,
		BaseDir:             testDir,
		PackageIDPrefix:     "com.example.hv",
	}

	if err := hvgen.HVGenerator(&config); err != nil {
		t.Fatalf("Failed to generate high-volume test data: %v", err)
	}

	// Run the system test
	runSysTest(t, "highvolume")
}
