// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package hvgen

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
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
	if config == nil {
		config = DefaultHVGeneratorConfig()
	}

	// Validate config
	if config.NumMarkdownFiles < 1 {
		return fmt.Errorf("number of markdown files must be at least 1")
	}
	if config.NumSourceFiles < 1 {
		return fmt.Errorf("number of source files must be at least 1")
	}
	if config.ReqsPerMarkdownFile < 1 {
		return fmt.Errorf("requirements per markdown file must be at least 1")
	}
	if config.ImplsPerRequirement < 1 {
		return fmt.Errorf("implementations per requirement must be at least 1")
	}
	if len(config.CoverageTypes) == 0 {
		return fmt.Errorf("at least one coverage type must be specified")
	}

	// Initialize random number generator (Go 1.20+ approach)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate requirement file descriptions
	reqFileDescrs := generateReqFileDescrs(config, rng)

	// Generate source file descriptions
	srcFileDescrs := generateSrcFileDescrs(config, reqFileDescrs, rng, config.CoverageTypes)

	// Generate actual files
	if err := generateFiles(config, reqFileDescrs, srcFileDescrs, config.GenerateGoldenFiles); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	return nil
}

// generateReqFileDescrs creates requirement file descriptors according to parameters
func generateReqFileDescrs(config *HVGeneratorConfig, _ *rand.Rand) []ReqFileDescr {
	reqFileDescrs := make([]ReqFileDescr, config.NumMarkdownFiles)

	for i := 0; i < config.NumMarkdownFiles; i++ {
		packageName := fmt.Sprintf("%s.module%d", config.PackageIDPrefix, i+1)
		filename := fmt.Sprintf("req_%d.md", i+1)

		reqSites := make([]string, config.ReqsPerMarkdownFile)
		for j := 0; j < config.ReqsPerMarkdownFile; j++ {
			reqSites[j] = fmt.Sprintf("REQ%03d", j+1)
		}

		reqFileDescrs[i] = ReqFileDescr{
			RelDir:   filepath.Join("req"),
			Name:     filename,
			Package:  packageName,
			ReqSites: reqSites,
			Coverers: []Coverer{}, // Will be populated later
		}
	}

	return reqFileDescrs
}

// generateSrcFileDescrs creates source file descriptors and links them to requirements
func generateSrcFileDescrs(config *HVGeneratorConfig, reqFileDescrs []ReqFileDescr, rng *rand.Rand, coverageTypes []string) []SrcFileDescr {
	srcFileDescrs := make([]SrcFileDescr, config.NumSourceFiles)

	// Create source file descriptors
	for i := 0; i < config.NumSourceFiles; i++ {
		subdir := fmt.Sprintf("pkg%d", i/1000)
		filename := fmt.Sprintf("file_%d.go", i+1)

		tags := []string{
			fmt.Sprintf("tag%d", rng.Intn(20)+1),
			fmt.Sprintf("tag%d", rng.Intn(20)+21),
		}

		srcFileDescrs[i] = SrcFileDescr{
			RelDir: filepath.Join("src", subdir),
			Name:   filename,
			Tags:   tags,
		}
	}

	// For each requirement, distribute implementations across source files
	for reqFileIdx, reqFile := range reqFileDescrs {
		for _, reqSite := range reqFile.ReqSites {
			// Determine how many implementations this requirement will have
			numImpls := rng.Intn(config.ImplsPerRequirement) + 1

			// Create implementations in random source files
			for i := 0; i < numImpls; i++ {
				srcFileIdx := rng.Intn(len(srcFileDescrs))
				srcFile := srcFileDescrs[srcFileIdx]
				coverageType := coverageTypes[rng.Intn(len(coverageTypes))]

				// Create a coverer and add it to the requirement
				coverer := Coverer{
					RelPath:      filepath.Join(srcFile.RelDir, srcFile.Name),
					Package:      filepath.Base(srcFile.RelDir),
					ReqSite:      reqSite,
					Line:         rng.Intn(100) + 10, // Random line number between 10-110
					CoverageType: coverageType,
				}

				reqFileDescrs[reqFileIdx].Coverers = append(reqFileDescrs[reqFileIdx].Coverers, coverer)
			}
		}
	}

	return srcFileDescrs
}

