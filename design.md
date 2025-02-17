# Design

Design of the reqmd tool.

## Overview

The solution is split into a CLI entry point (`main.go`) and an internal package (`internal/`). Within `internal/`, each file, structure, and function has a focused responsibility. All data structures are centralized in `models.go`, and all interfaces in `interfaces.go`. Implementations are named by removing the `I` prefix.

The tool follows a three-stage pipeline architecture:

1. **Scan** – Build the Domain Model
   - Recursively discovers Markdown and source files
   - Extracts requirement references from Markdown files
   - Identifies coverage tags in source code
   - Builds file structures with Git metadata
   - Collects any syntax errors during parsing

2. **Analyze** – Validate and Plan Changes
   - Performs semantic validation checks
   - Ensures requirement IDs are unique
   - Determines which coverage footnotes need updates
   - Identifies requirements needing coverage annotations
   - Verifies file hashes against reqmd.json
   - Generates a list of required file modifications

3. **Apply** – Update Files
   - Updates or creates coverage footnotes
   - Appends coverage annotations to requirements
   - Maintains reqmd.json for file tracking
   - Ensures changes are made only when no errors exist

The system is designed using SOLID principles:

1. **Single Responsibility Principle**  
   - Each component (scanner, analyzer, applier) is responsible for exactly one stage of the process.  

2. **Open/Closed Principle**  
   - New features can be added by creating new parsers or new analysis rules without modifying existing, stable components.
   - For example, if a new coverage system is added, you can create a new parser that returns coverage tags in the same data model.

3. **Liskov Substitution Principle**
   - Interfaces (`IScanner`, `IAnalyzer`, `IApplier`) can be replaced with new implementations as long as they respect the same contracts.

4. **Interface Segregation Principle**
   - Instead of having a single “monolithic” interface, smaller interfaces reflect the actual steps: scanning, analyzing, applying.
   - The consumer (the `Tracer`) depends only on the interfaces it needs.

5. **Dependency Inversion Principle**
   - `Tracer` depends on the abstract interfaces (`IScanner`, `IAnalyzer`, `IApplier`), not on specific implementations.
   - Concrete implementations (e.g., `Scanner`, `Analyzer`, `Applier`) are **injected** into `Tracer` via `NewTracer(...)`.

---

## File structure

```text
.
├── main.go
└── internal
    ├── interfaces.go
    ├── models.go
    ├── tracer.go
    ├── scanner.go
    ├── analyzer.go
    ├── applier.go
    ├── mdparser.go
    ├── srccoverparser.go
    ├── filehash.go
    └── utils.go
```

Summary of responsibilities

- **main.go**: CLI orchestration, argument parsing, creation of `Tracer`, top-level error handling.  
- **interfaces.go**: All high-level contracts (`ITracer`, `IScanner`, `IAnalyzer`, `IApplier`, etc.).  
- **models.go**: Domain entities and data structures (`FileStructure`, `Action`, errors, coverage descriptors...).  
- **tracer.go**: Implement `ITracer`, coordinate scanning, analyzing, and applying.  
- **scanner.go**: Implement `IScanner`, discover and parse files into structured data.  
- **mdparser.go / srccoverparser.go**: Specialized parsing logic for Markdown / source coverage tags.  
- **analyzer.go**: Implement `IAnalyzer`, checks for semantic errors, determine required transformations.  
- **applier.go**: Implement `IApplier`, applly transformations to markdown and `reqmd.json`.  
- **utils.go**: Common helper functions.
- **gogit.go**: Implement IGit interface using `go-git` library.

---

## File URL construction

The system needs to construct a `FileURL` for any given `FileStructure`. A `FileURL` consists of two main components:

- Repository root folder URL (`RepoRootFolderURL`)
- Relative path (`RelativePath`)

URL structure examples:

- GitHub: `https://github.com/voedger/voedger/blob/somebranch1/pkg/api/handler_test.go`
  - `RepoRootFolderURL`: `https://github.com/voedger/voedger/blob/somebranch1`
  - `RelativePath`: `pkg/api/handler_test.go`
- GitLab: `https://gitlab.com/myorg/project/-/blob/somebranch2/src/core/processor.ts`
  - `RepoRootFolderURL`: `https://gitlab.com/myorg/project/-/blob/somebranch2`
  - `RelativePath`: `src/core/processor.ts`

### (g *git).RepoRootFolderURL() string

- `git.RepoRootFolderURL()` returns data from the field `repoRootFolderURL`.
- The system obtains data for `RepoRootFolderURL()` once during `NewIGit()` initialization
  - This is implemented as a separate `git.constructRepoRootFolderURL()` function.
- `git.constructRepoRootFolderURL()` uses:
  - RepositoryURL which  is obtained from git remote named "origin", result is like `https://github.com/voedger/voedger`
    - Command would be `git remote get-url origin` but go-git shall be used instead.
  - Actual current branch name
  - Git provider-specific path elements:
    - GitHub: `blob/<actual-branch-name>`
    - GitLab: `-/blob/<actual-branch-name>`
- Git provider is detected based on the remote URL
  - GitHub: `github.com`
  - GitLab: `gitlab.com`
