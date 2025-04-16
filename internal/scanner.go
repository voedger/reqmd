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
	defaultMaxErrQueueSize = 50

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

	// Reset result
	res := &ScannerResult{
		Files:            []FileStructure{},
		ProcessingErrors: []ProcessingError{},
	}
	s.result = res

	// Reset statistics
	start := time.Now()
	s.stats.processedFiles.Store(0)
	s.stats.processedBytes.Store(0)
	s.stats.skippedFiles.Store(0)
	s.stats.skippedBytes.Store(0)

	err := s.scanPaths(paths)

	if err != nil {
		return nil, err
	}

	// Report statistics after scanning is complete
	Verbose("Scan complete (multi-path)",
		"processed files", s.stats.processedFiles.Load(),
		"processed size", ByteCountSI(s.stats.processedBytes.Load()),
		"skipped files", s.stats.skippedFiles.Load(),
		"skipped size", ByteCountSI(s.stats.skippedBytes.Load()),
		"duration", time.Since(start),
	)

	return s.result, nil
}

type scanner struct {
	sourceExtensions map[string]bool
	stats            struct {
		processedFiles atomic.Int64
		processedBytes atomic.Int64
		skippedFiles   atomic.Int64
		skippedBytes   atomic.Int64
	}
	mu     sync.Mutex
	result *ScannerResult
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
func (s *scanner) scanFile(filePath string, mctx *MarkdownContext, igit IGit) error {
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

	// Try to get file hash - this will fail for untracked files
	relPath, hash, err := igit.FileHash(filePath)
	if err != nil {
		// Skip untracked files
		if IsVerbose {
			Verbose("scanFile: skipping untracked file: " + filePath + ", error: " + err.Error())
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
		structure.RepoRootFolderURL = igit.RepoRootFolderURL()

		// Add to files list if it has requirements or coverage tags
		if (ext == markdownExtension && len(structure.Requirements) > 0) ||
			(ext != markdownExtension && len(structure.CoverageTags) > 0) {
			s.mu.Lock()
			s.result.Files = append(s.result.Files, *structure)
			s.mu.Unlock()
		}
	}

	// Add any errors found during parsing
	if len(errs) > 0 {
		s.mu.Lock()
		s.result.ProcessingErrors = append(s.result.ProcessingErrors, errs...)
		s.mu.Unlock()
	}

	return nil
}

func (s *scanner) scanPaths(paths []string) (err error) {

	// Process all paths
	for _, path := range paths {

		git, err := NewIGit(path)
		if err != nil {
			return fmt.Errorf("failed to initialize git for path %s: %w", path, err)
		}

		fp := func(filePath string) (FileProcessor, error) {
			return s.folderProcessor(filePath, git)
		}

		if errs := FoldersScanner(defaultMaxWorkers, defaultMaxErrQueueSize, path, fp); len(errs) > 0 {
			return fmt.Errorf("error scanning files in %s: %v", path, errs[0])
		}
	}

	return nil
}

func (s *scanner) folderProcessor(folderPath string, igit IGit) (FileProcessor, error) {

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
		return s.scanFile(filePath, mctx, igit)
	}, nil

}
