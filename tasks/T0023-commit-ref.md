# Replace CommitHash with CommitRef

## Motivation

Currently, the EBNF grammar for CoverageFootnote uses CommitHash to identify a specific commit in a Git repository.

Pros:

- CommitHash is a unique identifier for a commit, ensuring that the reference by FileURL is always accurate and unambiguous

Cons:

- Requires tracking file hashes in regmd.json so that the tool can determine if the file has changed and replace the commit hash with the new one
- Makes it impossible to run the reqmd tool in a branch that will be squashed

Better tradeoff balance: use commit reference with "main" as the default value.

Pros:

- No need to track file hashes in regmd.json
- Enables running the reqmd tool in a branch that will be squashed, since `main` is used as the default commit ref

Cons:

- FileURL references are not always accurate (line numbers may change)

## Prompts

- Suggest an implementation plan for "# Replace CommitHash with CommitRef". Break the work into small steps. Each step should lead to working tests. @internal @docs🏆

## Implementation plan

### Step 1: Update EBNF definitions in docs/ebnf.md

- Modify the EBNF grammar definition for `CommitRef` 
- Note: The EBNF already has the correct definition showing `CommitRef = "main" | "master" | BranchName | CommitHash`
- Ensure documentation reflects the preference for `main` as the default reference

### Step 2: Update IGit interface and implementation


#### 2.1

- Update the `git` implementation in `internal/gogit.go`:
  - constructRepoRootFolderURL
    - commitRef is calculated as follows:
      - If main branch exists then `main` is usued
      - If master branch exists then `master` is usued
      - Otherwise error is returned

### Step 3: Update URL construction in analyzer

- Update the coverer URL generation in `internal/analyzer.go` to use commit ref instead of commit hash
- Modify how file URLs are constructed using the commit ref

### Step 4: Remove file hash tracking from reqmd.json

- Update the `Reqmdjson` structure in `internal/models.go` to no longer require file hashes
- Modify `FileUrl2FileHash` to be optional or remove if not needed

### Step 5: Update tests for CommitRef changes

- Add unit tests for the new `CommitRef()` method
- Update existing tests for URL construction and comparison
- Test scenarios with branch names vs. commit hashes

### Step 6: Update system tests to use CommitRef

- Modify templates in system test files to use CommitRef instead of CommitHash
- Update test assertions to verify proper usage of CommitRef

### Step 7: Implement backward compatibility

- Add code to handle existing files with commit hashes
- Ensure migration path for projects already using commit hashes

### Step 8: Update URL comparison logic

- Modify comparison logic in the analyzer to ignore commit ref differences
- Focus comparison on the actual file path and line number

### Step 9: Update documentation

- Update relevant documentation to reflect the change
- Explain benefits of using commit ref over commit hash

### Detailed Implementation Steps

#### Step 1: Update EBNF definitions
```
1. Confirm the existing EBNF definition in docs/ebnf.md is already correct
2. Add documentation notes about the preference for "main" as default
```

#### Step 2: Update IGit interface
```go
// Add CommitRef() method to the IGit interface
type IGit interface {
    PathToRoot() string
    FileHash(absoluteFilePath string) (relPath, hash string, err error)
    RepoRootFolderURL() string
    CommitHash() string
    CommitRef() string  // Add this new method
}

// Implement CommitRef() in git struct
func (g *git) CommitRef() string {
    g.mu.RLock()
    defer g.mu.RUnlock()
    
    // Try to get branch name first
    ref, err := g.repo.Head()
    if err != nil {
        return "main" // Default to main on error
    }
    
    // Check if we have a branch name
    if ref.Name().IsBranch() {
        return ref.Name().Short()
    }
    
    // Fall back to "main" if not a branch
    return "main"
}

// Update constructRepoRootFolderURL to use CommitRef
func (g *git) constructRepoRootFolderURL() error {
    // ... existing code ...
    
    // Use commitRef instead of hash
    commitRef := g.CommitRef()
    
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
```

#### Step 3: Update analyzer and URL construction
```go
// Update analyzer.go to use CommitRef for URL construction
// In the section where coverer URLs are generated
coverer := &Coverer{
    CoverageLabel: file.RelativePath + ":" + fmt.Sprint(tag.Line) + ":" + tag.CoverageType,
    CoverageUrL:   file.FileURL() + "#L" + strconv.Itoa(tag.Line),
    // FileHash field can be removed or made optional
}
```

#### Step 4: Remove file hash tracking from reqmd.json
```go
// Update Reqmdjson structure in models.go
type Reqmdjson struct {
    // Empty or add other fields as needed
    // FileUrl2FileHash no longer needed
}

// Remove or update code that deals with FileUrl2FileHash
```

#### Step 5-9: Test and documentation updates
These steps involve careful updates to the test files and documentation, ensuring that all components work correctly with the new CommitRef approach.

This implementation plan breaks down the work into logical, testable steps while ensuring that each step can be tested individually. The focus is on maintaining backward compatibility while transitioning to the new, more flexible approach using commit references.