// generateFiles creates actual files based on the descriptors
func generateFiles(config *HVGeneratorConfig, reqFileDescrs []ReqFileDescr, srcFileDescrs []SrcFileDescr, generateGolden bool) error {
	// Create base directory structure
	if err := os.MkdirAll(config.BaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Generate markdown files and reqmd.json files
	reqmdJsonMap := make(map[string]*ReqmdJson)

	for _, reqFile := range reqFileDescrs {
		dirPath := filepath.Join(config.BaseDir, reqFile.RelDir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}

		filePath := filepath.Join(dirPath, reqFile.Name)
		if err := generateMarkdownFile(config, filePath, reqFile); err != nil {
			return fmt.Errorf("failed to generate markdown file %s: %w", filePath, err)
		}

		// Add entries to reqmd.json for this directory
		if _, exists := reqmdJsonMap[dirPath]; !exists {
			reqmdJsonMap[dirPath] = &ReqmdJson{
				FileHashes: make(map[string]string),
			}
		}

		// Add file hashes for coverers of this requirement file
		for _, coverer := range reqFile.Coverers {
			fileURL := constructFileURL(config.RepoURL, config.CommitHash, coverer.RelPath)
			reqmdJsonMap[dirPath].FileHashes[fileURL] = config.CommitHash
		}
	}

	// Write reqmd.json files
	for dirPath, reqmdJson := range reqmdJsonMap {
		if err := generateReqmdJson(filepath.Join(dirPath, "reqmd.json"), reqmdJson); err != nil {
			return fmt.Errorf("failed to generate reqmd.json in %s: %w", dirPath, err)
		}
	}

	// Generate source files
	for _, srcFile := range srcFileDescrs {
		dirPath := filepath.Join(config.BaseDir, srcFile.RelDir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}

		filePath := filepath.Join(dirPath, srcFile.Name)

		// Find all requirements covered by this source file
		var coveringReqs []struct {
			Package      string
			ReqSite      string
			Line         int
			CoverageType string
		}

		for _, reqFile := range reqFileDescrs {
			for _, coverer := range reqFile.Coverers {
				relPath := filepath.Join(srcFile.RelDir, srcFile.Name)
				if coverer.RelPath == relPath {
					coveringReqs = append(coveringReqs, struct {
						Package      string
						ReqSite      string
						Line         int
						CoverageType string
					}{
						Package:      reqFile.Package,
						ReqSite:      coverer.ReqSite,
						Line:         coverer.Line,
						CoverageType: coverer.CoverageType,
					})
				}
			}
		}

		if err := generateSourceFile(filePath, srcFile, coveringReqs); err != nil {
			return fmt.Errorf("failed to generate source file %s: %w", filePath, err)
		}
	}

	// Generate golden files if requested
	if generateGolden {
		if err := generateGoldenFiles(config.BaseDir, reqFileDescrs, srcFileDescrs, config); err != nil {
			return fmt.Errorf("failed to generate golden files: %w", err)
		}
	}

	return nil
}

// constructFileURL creates a FileURL according to the design
func constructFileURL(repoURL string, commitHash string, relPath string) string {
	if strings.Contains(repoURL, "github.com") {
		return fmt.Sprintf("%s/blob/%s/%s", repoURL, commitHash, relPath)
	} else if strings.Contains(repoURL, "gitlab.com") {
		return fmt.Sprintf("%s/-/blob/%s/%s", repoURL, commitHash, relPath)
	}
	// Default to GitHub format
	return fmt.Sprintf("%s/blob/%s/%s", repoURL, commitHash, relPath)
}

// generateMarkdownFile creates a markdown file with requirements following the design format
func generateMarkdownFile(config *HVGeneratorConfig, filePath string, reqFile ReqFileDescr) error {
	var content strings.Builder

	// Add header
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("reqmd.package: %s\n", reqFile.Package))
	content.WriteString("---\n\n")
	content.WriteString(fmt.Sprintf("# Module %s\n\n", reqFile.Package))
	content.WriteString("## Requirements\n\n")

	// Add requirements
	for _, reqSite := range reqFile.ReqSites {
		content.WriteString(fmt.Sprintf("### %s\n\n", reqSite))

		// Find coverers for this requirement
		var coverersForReq []Coverer
		for _, coverer := range reqFile.Coverers {
			if coverer.ReqSite == reqSite {
				coverersForReq = append(coverersForReq, coverer)
			}
		}

		// Determine if we have coverage
		hasCoverage := len(coverersForReq) > 0

		if hasCoverage {
			// Covered requirement
			content.WriteString(fmt.Sprintf("The system shall support `~%s~`covered[^~%s~]✅\n\n", reqSite, reqSite))
		} else {
			// Uncovered requirement
			content.WriteString(fmt.Sprintf("The system shall support `~%s~`uncvrd[^~%s~]❓\n\n", reqSite, reqSite))
		}

		// Add simple description
		content.WriteString("This requirement defines core functionality needed for system operations.\n\n")
	}

	// Add coverage footnotes
	footnoteAdded := false

	// Group coverers by requirementSite
	coverageFootnotes := make(map[string][]Coverer)
	for _, coverer := range reqFile.Coverers {
		coverageFootnotes[coverer.ReqSite] = append(coverageFootnotes[coverer.ReqSite], coverer)
	}

	// Process each requirement site that has coverage
	for reqSite, coverers := range coverageFootnotes {
		if len(coverers) > 0 {
			if !footnoteAdded {
				content.WriteString("\n") // Add empty line before first footnote
				footnoteAdded = true
			}

			// Sort coverers by coverage type, then by path, then by line
			sort.Slice(coverers, func(i, j int) bool {
				if coverers[i].CoverageType != coverers[j].CoverageType {
					return coverers[i].CoverageType < coverers[j].CoverageType
				}
				if coverers[i].RelPath != coverers[j].RelPath {
					return coverers[i].RelPath < coverers[j].RelPath
				}
				return coverers[i].Line < coverers[j].Line
			})

			// Create footnote hint
			footnoteLine := fmt.Sprintf("[^~%s~]: `[~%s/%s~%s]`",
				reqSite,
				reqFile.Package,
				reqSite,
				coverers[0].CoverageType) // Use the type of the first coverer

			// Add coverers
			for i, coverer := range coverers {
				fileURL := constructFileURL(config.RepoURL, config.CommitHash, coverer.RelPath)
				coverageURL := fmt.Sprintf("%s#L%d", fileURL, coverer.Line)
				coverageLabel := fmt.Sprintf("%s:%d:%s", coverer.RelPath, coverer.Line, coverer.CoverageType)

				if i > 0 {
					footnoteLine += ","
				}

				footnoteLine += fmt.Sprintf(" [%s](%s)", coverageLabel, coverageURL)
			}

			content.WriteString(footnoteLine + "\n")
		}
	}

	// Write to file
	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

// generateReqmdJson creates a reqmd.json file with the specified content
func generateReqmdJson(filePath string, reqmdJson *ReqmdJson) error {
	// Sort FileURLs lexically to avoid unnecessary changes
	var fileURLs []string
	for fileURL := range reqmdJson.FileHashes {
		fileURLs = append(fileURLs, fileURL)
	}
	sort.Strings(fileURLs)

	// Create ordered map
	orderedFileHashes := make(map[string]string)
	for _, fileURL := range fileURLs {
		orderedFileHashes[fileURL] = reqmdJson.FileHashes[fileURL]
	}
	reqmdJson.FileHashes = orderedFileHashes

	// Check if the file would be empty
	if len(reqmdJson.FileHashes) == 0 {
		// Delete the file if it exists (no need to create empty file)
		_ = os.Remove(filePath)
		return nil
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(reqmdJson, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal reqmd.json: %w", err)
	}

	// Write to file
	return os.WriteFile(filePath, data, 0644)
}

// generateSourceFile creates a source file that implements requirements with proper coverage tags
func generateSourceFile(filePath string, srcFile SrcFileDescr, coveringReqs []struct {
	Package      string
	ReqSite      string
	Line         int
	CoverageType string
}) error {
	var content strings.Builder

	// Add package declaration and imports
	packageName := filepath.Base(srcFile.RelDir)
	content.WriteString("// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors\n")
	content.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	content.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	content.WriteString("import (\n")
	content.WriteString("\t\"fmt\"\n")
	content.WriteString(")\n\n")

	// Add file tags as comments
	content.WriteString("// Tags: " + strings.Join(srcFile.Tags, ", ") + "\n\n")

	// Add functions implementing requirements
	lineCount := 1 + 8 // Account for header, imports, and tags comments

	for _, req := range coveringReqs {
		// Add padding lines if needed to reach the target line
		for lineCount < req.Line {
			content.WriteString("\n")
			lineCount++
		}

		funcName := fmt.Sprintf("Implement%s_%s", req.ReqSite, req.CoverageType)

		// Add coverage tag with proper format according to the design
		content.WriteString(fmt.Sprintf("// [~%s/%s~%s]\n", req.Package, req.ReqSite, req.CoverageType))
		lineCount++

		// Add function
		content.WriteString(fmt.Sprintf("func %s() {\n", funcName))
		lineCount++
		content.WriteString(fmt.Sprintf("\tfmt.Println(\"Implementation for %s/%s (%s)\")\n",
			req.Package, req.ReqSite, req.CoverageType))
		lineCount++
		content.WriteString("}\n\n")
		lineCount += 2
	}

	// Write to file
	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

// generateGoldenFiles creates golden files for testing
func generateGoldenFiles(baseDir string, reqFileDescrs []ReqFileDescr, srcFileDescrs []SrcFileDescr, config *HVGeneratorConfig) error {
	goldenDir := filepath.Join(baseDir, "golden")
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		return fmt.Errorf("failed to create golden directory: %w", err)
	}

	// Create a golden file with coverage information
	coverageFile := filepath.Join(goldenDir, "coverage.txt")
	var coverage strings.Builder

	// List requirements and their implementations
	for _, reqFile := range reqFileDescrs {
		coverage.WriteString(fmt.Sprintf("# %s\n", reqFile.Package))

		for _, reqSite := range reqFile.ReqSites {
			coverage.WriteString(fmt.Sprintf("## %s\n", reqSite))

			// Find coverers for this requirement
			var coverers []Coverer
			for _, coverer := range reqFile.Coverers {
				if coverer.ReqSite == reqSite {
					coverers = append(coverers, coverer)
				}
			}

			// Write coverage information
			coverage.WriteString(fmt.Sprintf("Implementations: %d\n", len(coverers)))
			for _, coverer := range coverers {
				fileURL := constructFileURL(config.RepoURL, config.CommitHash, coverer.RelPath)
				coverage.WriteString(fmt.Sprintf("- %s (line %d, type %s): %s#L%d\n",
					coverer.RelPath, coverer.Line, coverer.CoverageType, fileURL, coverer.Line))
			}
			coverage.WriteString("\n")
		}
		coverage.WriteString("\n")
	}

	// Write coverage file
	if err := os.WriteFile(coverageFile, []byte(coverage.String()), 0644); err != nil {
		return fmt.Errorf("failed to write coverage file: %w", err)
	}

	// Create a summary golden file
	summaryFile := filepath.Join(goldenDir, "summary.txt")
	var summary strings.Builder

	summary.WriteString("# High volume test summary\n\n")
	summary.WriteString(fmt.Sprintf("Requirement files: %d\n", len(reqFileDescrs)))
	summary.WriteString(fmt.Sprintf("Source files: %d\n", len(srcFileDescrs)))

	// Count total requirements and implementations
	totalReqs := 0
	totalImpls := 0
	implsByType := make(map[string]int)

	for _, reqFile := range reqFileDescrs {
		totalReqs += len(reqFile.ReqSites)
		totalImpls += len(reqFile.Coverers)

		for _, coverer := range reqFile.Coverers {
			implsByType[coverer.CoverageType]++
		}
	}

	// Count how many requirements have coverage
	coveredReqs := 0
	for _, reqFile := range reqFileDescrs {
		coveredReqsMap := make(map[string]bool)
		for _, coverer := range reqFile.Coverers {
			coveredReqsMap[coverer.ReqSite] = true
		}
		coveredReqs += len(coveredReqsMap)
	}

	summary.WriteString(fmt.Sprintf("Total requirements: %d\n", totalReqs))
	summary.WriteString(fmt.Sprintf("Requirements with coverage: %d\n", coveredReqs))
	summary.WriteString(fmt.Sprintf("Requirements without coverage: %d\n", totalReqs-coveredReqs))
	summary.WriteString(fmt.Sprintf("Total implementations: %d\n", totalImpls))
	summary.WriteString(fmt.Sprintf("Average implementations per requirement: %.2f\n", float64(totalImpls)/float64(totalReqs)))

	// Implementations by type
	summary.WriteString("\nImplementations by type:\n")
	for coverageType, count := range implsByType {
		summary.WriteString(fmt.Sprintf("- %s: %d\n", coverageType, count))
	}

	// Write summary file
	if err := os.WriteFile(summaryFile, []byte(summary.String()), 0644); err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}

	return nil
}
