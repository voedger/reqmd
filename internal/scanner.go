// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// Scan scans multiple paths that can each contain both markdown and source files
func (s *scanner) Scan(paths []string) (*ScannerResult, error) {
	// Reset statistics
	start := time.Now()
	s.stats.processedFiles.Store(0)
	s.stats.processedBytes.Store(0)
	s.stats.skippedFiles.Store(0)
	s.stats.skippedBytes.Store(0)

	files, errs, err := s.scanPaths(paths)
	if err != nil {
		return nil, err
	}

	result := &ScannerResult{
		Files:            files,
		ProcessingErrors: errs,
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

// scanFile handles both markdown and source files in a unified way
func (s *scanner) scanFile(mu *sync.Mutex, filePath string, mctx *MarkdownContext, gitRepos map[string]IGit, files *[]FileStructure, syntaxErrors *[]ProcessingError) error {
	filePath = filepath.ToSlash(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))

	if IsVerbose {
		Verbose("scanFile: filePath=" + filePath)
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

	// Skip files with unsupported extensions
	if ext != markdownExtension && !s.sourceExtensions[ext] {
		Verbose("Skipping unsupported file", "path", filePath, "extension", ext)
		return nil
	}

	// Always perform git verification for all files
	var git IGit
	for path, repo := range gitRepos {
		if strings.HasPrefix(filePath, path) {
			git = repo
			break
		}
	}
	if git == nil {
		return fmt.Errorf("no git repository found for file: %s", filePath)
	}

	// Get relative path for the file
	relPath, err := filepath.Rel(git.PathToRoot(), filePath)
	if IsVerbose {
		Verbose("scanFile: relPath=" + relPath + " PathToRoot=" + git.PathToRoot())
	}
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}
	relPath = filepath.ToSlash(relPath)

	if IsVerbose {
		Verbose("scanFile: before git.FileHash: " + filePath)
	}

	// Try to get file hash - this will fail for untracked files
	hash, err := git.FileHash(relPath)
	if err != nil {
		// Skip untracked files
		if IsVerbose {
			Verbose("scanFile: skipping untracked file: " + relPath + ", error: " + err.Error())
		}
		return nil
	}

	// Parse the file once

	if IsVerbose {
		Verbose("scanFile: before ParseFile: " + filePath)
	}

	structure, errs, err := ParseFile(mctx, filePath)
	if err != nil {
		return err
	}

	if structure != nil { // TODO: should check if it has requirements or coverage tags
		// Set FileStructure fields for URL construction for all files
		structure.FileHash = hash
		structure.RelativePath = relPath
		structure.RepoRootFolderURL = git.RepoRootFolderURL()

		// Add to files list if it has requirements or coverage tags
		if (ext == markdownExtension && len(structure.Requirements) > 0) ||
			(ext != markdownExtension && len(structure.CoverageTags) > 0) {
			mu.Lock()
			*files = append(*files, *structure)
			mu.Unlock()
		}
	}

	// Add any errors found during parsing
	if len(errs) > 0 {
		mu.Lock()
		*syntaxErrors = append(*syntaxErrors, errs...)
		mu.Unlock()
	}

	return nil
}

func (s *scanner) scanPaths(paths []string) ([]FileStructure, []ProcessingError, error) {
	var files []FileStructure
	var syntaxErrors []ProcessingError

	// Initialize git repositories for all paths
	gitRepos := make(map[string]IGit)
	for _, path := range paths {
		path = filepath.ToSlash(path)
		var gitPath string
		currentPath := path
		for {
			if _, err := os.Stat(filepath.Join(currentPath, gitFolderName)); err == nil {
				gitPath = currentPath
				break
			}
			parent := filepath.Dir(currentPath)
			if parent == currentPath {
				return nil, nil, fmt.Errorf("no git repository found for path: %s", path)
			}
			currentPath = parent
		}

		git, err := NewIGit(gitPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize git for path %s: %w", path, err)
		}
		gitRepos[path] = git
	}

	var mu sync.Mutex

	// Create a unified file processor that handles both markdown and source files
	folderProcessor := func(folderPath string) (FileProcessor, error) {

		// If folder name starts with a dot, skip it
		if strings.HasPrefix(filepath.Base(folderPath), ".") {
			Verbose("Skipping folder", "path", folderPath)
			return nil, nil
		}

		// Initialize markdown context for this folder
		mctx := &MarkdownContext{
			rfiles: &Reqmdjson{
				FileUrl2FileHash: make(map[string]string),
			},
		}

		// Try to load reqmd.json if it exists
		reqmdPath := filepath.Join(folderPath, ReqmdjsonFileName)
		if content, err := os.ReadFile(reqmdPath); err == nil {
			if err := json.Unmarshal(content, &mctx.rfiles); err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", ReqmdjsonFileName, err)
			}
		}

		return func(filePath string) error {
			return s.scanFile(&mu, filePath, mctx, gitRepos, &files, &syntaxErrors)
		}, nil
	}

	// Process all paths
	for _, path := range paths {
		if errs := FoldersScanner(defaultMaxWorkers, defaultMaxErrQueueSize, path, folderProcessor); len(errs) > 0 {
			return nil, nil, fmt.Errorf("error scanning files in %s: %v", path, errs[0])
		}
	}

	return files, syntaxErrors, nil
}
