# reqmd requirements

## Introduction

This document defines the requirements for a command-line tool that traces requirements from markdown files to their corresponding coverage in source files. The tool establishes traceability links between requirement identifiers and coverage tags, automatically generating footnotes that link requirement identifiers to coverage tags.

## Syntax and structure of input files

Ref. [ebnf.md](ebnf.md)

## Use cases

### Installation

To install the tool:

```bash
go install github.com/voedger/reqmd@latest
```

### Ignore `.*` folders

Folder with names that start with a `.` are ignored.

### Tracing

#### SYNOPSIS

```bash
reqmd [-v] trace [ (-e | --extensions) <extensions>] [--dry-run | -n] <paths>...
```

#### DESCRIPTION

Analyzes markdown requirement files and their corresponding source code implementation to establish traceability links. The command processes requirement identifiers in markdown files and maps them to their implementation coverage tags in source code.

General processing rules:

- Files that are larger than 128K are skipped
- Only source files that are tracked by git (hash can be obtained) are processed
- Each path can contain both markdown and source files
- Multiple paths can be specified to process different parts of a repository

#### OPTIONS

- `-v`:
  - Enable verbose output showing detailed processing information.
- `-e`, `--extensions`:
  - Comma-separated list of source file extensions to process (e.g., ".go,.ts,.js").
  - When omitted, defaults to:
    ```text
    .go,.js,.ts,.jsx,.tsx,.java,.cs,.cpp,.c,.h,.hpp,.py,.rb,.php,.rs,.kt,.scala,.m,.swift,.fs,.md,.sql,.vsql
    ```
  - Extensions must include the dot prefix.  
- `-n`, `--dry-run`:
  - Perform a dry run without modifying files.

#### ARGUMENTS

- `<paths>`:
  - One or more paths to process. Each path can contain both markdown requirement files and source code with coverage tags
  - At least one path must be provided
  - When multiple paths are provided, they are processed in sequence

#### OUTPUT FILES

- `reqmdfiles.json`:
  - Created or updated in each directory containing markdown files when FileURLs are present
  - Maps FileURLs to their git hashes

- Markdown files:
  - Updated with:
  - Coverage annotations for requirement sites
  - Coverage footnotes linking requirements to implementations

- Error handling
  - Files may be left in inconsistent state if error occurs, e.g.:
    - Partially updated footnotes
    - Missing coverage annotations
  - No rollback mechanism is provided

#### EXIT STATUS

- 0: Success
- 1: Syntax/Semantic errors found during scan phase or other errors have occurred

#### EXAMPLES

Process a single directory containing both markdown and source files:

```bash
reqmd trace project/
```

Process multiple directories with mixed content:

```bash
reqmd trace docs/ src/ tests/
```

Process only Go and TypeScript files in multiple directories:

```bash
reqmd trace -e .go,.ts docs/ src/ tests/
```

Process with verbose output:

```bash
reqmd trace -v docs/ src/ tests/
```

## Processing requirements

### Syntax errors

- See [internal/errors.go](../internal/errors.go)
- RequirementName shall be an identifier
- Opening fence found without matching closing fence
  - Message includes line information about the opening fence

### Semantic error

- Duplicate RequirementId detected
  - Message should include information about the files where the duplicates are found.
- Markdown file with RequirementSites shall define PackageID
  - Message inclides line information about the first RequirementSite

### Phases

- Scan
  - Parse all InputFiles and generate FileStructures and the list of SyntaxErrors.
  - InputFiles shall be processed per-subfolder by the goroutines pool.
- Analyze
  - Preconditions: there are no SyntaxErrors
  - Parse all FileStructures and generate list of SemanticErrors and list of Actions.
- Apply
  - Preconditions: there are no SemanticErrors
  - Apply all Actions to the InputFiles.

## Construction requirements

- The tool shall be implemented in Go.
- All files but main.go shall be in single `internal` folder, there shall be no subfolders.
- File hashes shall be calculated using `git hash-object`
- Design of the solution shall follow SOLID principles
  - Tracing shall be abstracted by ITracer interface, implemented by Tracer
  - All necessary intarfaces shall be injected into Tracer during construction (NewTracer)
- Naming
  - Interface names shall start with I
  - Interface implementation names shall be deduced from the interface name by removing the I prefix
  - All interfaces shall be defined in a separate file interfaces.go
  - All data structures used across the application shall be defined in the models.go file.
- "github.com/go-git/go-git/v5" shall be used for git operations

## Decisions

- RequirementSiteStatus
  - `covered` denotes the covered status.
  - `uncvrd` denotes the uncovered status.
  - Motivation: use short words with a high level of uniqueness for uncovered status.
- Separation of the `<path-to-markdowns>` and `<path-to-sources>`
  - Paths are separated to avoid modificationd of sources
- SSH URLs (like git@github.com:org/repo.git) are not supported
- Commit references
  - `main/master` is used as the default reference for file URLs instead of commit hashes
  - Motivation: 
    - Simplifies maintenance by eliminating the need to track file changes
    - Enables working in branches that will be squashed
    - Provides more readable and stable URLs in documentation
