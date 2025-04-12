// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

const (
	// File extensions and patterns
	markdownExtension = ".md"
	gitFolderName     = ".git"

	// Scanner configuration
	defaultMaxWorkers      = 32
	defaultMaxErrQueueSize = 1000

	// Default source file extensions
	defaultSourceExtensions = ".go,.js,.ts,.jsx,.tsx,.java,.cs,.cpp,.c,.h,.hpp,.py,.rb,.php,.rs,.kt,.scala,.m,.swift,.fs,.md,.sql,.vsql"

	// Maximum file size for processing
	maxFileSize = 128 * 1024 // 128KB
)

func NewScanner(extensions string) IScanner {
	s := &scanner{
		sourceExtensions: make(map[string]bool),
	}
	// Use provided extensions or fallback to defaults
	exts := extensions
	if exts == "" {
		exts = defaultSourceExtensions
	}
	// Initialize extensions map
	for _, ext := range strings.Split(exts, ",") {
		s.sourceExtensions[strings.TrimSpace(ext)] = true
	}
	return s
}

func (s *scanner) Scan(reqPath string, srcPaths []string) (*ScannerResult, error) {
	return s.scan(reqPath, srcPaths)
}

// ScanMultiPath scans multiple paths that can each contain both markdown and source files
func (s *scanner) ScanMultiPath(paths []string) (*ScannerResult, error) {
	// Reset statistics
	start := time.Now()
	s.stats.processedFiles.Store(0)
	s.stats.processedBytes.Store(0)
	s.stats.skippedFiles.Store(0)
	s.stats.skippedBytes.Store(0)

	result := &ScannerResult{}

	// For each path, scan both markdown and source files
	for _, path := range paths {
		// Scan for markdown files in this path
		mdFiles, mdErrs, err := scanMarkdowns(path)
		if err != nil {
			return nil, err
		}
		result.Files = append(result.Files, mdFiles...)
		result.ProcessingErrors = append(result.ProcessingErrors, mdErrs...)

		// Scan for source files in this path
		srcFiles, srcErrs, err := s.scanSources([]string{path})
		if err != nil {
			return nil, err
		}
		result.Files = append(result.Files, srcFiles...)
		result.ProcessingErrors = append(result.ProcessingErrors, srcErrs...)
	}

	// Report statistics after scanning is complete
	Verbose("Scan complete (multi-path)",
		"processed files", s.stats.processedFiles.Load(),
		"processed size", ByteCountSI(s.stats.processedBytes.Load()),
		"skipped files", s.stats.skippedFiles.Load(),
		"skipped size", ByteCountSI(s.stats.skippedBytes.Load()),
		"duration", time.Since(start),
	)

	return result, nil
}

type scanner struct {
	sourceExtensions map[string]bool
	stats            struct {
		processedFiles atomic.Int64
		processedBytes atomic.Int64
		skippedFiles   atomic.Int64
		skippedBytes   atomic.Int64
	}
}

/*

- Scan accepts list of source file extensions as a parameter (besides other parameters) to allow scanning only specific types of source files
- If it is empty, default set of the source file extensions is used that covers all popular programming languages

*/

func (s *scanner) scan(reqPath string, srcPaths []string) (*ScannerResult, error) {
	// Reset statistics
	start := time.Now()
	s.stats.processedFiles.Store(0)
	s.stats.processedBytes.Store(0)
	s.stats.skippedFiles.Store(0)
	s.stats.skippedBytes.Store(0)

	result := &ScannerResult{}

	// Scan markdown files
	files, errs, err := scanMarkdowns(reqPath)
	if err != nil {
		return nil, err
	}
	result.Files = append(result.Files, files...)
	result.ProcessingErrors = append(result.ProcessingErrors, errs...)

	// Scan source files if any paths provided
	if len(srcPaths) > 0 {
		files, errs, err := s.scanSources(srcPaths)
		if err != nil {
			return nil, err
		}
		result.Files = append(result.Files, files...)
		result.ProcessingErrors = append(result.ProcessingErrors, errs...)
	}

	// Report statistics after scanning is complete
	Verbose("Scan complete",
		"processed files", s.stats.processedFiles.Load(),
		"processed size", ByteCountSI(s.stats.processedBytes.Load()),
		"skipped files", s.stats.skippedFiles.Load(),
		"skipped size", ByteCountSI(s.stats.skippedBytes.Load()),
		"duration", time.Since(start),
	)

	return result, nil
}

