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