- If `git.constructRepoRootFolderURL()` fails, `NewIGit()` initialization fails

### FileStructure.RelativePath construction

- `Scanner.Scan()` calculates the path using:
  - `IGit.PathToRoot()`
  - `filepath.Rel()`
- The result is stored in `FileStructure.RelativePath`

### FileStructure.RepoRootFolderURL construction

- `Scanner.Scan()` sets `FileStructure.RepoRootFolderURL` using `IGit.RepoRootFolderURL()`

### FileURL assembly: FileStructure.FileURL()

- Final `FileURL` is retured by FileStructure.FileURL() and constructed by combining:
  - `FileStructure.RepoRootFolderURL`
  - `FileStructure.RelativePath`

## Changes processing

### Problem statement

The following files may have to be changed

- reqmd.json
  - New FileURL is added
  - FileHash is updated
- Markdown files updated if
  - Coverer with new FileURL is added
  - Coverer with existing FileURL does not exist anymore
  - Coverer.FileHash is updated  
  - Some RequirementSite are BareRequirementSite and there are no new Coverers

### Data structures

Action types

- ActionAddFileURL: Add new FileURL + FileHash to reqmd.json
- ActionUpdateHash: Update FileHash in reqmd.json for a given FileURL
- ActionFootnote: Create/Update a CoverageFootnote
  - Only one action per line is allowed
- ActionUpdateStatus: Update RequirementSite.CoverageStatusWord (covered/uncvrd)
  - Only one action per line is allowed
- ActionAnnotate: Convert BareRequirementSite to annotated
  - Only one action per line is allowed

RequirementCoverage

- Site *RequirementSite
- FileStructure *FileStructure
- CurrentCoverers []*Coverer
- NewCoverers []*Coverer
- ActionFootnote *ActionFootnote
- ActionUpdateStatus *ActionUpdateStatus

### Analysis

Construction of ActionFootnote-s

- `coverages map[RequirementID]*RequirementCoverage` are calculated from all FileStructures
  - Note that NewErrDuplicateRequirementID can arise here
  - `coverages.CurrentCoverers` are constructed using FileStructure and FileStructure.CoverageFootnotes.Coverers
  - `coverages.NewCoverers` are constructed using FileStructure.CoverageTags
- foreach coverage in `coverages`
  - if sorted keys of the `coverage.currentCoverers` do not match the sorted keys of the `coverage.newCoverers` then
    - `newCf CoverageFootnote` is constructed from coverages.newCoverers, coverers are sorted by FileURL
    - string representation `newCfStr string` of newCf is constructed
    - New ActionFootnote is created using newCfStr as Data
    - `coverageStatus` is set to CoverageStatusWordUncvrd if there are no NewCoverers and to CoverageStatusWordCovered otherwise
    - New ActionUpdateStatus is created using coverageStatus as Data

Construction of ActionAnnotate-s

- foreach coverage in `coverages`
  - if coverage.ActionFootnote is nil then
    - New ActionAnnotate is created using RequirementSite as Data

### Applying

**Grouping per file**:

- Actions are grouped and applied per file to minimize I/O - so each file is loaded, processed, and saved once

**Order of processing for reqmd.json file**:

- ActionAddFileURL
- ActionUpdateHash

**Order of processing for markdown file**:

- ActionAnnotate
- ActionAddCoverer
- ActionRemoveCoverer
- ActionUpdateStatus

**Grouping per line of markdown file**:

- All actions for a single line are grouped, new resulting line structure is created and then the apporopriate line is replaced
  - Resulting line structure can be: RequirementSite, CoverageFootnote
- There is not specific order of processing for actions on a single line

**Loading**:

- Each file (if it exists) is loaded entirely into memory
- OS-specific line endings are preserved and used for writing
- No backup files are created

**Line Validation**:

- Each Action contains Line and RequirementID
- It is expected that the line with the number exists and contains the RequirementSite or CoverageFootnote with the given RequirementID
- If line number doesn't exist or RequirementID in this line doesn't match:
  - Return error
  - Stop all processing immediately

**Footnotes**:

- If the file does not end with an empty line and the original FileStructure does not have CoverageFootnotes then a new empty line is added (to separate footnotes from the rest of the file)
  - This is done when we're about to add the first footnote
- New footnotes are added at end of file
- No specific ordering of new footnotes required
- Existing footnote ordering shall be preserved

### Error Handling

**Processing Errors**:

- Processing stops immediately on first error
- Remaining actions are not processed and the caller receives an error

**Non-atomic Changes**:

- Files may be left in inconsistent state if error occurs, e.g.:
  - Partially updated footnotes
  - Missing coverage annotations
- No rollback mechanism required

### Action Structure

```go
type Action struct {
    Type            ActionType
    FileStruct      *FileStructure 
    Line            int           
    Data            string        // New content for the line
    RequirementID   string        // Line is expected to contain this RequirementID
}
```

## Implementation details

- SSH URLs (like git@github.com:org/repo.git) are not supported
- It is not necessary to define specific error types for URL construction failures
- Path are stored and processed using URL separation, on Windows initial conversion is needed.