func scanMarkdowns(reqPath string) ([]FileStructure, []ProcessingError, error) {
	var files []FileStructure
	var syntaxErrors []ProcessingError

	reqmdProcessor := func(folder string) (FileProcessor, error) {
		mctx := &MarkdownContext{
			rfiles: &Reqmdjson{
				FileUrl2FileHash: make(map[string]string),
			},
		}

		reqmdPath := filepath.Join(folder, ReqmdjsonFileName)
		if content, err := os.ReadFile(reqmdPath); err == nil {
			if err := json.Unmarshal(content, &mctx.rfiles); err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", ReqmdjsonFileName, err)
			}
		}

		return func(filePath string) error {
			filePath = filepath.ToSlash(filePath)

			if !strings.HasSuffix(strings.ToLower(filePath), markdownExtension) {
				return nil
			}

			structure, errs, err := ParseMarkdownFile(mctx, filePath)
			if err != nil {
				return err
			}

			if structure != nil && len(structure.Requirements) > 0 {
				files = append(files, *structure)
			}
			if len(errs) > 0 {
				syntaxErrors = append(syntaxErrors, errs...)
			}
			return nil
		}, nil
	}

	if errs := FoldersScanner(defaultMaxWorkers, defaultMaxErrQueueSize, reqPath, reqmdProcessor); len(errs) > 0 {
		return nil, nil, fmt.Errorf("error scanning markdown files: %v", errs[0])
	}

	return files, syntaxErrors, nil
}

func (s *scanner) scanSources(srcPaths []string) ([]FileStructure, []ProcessingError, error) {
	var files []FileStructure
	var syntaxErrors []ProcessingError

	// Initialize git repositories for all source paths
	gitRepos := make(map[string]IGit)
	for _, srcPath := range srcPaths {
		var gitPath string
		currentPath := srcPath
		for {
			if _, err := os.Stat(filepath.Join(currentPath, gitFolderName)); err == nil {
				gitPath = currentPath
				break
			}
			parent := filepath.Dir(currentPath)
			if parent == currentPath {
				return nil, nil, fmt.Errorf("no git repository found for path: %s", srcPath)
			}
			currentPath = parent
		}

		git, err := NewIGit(gitPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize git for path %s: %w", srcPath, err)
		}
		gitRepos[srcPath] = git
	}

	for srcPath, git := range gitRepos {
		srcProcessor := func(folder string) (FileProcessor, error) {
			return func(filePath string) error {
				return s.processSourceFile(filePath, git, &files, &syntaxErrors)
			}, nil
		}

		if errs := FoldersScanner(defaultMaxWorkers, defaultMaxErrQueueSize, srcPath, srcProcessor); len(errs) > 0 {
			return nil, nil, fmt.Errorf("error scanning source files in %s: %v", srcPath, errs[0])
		}
	}

	return files, syntaxErrors, nil
}

func (s *scanner) processSourceFile(filePath string, git IGit, files *[]FileStructure, syntaxErrors *[]ProcessingError) error {

	filePath = filepath.ToSlash(filePath)

	// Check if file extension is supported
	ext := strings.ToLower(filepath.Ext(filePath))
	if !s.sourceExtensions[ext] {
		return nil
	}

	// Get file info to check size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Skip large files
	if fileInfo.Size() > maxFileSize {
		s.stats.skippedFiles.Add(1)
		s.stats.skippedBytes.Add(fileInfo.Size())
		Verbose("Skipping large file", "path", filePath, "size", ByteCountSI(fileInfo.Size()))
		return nil
	}

	// Track processed file
	s.stats.processedFiles.Add(1)
	s.stats.processedBytes.Add(fileInfo.Size())

	// Get relative path for the file
	relPath, err := filepath.Rel(git.PathToRoot(), filePath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}
	relPath = filepath.ToSlash(relPath)

	// Try to get file hash - this will fail for untracked files
	hash, err := git.FileHash(relPath)
	if err != nil {
		// Skip untracked files
		return nil
	}

	structure, errs, err := ParseSourceFile(filePath)
	if err != nil {
		return err
	}

	if structure != nil && len(structure.CoverageTags) > 0 {
		// Set FileStructure fields for URL construction
		structure.FileHash = hash
		structure.RelativePath = filepath.ToSlash(relPath)
		structure.RepoRootFolderURL = git.RepoRootFolderURL()

		*files = append(*files, *structure)
	}

	if len(errs) > 0 {
		*syntaxErrors = append(*syntaxErrors, errs...)
	}

	return nil
}

// ByteCountSI converts bytes to human readable string using SI (decimal) units
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
