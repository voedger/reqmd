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

	// Default source file extensions
	defaultSourceExtensions = ".go,.js,.ts,.jsx,.tsx,.java,.cs,.cpp,.c,.h,.hpp,.py,.rb,.php,.rs,.kt,.scala,.m,.swift,.fs,.md,.sql,.vsql"
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

func (s *scanner) Scan(reqPath string, srcPaths []string) ([]FileStructure, []ProcessingError, error) {
	result, err := s.scan(reqPath, srcPaths)
	if err != nil {
		return nil, nil, err
	}
	return result.Files, result.ProcessingErrors, nil
}

type scanner struct {
	sourceExtensions map[string]bool
}

type ScanResult struct {
	Files        []FileStructure
	ProcessingErrors []ProcessingError
}

/*

- Scan accepts list of source file extensions as a parameter (besides other parameters) to allow scanning only specific types of source files
- If it is empty, default set of the source file extensions is used that covers all popular programming languages

*/

func (s *scanner) scan(reqPath string, srcPaths []string) (*ScanResult, error) {
	result := &ScanResult{}

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

	return result, nil
}

func scanMarkdowns(reqPath string) ([]FileStructure, []ProcessingError, error) {
	var files []FileStructure
	var syntaxErrors []ProcessingError

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
	if strings.Contains(filePath, gitFolderName) {
		return nil
	}

	// Check if file extension is supported
	ext := strings.ToLower(filepath.Ext(filePath))
	if !s.sourceExtensions[ext] {
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
