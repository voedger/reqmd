# Design

Design of the reqmd tool.

## Overview

The solution is split into a CLI entry point (`main.go`) and an internal package (`internal/`). Within `internal/`, each file, structure, and function has a focused responsibility. All data structures are centralized in `models.go`, and all interfaces in `interfaces.go`. Implementations are named by removing the `I` prefix.

The tool follows a three-stage pipeline architecture:

1. **Scan** – Build the Domain Model
   - Recursively discovers Markdown and source files from multiple root paths
   - Each path can contain both markdown and source files
   - Extracts requirement references from Markdown files
   - Identifies coverage tags in source code
   - Builds file structures with Git metadata
   - Collects syntax errors during parsing

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
   - Makes changes only when no errors exist

The system is designed using SOLID principles:

1. **Single Responsibility Principle**  
   - Each component (scanner, analyzer, applier) is responsible for exactly one stage of the process.  

2. **Open/Closed Principle**  
   - New features can be added by creating new parsers or new analysis rules without modifying existing, stable components.
   - For example, if a new coverage system is added, you can create a new parser that returns coverage tags in the same data model.

3. **Liskov Substitution Principle**
   - Interfaces (`IScanner`, `IAnalyzer`, `IApplier`) can be replaced with new implementations as long as they respect the same contracts.

4. **Interface Segregation Principle**
   - Instead of having a single "monolithic" interface, smaller interfaces reflect the actual steps: scanning, analyzing, applying.
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
    ├── fprocessor.go
    ├── fileparser.go
    ├── fileparser_md.go
    ├── fileparser_src.go
    ├── analyzer.go
    ├── applier.go
    ├── errors.go
    ├── utils.go
    ├── gogit.go
    └── main.go
```

Summary of responsibilities

- **main.go** (root): CLI entry point, argument parsing, package initialization
- **internal/main.go**: Internal CLI orchestration, implementation details for commands
- **interfaces.go**: All high-level contracts (`ITracer`, `IScanner`, `IAnalyzer`, `IApplier`, etc.)
- **models.go**: Domain entities and data structures (`FileStructure`, `Action`, coverage descriptors...)
- **errors.go**: Error types, constructors and handlers for both syntax and semantic errors
- **tracer.go**: Implement `ITracer`, coordinate scanning, analyzing, and applying
- **scanner.go**: Implement `IScanner`, discover and parse files from multiple root paths into structured data
- **fprocessor.go**: Provides concurrent file system scanning functionality with worker pools, breadth-first directory traversal, and error handling
- **fileparser.go**: Handles general file parsing operations for both markdown and source files
- **fileparser_md.go**: Specialized parsing logic for Markdown files
- **fileparser_src.go**: Specialized parsing for source files
- **analyzer.go**: Implement `IAnalyzer`, checks for semantic errors, determine required transformations
- **applier.go**: Implement `IApplier`, apply transformations to markdown and `reqmd.json`
- **utils.go**: Common helper functions
- **gogit.go**: Implement IGit interface using `go-git` library

---

## File URL construction

The system constructs a `FileURL` for any given `FileStructure`. A `FileURL` consists of two main components:

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

- `git.RepoRootFolderURL()` returns data from the `repoRootFolderURL` field
- The system obtains data for `RepoRootFolderURL()` during `NewIGit()` initialization
  - This is implemented as a separate `git.constructRepoRootFolderURL()` function
- `git.constructRepoRootFolderURL()` uses:
  - RepositoryURL obtained from git remote named "origin" (e.g., `https://github.com/voedger/voedger`)
  - Current branch name
  - Git provider-specific path elements:
    - GitHub: `blob/<commit-hash>`
    - GitLab: `-/blob/<commit-hash>`
- Git provider is determined based on the remote URL
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

- Final `FileURL` is returned by FileStructure.FileURL() and constructed by combining:
  - `FileStructure.RepoRootFolderURL`
  - `FileStructure.RelativePath`

## Analyse and Apply

### Problem statement

The following files may require changes:

- reqmd.json is updated when:
  - FileUrl is added/removed
  - FileHash is updated
- Markdown files are updated when:
  - A Coverer with new FileURL is added
  - A Coverer with existing FileURL no longer exists
  - Coverer.FileHash is updated  
  - Some RequirementSites are BareRequirementSites and have no new Coverers

### Analyze Reqmdjson actions

Principles:

- If a folder has any requirement with changed footnotes, the whole folder's reqmd.json needs updating
- As a result reqmd.json can be empty, in this case applier shall make an attempt to delete it, if exists
- FileUrl() helper function is used to strip line numbers from CoverageURLs

