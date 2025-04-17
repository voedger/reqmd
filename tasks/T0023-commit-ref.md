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

## Implementation plan

### 1. Update EBNF definitions

- Modify the EBNF grammar in [docs/ebnf.md](../docs/ebnf.md) to replace CommitHash with CommitRef
- Update related EBNF definitions for CoverageFootnote, CoverageURL, and other affected elements

### 2. Update data models

- Modify the models in [internal/models.go](../internal/models.go):
  - Replace CommitHash field with CommitRef in relevant structs
  - Update FileURL construction to use CommitRef instead of CommitHash

### 3. Update GitHash retrieval

- Modify [internal/gogit.go](../internal/gogit.go) to:
  - Add GetDefaultBranch() function that returns "main" by default
  - Update interface and implementation methods to work with CommitRef instead of CommitHash

### 4. Update Analyzer logic

- Modify [internal/analyzer.go](../internal/analyzer.go):
  - Remove hash comparison logic since it's no longer needed
  - Update action generation to use CommitRef instead of CommitHash

### 5. Update URL construction

- Update the FileURL construction logic to use CommitRef
- Ensure correct URL formatting for different Git providers (GitHub, GitLab)

### 6. Remove reqmdfiles.json handling

- Remove code that reads/writes reqmdfiles.json since we don't need to track file hashes anymore
- Update [internal/applier.go](../internal/applier.go) to remove ApplyReqmdjsonAction function
- Update tests to reflect the removal of this functionality

### 7. Update tests

- Modify existing tests to use CommitRef instead of CommitHash
- Update test fixtures to reflect the new URL format
- Add tests that verify the behavior with different branch names

### 8. Documentation updates

- Update [docs/design.md](../docs/design.md) to reflect the new approach
- Update [README.md](../README.md) to explain the changes
- Add explanation that line numbers may change if files are modified

### 9. System tests

- Update [docs/design-systests.md](../docs/design-systests.md) test cases to reflect these changes
- Add tests that verify correct behavior with different branch names
