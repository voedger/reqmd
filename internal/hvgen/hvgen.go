// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package hvgen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Coverer represents a source file that covers a requirement
type Coverer struct {
	RelPath      string // path relative to the repo root
	Package      string
	ReqSite      string
	Line         int
	CoverageType string // "impl", "test", etc.
}

// ReqFileDescr describes a requirements markdown file
type ReqFileDescr struct {
	RelDir   string // dir relative to the repo root
	Name     string
	Package  string
	ReqSites []string
	Coverers []Coverer
}

// SrcFileDescr describes a source code file
type SrcFileDescr struct {
	RelDir string // dir relative to the repo root
	Name   string
	Tags   []string
}

// ReqmdJson represents the structure of the reqmd.json file
type ReqmdJson struct {
	FileHashes map[string]string `json:"FileHashes"`
}

// HVGeneratorConfig defines configurable parameters for hvGenerator
type HVGeneratorConfig struct {
	// Number of markdown files to generate (1,000+)
	NumMarkdownFiles int
	// Number of source files to generate (10,000+)
	NumSourceFiles int
	// Number of requirements per markdown file (10-100)
	ReqsPerMarkdownFile int
	// Number of implementations per requirement (1-20)
	ImplsPerRequirement int
	// Base directory to create files in
	BaseDir string
	// Package ID prefix for requirements
	PackageIDPrefix string
	// Whether to generate golden files
	GenerateGoldenFiles bool
	// Simulated repository URL (for FileURL generation)
	RepoURL string
	// Simulated commit hash
	CommitHash string
	// Coverage types to generate
	CoverageTypes []string
}

// DefaultHVGeneratorConfig returns default configuration values for hvGenerator
func DefaultHVGeneratorConfig() *HVGeneratorConfig {
	return &HVGeneratorConfig{
		NumMarkdownFiles:    1000,
		NumSourceFiles:      10000,
		ReqsPerMarkdownFile: 50,
		ImplsPerRequirement: 10,
		BaseDir:             "hvtest",
		PackageIDPrefix:     "com.example",
		GenerateGoldenFiles: true,
		RepoURL:             "https://github.com/example/reqmd",
		CommitHash:          "abcdef1234567890abcdef1234567890abcdef12",
		CoverageTypes:       []string{"impl", "test", "doc"},
	}
}

