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
   - Verifies file hashes against reqmdfiles.json
   - Generates a list of required file modifications

3. **Apply** – Update Files
   - Updates or creates coverage footnotes
   - Appends coverage annotations to requirements
   - Maintains reqmdfiles.json for file tracking
   - Ensures changes are made only when no errors exist

The system is designed using SOLID principles:

1. **Single Responsibility Principle**  
   - Each component (scanner, analyzer, applier) is responsible for exactly one stage of the process.  
   - `mdparser.go` deals **only** with parsing Markdown; `srccoverparser.go` deals **only** with source coverage tags.

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

Explanation of each file:

### **main.go**  

- **Purpose**: CLI entry point.  
- **Responsibilities**:  
  - Parse command-line arguments (e.g., `reqmd trace <path-to-markdowns> [<path-to-cloned-repo>...]`).  
  - Construct a `Tracer` (using `NewTracer`) with the required dependencies.  
  - Invoke the high-level operations (`Scan`, `Analyze`, `Apply`).  
  - Handle application-level logging/errors.  

### **interfaces.go**  

- **Purpose**: Define all interfaces (start with `I`).  
- **Responsibilities**:  
  - `ITracer`: Orchestrates the main steps (Scan, Analyze, Apply).  
  - `IScanner`: Scans input directories/files to build a high-level model (`FileStructure` objects, coverage tags, etc.).  
  - `IAnalyzer`: Performs semantic checks, identifies required transformations, and generates a list of `Actions`.  
  - `IApplier`: Applies transformations (e.g., updating coverage footnotes, generating `reqmdfiles.json`).  

### **models.go**  

- **Purpose**: Define all data structures shared across the application.  
- **Key structures**:  
  - `FileStructure`: represents an input file, including path and parsed details (e.g., for Markdown or source).  
  - `Action`: describes a transformation: type (`Add`, `Update`, `Delete`), target file, line number, and new data.  
  - `SyntaxError`: structure containing details about syntax errors (if any).  
  - `SemanticError`: structure describing higher-level semantic issues (e.g., requirement ID collisions).  
  - `CoverageTag`: stores found coverage annotation details (e.g., requirement ID, coverage type).  
  - `CoverageFootnote`: stores the requirement footnote details including coverage references.  
  - `ReqmdfilesMap`: representation of the `reqmdfiles.json` data (mapping from `FileURL` -> `FileHash`).

### **tracer.go**  

- **Purpose**: Implements `ITracer`. This is the **facade** that orchestrates the scanning, analyzing, and applying phases.  
- **Key functions**:  
  - `Trace()`
- **Responsibilities**:  
  - High-level workflow control.  
  - Enforce the steps: if syntax errors exist, abort; if semantic errors exist, abort; otherwise apply actions.  
  - Depend on **abstractions** (`IScanner`, `IAnalyzer`, `IApplier`), not on concrete implementations.  

### **scanner.go**  

- **Purpose**: Implements `IScanner`.  
- **Key functions**:  
  - `Scan`:  
    - Recursively discover Markdown and source files.
    - Delegate parsing to specialized components (`mdparser.go`, `srccoverparser.go`).
    - Build a unified list of `FileStructure` objects for each file.
      - Files without Requirements or CoverageTags are skipped.
    - Collect any `SyntaxError`s.
- **Responsibilities**:  
  - Single responsibility: collecting raw data (files, coverage tags, requirement references) and building the domain model.  
  - Potential concurrency (goroutines) for scanning subfolders.  

### **analyzer.go**  

- **Purpose**: Implements `IAnalyzer`.  
- **Key functions**:  
  - `Analyze(files []FileStructure) (actions []Action, errs []SemanticError)`:  
    - Perform semantic checks (e.g., unique `RequirementID`s).
    - Determine which coverage footnotes need to be updated or created.  
    - Identify which bare requirement names need coverage annotations appended.  
    - Compare file hashes in `reqmdfiles.json` to actual `git hash-object` results to see if coverage references are stale.  
    - Construct the list of `Action` items describing needed transformations.
- **Responsibilities**:  
  - Single responsibility: interpret the domain data, detect required changes, produce actionable tasks.  

### **applier.go**  

- **Purpose**: Implements `IApplier`.  
- **Key functions**:  
  - `Apply(actions []Action) error`:  
    - For each `Action`, open the target file, apply the transformation (insert footnotes, update coverage lines, etc.).  
    - Update or create `reqmdfiles.json`.  
  - Ensures that changes are made **only** when no semantic or syntax errors are present.  
- **Responsibilities**:  
  - Single responsibility: physically write changes (like footnote references or coverage annotations) to the files.

### **mdparser.go**

- **Purpose**: Specialized logic for parsing Markdown files (headers, footnote references, requirement names, etc.).  
- **Responsibilities**:  
  - Single responsibility: read a Markdown file, produce domain objects (`FileStructure` details) or syntax errors.  

### **srccoverparser.go**

- **Purpose**: Specialized logic for parsing coverage tags from source files.  
- **Responsibilities**:  
  - Single responsibility: read a source file, find coverage tags, produce domain objects.  

### **filehash.go**  

- **Purpose**: Query Git for file hashes using `git hash-object`.  
- **Key functions**:  
  - `GetFileHash(path string) (string, error)`: calculates the hash.  
- **Responsibilities**:  
  - Encapsulate how file hashing is performed so it’s easy to maintain or replace.  

### **utils.go**

- **Purpose**: Collect small helper functions that do not belong in the main business logic (string manipulations, logging helpers, sorting, etc.).  
- **Responsibilities**:  
  - Common or cross-cutting concerns without creating cyclical dependencies.  

### **gogit.go**

- **Purpose**: Provide IGit interface using `go-git` library.

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

### Implementation details

- SSH URLs (like git@github.com:org/repo.git) are not supported
- It is not necessary to define specific error types for URL construction failures
- Path are stored and processed using URL separation, on Windows initial conversion is needed.

---

## Summary of Responsibilities

- **main.go**: CLI orchestration, argument parsing, creation of `Tracer`, top-level error handling.  
- **interfaces.go**: All high-level contracts (`ITracer`, `IScanner`, `IAnalyzer`, `IApplier`, etc.).  
- **models.go**: Domain entities and data structures (`FileStructure`, `Action`, errors, coverage descriptors...).  
- **tracer.go**: Implements `ITracer`, coordinates scanning, analyzing, and applying.  
- **scanner.go**: Implements `IScanner`, discovers and parses files into structured data.  
- **mdparser.go / srccoverparser.go**: Specialized parsing logic for Markdown / source coverage tags.  
- **analyzer.go**: Implements `IAnalyzer`, checks for semantic errors, determines required transformations.  
- **applier.go**: Implements `IApplier`, applies transformations to markdown and `reqmdfiles.json`.  
- **filehash.go**: Encapsulates Git-based hashing logic.  
- **utils.go**: Common helper functions.
