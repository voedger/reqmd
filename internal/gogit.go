package internal

import (
	gog "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func NewIGit(path string) (IGit, error) {
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

	return &git{
		pathToRoot: path,
		repo:       repo,
		commit:     commit,
		tree:       tree,
	}, nil
}

type git struct {
	pathToRoot string
	repo       *gog.Repository
	commit     *object.Commit
	tree       *object.Tree
}

// Returns the hash of a file in the git repository.
// filePath is not necessary relative to the repository root.
func (g *git) FileHash(filePath string) (string, error) {
	file, err := g.tree.FindEntry(filePath)
	if err != nil {
		return "", err
	}
	return file.Hash.String(), nil
}

func (g *git) PathToRoot() string {
	return g.pathToRoot
}

func (g *git) CommitHash() string {
	return g.commit.Hash.String()
}
