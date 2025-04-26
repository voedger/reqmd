// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	gog "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func NewGitVCS(path string) (IVCS, error) {

	// Find path to the root of the git repository
	{
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		absPath = filepath.ToSlash(absPath)
		var gitPath string
		currentPath := absPath
		for {
			if _, err := os.Stat(filepath.Join(currentPath, gitFolderName)); err == nil {
				gitPath = currentPath
				break
			}
			parent := filepath.Dir(currentPath)
			if parent == currentPath {
				return nil, fmt.Errorf("no git repository found for path: %s", path)
			}
			currentPath = parent
		}
		path = gitPath
	}

	repo, err := gog.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	g := &git{
		pathToRoot: path,
		repo:       repo,
		commit:     commit,
		tree:       tree,
	}

	if err := g.constructRepoRootFolderURL(); err != nil {
		return nil, fmt.Errorf("failed to construct repo URL: %w", err)
	}
	return g, nil
}

type git struct {
	pathToRoot        string
	repo              *gog.Repository
	commit            *object.Commit
	tree              *object.Tree
	repoRootFolderURL string // Cached during initialization
	mu                sync.RWMutex
}

// Returns the hash of a file in the git repository.
// filePath is not necessary relative to the repository root.
func (g *git) FileHash(filePath string) (relPath, hash string, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	relPath, err = filepath.Rel(g.PathToRoot(), filePath)
	relPath = filepath.ToSlash(relPath)
	if err != nil {
		return "", "", err
	}
	file, err := g.tree.FindEntry(relPath)
	if err != nil {
		return "", "", err
	}
	return relPath, file.Hash.String(), nil
}

func (g *git) PathToRoot() string {
	return g.pathToRoot
}

func (g *git) constructRepoRootFolderURL() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Get remote URL
	remote, err := g.repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote: %w", err)
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return fmt.Errorf("no URLs found for origin remote")
	}
	remoteURL := urls[0]

	// Determine which branch to use
	var commitRef string

	// Check if main branch exists
	_, err = g.repo.Reference("refs/heads/main", false)
	if err == nil {
		// Main branch exists
		commitRef = "main"
	} else {
		// Check if master branch exists
		_, err = g.repo.Reference("refs/heads/master", false)
		if err == nil {
			// Master branch exists
			commitRef = "master"
		} else {
			return fmt.Errorf("neither main nor master branch exists")
		}
	}

	// Detect provider and construct URL
	switch {
	case strings.Contains(remoteURL, "github.com"):
		g.repoRootFolderURL = fmt.Sprintf("%s/blob/%s", remoteURL, commitRef)
	case strings.Contains(remoteURL, "gitlab.com"):
		g.repoRootFolderURL = fmt.Sprintf("%s/-/blob/%s", remoteURL, commitRef)
	default:
		return fmt.Errorf("unsupported git provider: %s", remoteURL)
	}

	return nil
}

func (g *git) RepoRootFolderURL() string {
	return g.repoRootFolderURL
}
