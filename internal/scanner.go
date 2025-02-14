package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// File extensions and patterns
	markdownExtension = ".md"
	gitFolderName     = ".git"
	reqmdConfigFile   = "reqmdfiles.json"

	// Scanner configuration
	defaultMaxWorkers      = 32
	defaultMaxErrQueueSize = 1000
)

/*

An exerpt from design.md

- **Purpose**: Implements `IScanner`.
- **Key functions**:
  - `Scan`:
    - Recursively discover Markdown and source files.
    - Delegate parsing to specialized components (`mdparser.go`, `srccoverparser.go`).
    - Build a unified list of `FileStructure` objects for each file.
    - Collect any `SyntaxError`s.
- **Responsibilities**:
  - Single responsibility: collecting raw data (files, coverage tags, requirement references) and building the domain model.
  - Potential concurrency (goroutines) for scanning subfolders.

*/

func NewScanner() IScanner {
	return &scanner{}
}

func (s *scanner) Scan(reqPath string, srcPaths []string) ([]FileStructure, []SyntaxError, error) {
	result, err := Scan(reqPath, srcPaths)
	if err != nil {
		return nil, nil, err
	}
	return result.Files, result.SyntaxErrors, nil
}

type scanner struct{}

type ScanResult struct {
	Files        []FileStructure
	SyntaxErrors []SyntaxError
}

/*

- First path is processed as path to requirement files
- Other paths are processed as path to source files using
- Paths are processes sequentially by FoldersScanner

Requirement files
- Uses FoldersScanner and ParseMarkdownFile
- FolderProcessor parses reqmdfiles.json (if exists) and passes to FileProcessor

Source files
- git repository shall be found in Path or parent directories
- IGit is instantiated using NewIGit and a folder that contains .git folder
- IGit should be passed to FolderProcessor and all FileProcessors
- Uses FoldersScanner and ParseSourceFile

*/

func Scan(reqPath string, srcPaths []string) (*ScanResult, error) {
	result := &ScanResult{}

	// Scan markdown files
	files, errs, err := ScanMarkdowns(reqPath)
	if err != nil {
		return nil, err
	}
	result.Files = append(result.Files, files...)
	result.SyntaxErrors = append(result.SyntaxErrors, errs...)

	// Scan source files if any paths provided
	if len(srcPaths) > 0 {
		files, errs, err := ScanSources(srcPaths)
		if err != nil {
			return nil, err
		}
		result.Files = append(result.Files, files...)
		result.SyntaxErrors = append(result.SyntaxErrors, errs...)
	}

	return result, nil
}

func ScanMarkdowns(reqPath string) ([]FileStructure, []SyntaxError, error) {
	var files []FileStructure
	var syntaxErrors []SyntaxError

	reqmdProcessor := func(folder string) (FileProcessor, error) {
		mctx := &MarkdownContext{
			rfiles: make(ReqmdfilesMap),
		}

		reqmdPath := filepath.Join(folder, reqmdConfigFile)
		if content, err := os.ReadFile(reqmdPath); err == nil {
			if err := json.Unmarshal(content, &mctx.rfiles); err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", reqmdConfigFile, err)
			}
		}

		return func(filePath string) error {
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

func ScanSources(srcPaths []string) ([]FileStructure, []SyntaxError, error) {
	var files []FileStructure
	var syntaxErrors []SyntaxError

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
				return processSourceFile(filePath, git, &files, &syntaxErrors)
			}, nil
		}

		if errs := FoldersScanner(defaultMaxWorkers, defaultMaxErrQueueSize, srcPath, srcProcessor); len(errs) > 0 {
			return nil, nil, fmt.Errorf("error scanning source files in %s: %v", srcPath, errs[0])
		}
	}

	return files, syntaxErrors, nil
}

func processSourceFile(filePath string, git IGit, files *[]FileStructure, syntaxErrors *[]SyntaxError) error {
	if strings.Contains(filePath, gitFolderName) {
		return nil
	}

	structure, errs, err := ParseSourceFile(filePath)
	if err != nil {
		return err
	}

	if structure != nil && len(structure.CoverageTags) > 0 {
		// Get relative path for the file
		relPath, err := filepath.Rel(git.PathToRoot(), filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Get file hash
		hash, err := git.FileHash(relPath)
		if err != nil {
			return fmt.Errorf("failed to get file hash: %w", err)
		}

		// Set FileStructure fields for URL construction
		structure.FileHash = hash
		structure.RelativePath = filepath.ToSlash(relPath)    // Convert Windows paths to URL format
		structure.RepoRootFolderURL = git.RepoRootFolderURL() // Get base URL from git

		*files = append(*files, *structure)
	}

	if len(errs) > 0 {
		*syntaxErrors = append(*syntaxErrors, errs...)
	}

	return nil
}