// HVGenerator generates test files with configurable parameters for high-volume testing
func HVGenerator(config *HVGeneratorConfig) error {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(config.BaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Generate requirement file descriptors
	reqFileDescrs, err := generateReqFileDescrs(config)
	if err != nil {
		return fmt.Errorf("failed to generate requirement file descriptors: %w", err)
	}

	// Generate source file descriptors and link them to requirements
	srcFileDescrs, err := generateSrcFileDescrs(config, reqFileDescrs)
	if err != nil {
		return fmt.Errorf("failed to generate source file descriptors: %w", err)
	}

	// Generate actual files based on the descriptors
	if err := generateFiles(config, reqFileDescrs, srcFileDescrs); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	// Generate reqmd.json if golden files are requested
	if config.GenerateGoldenFiles {
		if err := generateReqmdJson(config, reqFileDescrs); err != nil {
			return fmt.Errorf("failed to generate reqmd.json: %w", err)
		}
	}

	return nil
}

// generateReqFileDescrs creates descriptors for requirement markdown files
func generateReqFileDescrs(config *HVGeneratorConfig) ([]ReqFileDescr, error) {
	result := make([]ReqFileDescr, config.NumMarkdownFiles)

	// Create subdirectories for better file organization
	numSubDirs := int(math.Sqrt(float64(config.NumMarkdownFiles)))
	if numSubDirs < 1 {
		numSubDirs = 1
	}

	for i := 0; i < config.NumMarkdownFiles; i++ {
		// Organize files into subdirectories
		subDirIndex := i % numSubDirs
		relDir := filepath.Join("req", fmt.Sprintf("subdir_%03d", subDirIndex))

		// Create package ID based on the file index
		packageID := fmt.Sprintf("%s.req.%04d", config.PackageIDPrefix, i)

		// Generate requirement sites
		reqSites := make([]string, config.ReqsPerMarkdownFile)
		for j := 0; j < config.ReqsPerMarkdownFile; j++ {
			reqID := fmt.Sprintf("%s.REQ-%04d-%03d", packageID, i, j)
			reqSites[j] = reqID
		}

		result[i] = ReqFileDescr{
			RelDir:   relDir,
			Name:     fmt.Sprintf("req_%04d.md", i),
			Package:  packageID,
			ReqSites: reqSites,
			Coverers: []Coverer{}, // Will be populated later when generating source files
		}
	}

	return result, nil
}

// generateSrcFileDescrs creates descriptors for source code files and links them to requirements
func generateSrcFileDescrs(config *HVGeneratorConfig, reqFileDescrs []ReqFileDescr) ([]SrcFileDescr, error) {
	result := make([]SrcFileDescr, config.NumSourceFiles)

	// Create subdirectories for better file organization
	numSubDirs := int(math.Sqrt(float64(config.NumSourceFiles)))
	if numSubDirs < 1 {
		numSubDirs = 1
	}

	// Calculate requirements distribution across source files
	// Each requirement should have approximately config.ImplsPerRequirement implementations
	totalReqs := config.NumMarkdownFiles * config.ReqsPerMarkdownFile
	implsPerSourceFile := (totalReqs * config.ImplsPerRequirement) / config.NumSourceFiles
	if implsPerSourceFile < 1 {
		implsPerSourceFile = 1
	}

	// Random source for more realistic distribution
	r := rand.New(rand.NewSource(42)) // Use fixed seed for reproducibility

	for i := 0; i < config.NumSourceFiles; i++ {
		// Organize files into subdirectories
		subDirIndex := i % numSubDirs
		relDir := filepath.Join("src", fmt.Sprintf("subdir_%03d", subDirIndex))

		// Generate tags for this source file
		tags := []string{}

		// Randomly select requirements to implement in this file
		for j := 0; j < implsPerSourceFile; j++ {
			// Select random requirement file and requirement
			reqFileIndex := r.Intn(len(reqFileDescrs))
			reqFile := &reqFileDescrs[reqFileIndex]

			if len(reqFile.ReqSites) == 0 {
				continue
			}

			reqIndex := r.Intn(len(reqFile.ReqSites))
			reqSite := reqFile.ReqSites[reqIndex]

			// Select random coverage type
			coverageType := config.CoverageTypes[r.Intn(len(config.CoverageTypes))]

			// Add tag in format REQ:requirement-id
			tag := fmt.Sprintf("REQ:%s", reqSite)
			tags = append(tags, tag)

			// Add this file as coverer for the requirement
			coverer := Coverer{
				RelPath:      filepath.Join(relDir, fmt.Sprintf("src_%04d.go", i)),
				Package:      fmt.Sprintf("%s.src.%04d", config.PackageIDPrefix, i),
				ReqSite:      reqSite,
				Line:         r.Intn(500) + 1, // Random line number between 1-500
				CoverageType: coverageType,
			}

			reqFile.Coverers = append(reqFile.Coverers, coverer)
		}

		result[i] = SrcFileDescr{
			RelDir: relDir,
			Name:   fmt.Sprintf("src_%04d.go", i),
			Tags:   tags,
		}
	}

	return result, nil
}

// generateFiles creates actual files based on the descriptors
func generateFiles(config *HVGeneratorConfig, reqFileDescrs []ReqFileDescr, srcFileDescrs []SrcFileDescr) error {
	// Generate requirement markdown files
	for _, reqFile := range reqFileDescrs {
		if err := generateReqFile(config, reqFile); err != nil {
			return err
		}
	}

	// Generate source code files
	for _, srcFile := range srcFileDescrs {
		if err := generateSrcFile(config, srcFile); err != nil {
			return err
		}
	}

	return nil
}

// generateReqFile creates a markdown file with requirements and coverage
func generateReqFile(config *HVGeneratorConfig, reqFile ReqFileDescr) error {
	// Create directory if it doesn't exist
	dirPath := filepath.Join(config.BaseDir, reqFile.RelDir)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	// Prepare the file content
	var content strings.Builder

	// Add header
	content.WriteString(fmt.Sprintf("# %s\n\n", reqFile.Package))
	content.WriteString("## Requirements\n\n")

	// Add requirements with IDs
	for i, reqSite := range reqFile.ReqSites {
		content.WriteString(fmt.Sprintf("### %s\n\n", reqSite))
		content.WriteString(fmt.Sprintf("This is requirement %d in file %s\n\n", i+1, reqFile.Name))

		// Add some random content for realism
		paragraphCount := rand.Intn(3) + 1
		for p := 0; p < paragraphCount; p++ {
			content.WriteString(generateRandomParagraph())
			content.WriteString("\n\n")
		}
	}

	// Add coverage footnotes
	coverageByReq := make(map[string][]Coverer)
	for _, coverer := range reqFile.Coverers {
		coverageByReq[coverer.ReqSite] = append(coverageByReq[coverer.ReqSite], coverer)
	}

	// Only add footnotes section if we have coverage
	if len(reqFile.Coverers) > 0 {
		content.WriteString("\n")

		// Add footnotes for each requirement that has coverage
		for reqSite, coverers := range coverageByReq {
			footnote := fmt.Sprintf("[%s]: ", reqSite)

			for i, coverer := range coverers {
				if i > 0 {
					footnote += ", "
				}

				fileURL := fmt.Sprintf("%s/%s#L%d", config.RepoURL, coverer.RelPath, coverer.Line)
				footnote += fmt.Sprintf("%s:%s", coverer.CoverageType, fileURL)
			}

			content.WriteString(footnote + "\n")
		}
	}

	// Write the file
	filePath := filepath.Join(dirPath, reqFile.Name)
	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

// generateSrcFile creates a source code file with coverage tags
func generateSrcFile(config *HVGeneratorConfig, srcFile SrcFileDescr) error {
	// Create directory if it doesn't exist
	dirPath := filepath.Join(config.BaseDir, srcFile.RelDir)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	// Prepare the file content
	var content strings.Builder

	// Add header
	content.WriteString(fmt.Sprintf("// Auto-generated file: %s\n", srcFile.Name))
	content.WriteString(fmt.Sprintf("// Package %s\n", strings.ReplaceAll(srcFile.RelDir, "/", "_")))
	content.WriteString("package main\n\n")

	// Add some random functions
	funcCount := rand.Intn(5) + 3

	for i := 0; i < funcCount; i++ {
		funcName := fmt.Sprintf("Function%d", i+1)
		content.WriteString(fmt.Sprintf("func %s() {\n", funcName))

		// Add tags as comments for some functions
		if i < len(srcFile.Tags) {
			content.WriteString(fmt.Sprintf("\t// %s\n", srcFile.Tags[i]))
		}

		// Add some random content
		lineCount := rand.Intn(10) + 5
		for j := 0; j < lineCount; j++ {
			content.WriteString("\t// Some code here\n")
		}

		content.WriteString("}\n\n")
	}

	// Write the file
	filePath := filepath.Join(dirPath, srcFile.Name)
	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

// generateReqmdJson creates the reqmd.json file for tracking file hashes
func generateReqmdJson(config *HVGeneratorConfig, reqFileDescrs []ReqFileDescr) error {
	// Gather all files that need to be tracked
	fileHashes := make(map[string]string)

	// Create reqmd.json in each requirement directory
	reqDirs := make(map[string]bool)
	for _, reqFile := range reqFileDescrs {
		dirPath := filepath.Join(config.BaseDir, reqFile.RelDir)
		reqDirs[dirPath] = true

		// Add markdown file hash
		filePath := filepath.Join(dirPath, reqFile.Name)
		hash, err := calculateFileHash(filePath)
		if err != nil {
			return err
		}
		relativePath := filepath.Join(reqFile.RelDir, reqFile.Name)
		fileHashes[relativePath] = hash
	}

	// Create reqmd.json in each directory
	for dirPath := range reqDirs {
		reqmdJson := ReqmdJson{
			FileHashes: fileHashes,
		}

		jsonData, err := json.MarshalIndent(reqmdJson, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal reqmd.json: %w", err)
		}

		reqmdPath := filepath.Join(dirPath, "reqmd.json")
		if err := os.WriteFile(reqmdPath, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write reqmd.json: %w", err)
		}
	}

	return nil
}

// calculateFileHash computes SHA256 hash for a file
func calculateFileHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// generateRandomParagraph creates random text for requirements
func generateRandomParagraph() string {
	sentences := []string{
		"The system shall process input data according to the specification.",
		"All error conditions must be properly handled and reported.",
		"Performance requirements specify that response time should not exceed 100ms.",
		"Authentication must be performed using the approved security protocols.",
		"The component shall maintain backward compatibility with previous versions.",
		"Data persistence must ensure that no information is lost during system failures.",
		"User interface elements should follow the design system guidelines.",
		"The API must provide appropriate error codes and messages.",
		"Configuration options should be externalized for environment-specific settings.",
		"All communications must be encrypted using TLS 1.3 or later.",
		"Logging functionality should capture relevant diagnostic information.",
		"The system must handle concurrent requests without data corruption.",
		"Memory usage should not exceed predefined thresholds during normal operation.",
		"Internationalization support is required for all user-facing content.",
		"The application must gracefully degrade when services are unavailable.",
		"Regular automated backups should be performed according to the schedule.",
		"The system must notify administrators of critical failures.",
		"Documentation should be maintained with each code update.",
		"All public interfaces must be thoroughly tested.",
		"User data must be anonymized before being used for analytics.",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sentenceCount := (r.Intn(4) + 2) * 2 // Double the 2-5 sentences to 4-10 sentences

	var paragraph strings.Builder
	for i := 0; i < sentenceCount; i++ {
		paragraph.WriteString(sentences[r.Intn(len(sentences))])
		paragraph.WriteString(" ")
	}

	return strings.TrimSpace(paragraph.String())
}
