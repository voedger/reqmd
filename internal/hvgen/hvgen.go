// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package hvgen

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
	}
}

// hvGenerator generates test files with configurable parameters for high-volume testing
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

	// Initialize random number generator (Go 1.20+ approach)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create base directory structure
	mdDir := filepath.Join(config.BaseDir, "req")
	srcDir := filepath.Join(config.BaseDir, "src")

	if err := os.MkdirAll(mdDir, 0755); err != nil {
		return fmt.Errorf("failed to create markdown directory: %w", err)
	}
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Track all generated requirements for coverage implementation
	allRequirements := make([]struct {
		ModuleName string
		ReqName    string
	}, 0, config.NumMarkdownFiles*config.ReqsPerMarkdownFile)

	// Generate markdown files with requirements
	for i := range config.NumMarkdownFiles {
		moduleName := fmt.Sprintf("%s.module%d", config.PackageIDPrefix, i+1)
		mdFilePath := filepath.Join(mdDir, fmt.Sprintf("req_%d.md", i+1))

		requirements, err := generateMarkdownFile(mdFilePath, moduleName, config.ReqsPerMarkdownFile, rng)
		if err != nil {
			return fmt.Errorf("failed to generate markdown file %s: %w", mdFilePath, err)
		}

		for _, req := range requirements {
			allRequirements = append(allRequirements, struct {
				ModuleName string
				ReqName    string
			}{moduleName, req})
		}
	}

	// Generate source files with implementations
	// Distribute implementations among source files
	for i := range config.NumSourceFiles {
		// Create subdirectories for better structure
		subdir := fmt.Sprintf("pkg%d", i/1000)
		fileDir := filepath.Join(srcDir, subdir)
		if err := os.MkdirAll(fileDir, 0755); err != nil {
			return fmt.Errorf("failed to create source subdirectory: %w", err)
		}

		srcFilePath := filepath.Join(fileDir, fmt.Sprintf("file_%d.go", i+1))

		// Choose random requirements to implement in this file
		numRequirementsToImplement := min(
			rng.Intn(5)+1, // 1 to 5 requirements per file
			len(allRequirements),
		)

		implementedReqs := make([]struct {
			ModuleName string
			ReqName    string
		}, 0, numRequirementsToImplement)

		// Randomly select requirements to implement
		for range numRequirementsToImplement {
			reqIdx := rng.Intn(len(allRequirements))
			implementedReqs = append(implementedReqs, allRequirements[reqIdx])
		}

		if err := generateSourceFile(srcFilePath, implementedReqs, rng); err != nil {
			return fmt.Errorf("failed to generate source file %s: %w", srcFilePath, err)
		}
	}

	return nil
}

// generateMarkdownFile creates a markdown file with the specified number of requirements
func generateMarkdownFile(filePath, moduleName string, numRequirements int, rng *rand.Rand) ([]string, error) {
	var content strings.Builder
	var requirements []string

	// Add header
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("reqmd.package: %s\n", moduleName))
	content.WriteString("---\n\n")
	content.WriteString(fmt.Sprintf("# Module %s\n\n", moduleName))
	content.WriteString("## Requirements\n\n")

	// Add requirements
	for i := range numRequirements {
		reqName := fmt.Sprintf("REQ%03d", i+1)
		requirements = append(requirements, reqName)

		// Add requirement with random complexity/description
		content.WriteString(fmt.Sprintf("### %s\n\n", reqName))
		content.WriteString(fmt.Sprintf("The system shall support `~%s~`\n\n", reqName))

		// Add random description
		descriptionLength := rng.Intn(5) + 1 // 1 to 5 paragraphs
		for j := 0; j < descriptionLength; j++ {
			sentenceCount := rng.Intn(5) + 1 // 1 to 5 sentences
			var paragraph strings.Builder
			for k := 0; k < sentenceCount; k++ {
				sentence := getRandomSentence(rng)
				paragraph.WriteString(sentence)
				paragraph.WriteString(" ")
			}
			content.WriteString(paragraph.String())
			content.WriteString("\n\n")
		}
	}

	// Write content to file
	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return nil, err
	}

	return requirements, nil
}

// generateSourceFile creates a source file that implements the specified requirements
func generateSourceFile(filePath string, requirements []struct {
	ModuleName string
	ReqName    string
}, rng *rand.Rand) error {
	var content strings.Builder

	// Add package declaration and imports
	packageName := filepath.Base(filepath.Dir(filePath))
	content.WriteString("// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors\n")
	content.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	content.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	content.WriteString("import (\n")
	content.WriteString("\t\"fmt\"\n")
	content.WriteString(")\n\n")

	// Add random functions for each requirement
	for i, req := range requirements {
		funcName := fmt.Sprintf("Implement%s_%d", req.ReqName, i+1)

		// Add coverage tag
		content.WriteString(fmt.Sprintf("// [~ %s/%s ~impl]\n", req.ModuleName, req.ReqName))

		// Add function implementation
		content.WriteString(fmt.Sprintf("func %s() {\n", funcName))
		content.WriteString(fmt.Sprintf("\tfmt.Println(\"Implementation for %s/%s\")\n", req.ModuleName, req.ReqName))

		// Add some random code to make files larger/more diverse
		linesOfCode := rng.Intn(20) + 5 // 5 to 25 lines
		for j := range linesOfCode {
			content.WriteString(fmt.Sprintf("\t// Line of code %d\n", j+1))
		}

		content.WriteString("}\n\n")
	}

	// Write content to file
	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

// getRandomSentence returns a random sentence for requirement descriptions
func getRandomSentence(rng *rand.Rand) string {
	sentences := []string{
		"The system must process inputs efficiently.",
		"All transactions shall be logged for audit purposes.",
		"Users should receive real-time notifications for important events.",
		"The interface must be intuitive and responsive.",
		"Data should be encrypted both at rest and in transit.",
		"Performance metrics should be collected for all operations.",
		"The API must follow RESTful principles.",
		"All errors must be captured and properly reported.",
		"The system shall maintain backward compatibility.",
		"Documentation must be comprehensive and up-to-date.",
	}
	return sentences[rng.Intn(len(sentences))]
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