### Apply markdown actions

#### Load file

- Each file (if it exists) is loaded entirely into memory
- OS-specific line endings are preserved when writing
- No backup files are created

#### Line Validation for markdown files

- There are two Actions: ActionFootnote and ActionSite
- Each Action contains Line and RequirementId
- RequirementSiteRegex and CoverageFootnoteRegex from models.go are used to match lines with RequirementId
- Note that RequirementId is unique within all markdown files
- If Action.Line > 0
  - It is expected that the line with the number exists and contains the RequirementSite or CoverageFootnote with the given RequirementId
  - RequirementSiteRegex or CoverageFootnoteRegex is used to validate the line
  - If line number doesn't exist or RequirementId in this line doesn't match:
    - Return error
    - Stop all processing immediately
  - Only the regex-matched part of the line is replaced; the rest is preserved
- Otherwise:
  - The action is CoverageFootnote and the line is appended to the file end (see Footnotes)

#### Footnotes

Strip trailing empty lines:

- All trailing empty lines are removed.

Add an empty line:

- If there are new footnotes (ActionFootnote.Line = 0) and the last line of the file is not a footnote then an empty line is added to separate footnotes from the rest of the file

Process ActionFootnotes:

- Existing footnotes are updated with new content
- New footnotes (ActionFootnote.Line = 0) are added at end of the list with no extra blank lines in between
- No specific ordering of new footnotes required
- Existing footnote ordering shall be preserved
- If there new footnotes then an empty line is added at the end of the file

### Apply Regmdjson actions

- If Reqmdjson is empty and the file specified by FilePath exists then it is deleted
- Else Reqmdjson is jsonized with indentation and written to the file specified by FilePath

### Error handling

**Processing errors**:

- Processing stops immediately on first error
- Remaining actions are not processed and the caller receives an error

**Non-atomic changes**:

- Files may be left in inconsistent state if error occurs, e.g.:
  - Partially updated footnotes
  - Missing coverage annotations
- No rollback mechanism required

---

## System tests

### System tests. Core principles

- System tests (SysTests) are located in the `internal/sys_test.go` file
- Each SysTest has a symbolic TestID
- Each SysTest is associated with a  SysTestData folder located in `testdata/systest/<TestID>` folder
  - `req`: TestMarkdown-s, Reqmd-s, GoldenFile-s
  - `src`: Source files
- TestMarkdown file contains NormalLines and GoldenLines (for errors), see below
- GoldenLines represent expected errors for the previous NormalLine
- SysTestData are loaded and processed by the `internal/systest/RunSysTest` function
- `RunSysTest` uses `parseGoldenData()` function to parse the Golden Data and returns `goldenData` struct
- `RunSysTest` uses `actualizeGoldenData()` function to replace `{{.CommitHash}}` with actual commit hash
- `goldenData` struct
  - `errors map[Path]map[int][]*regexp.Regexp` - expected errors (compiled regexes)
  - `lines map[Path][]string` - expected lines
- `parseGoldenData`
  - Definitions
    - GoldenFile is a file whose path ends with "_", e.q. `req.md_`
    - NormalFile is a file whose path does not end with "_"
    - NormalizedPath is the path with "_" removed from the path
  - Takes the path to the `req` folder as a parameter
  - NormalFiles that ends with ".md" are processed to extract GoldenErrors (see below)
  - NormalFiles that do not have GoldenFile counterpart are loaded to goldenData.lines
  - GoldenFiles are loaded to goldenData.lines, path is normalized ("_" is removed)
  - Processing of goldenData.lines:
    - For each NormalFile without a GoldenFile counterpart, read the file content line by line and store in goldenData.lines[normalizedPath]
    - For each GoldenFile, read the file content line by line and store in goldenData.lines[normalizedPath]
    - The lines are stored in the same order as they appear in the file
    - Empty lines and whitespace are preserved exactly as they appear in the files

### TestMarkdown

File structure:

```ebnf
WS       = { " " | "\t" } .
Body     = { NormalLine | GoldenLine} .
GoldenLine = "//" {WS} (GoldenErrors) .
GoldenErrors = "errors:" {WS} {"""" ErrRegex """" {WS}} .
```

Specification:

- GoldenErrors represent the expected errors for the previous line
- ErrRegex is a regular expression that matches the error message
  - If line is related to multiple errors then multiple ErrRegexes are used
  - errors and ErrRegexes are related one-to-one

---

## Implementation details

- SSH URLs (e.g., git@github.com:org/repo.git) are not supported
- Specific error types for URL construction failures are not required
- Paths are stored and processed using URL separators; Windows paths need initial conversion
