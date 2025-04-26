# reqmd requirements

This document defines the requirements for a command-line tool that traces requirements from markdown files to their corresponding coverage in source files. The tool establishes traceability links between requirement identifiers and coverage tags, automatically generating footnotes that link requirement identifiers to coverage tags.

---

## Syntax and structure of input files

Ref. [ebnf.md](ebnf.md)

---

## Use cases

- [Installation](uc-installation.md)
- [Tracing](uc-tracing.md)
- [Ignore `.*` folders](uc-ignore-dot-folders.md)

---

## Processing requirements

### Syntax/Semantic errors

- See [internal/errors.go](../internal/errors.go)

### Phases

- Scan
  - Parse all InputFiles and generate FileStructures and the list of SyntaxErrors
  - InputFiles shall be processed per-subfolder by the goroutines pool
- Analyze
  - Preconditions: there are no SyntaxErrors
  - Parse all FileStructures and generate list of SemanticErrors and list of Actions
- Apply
  - Preconditions: there are no SemanticErrors
  - Apply all Actions to the InputFiles

---

## Construction requirements

- The tool shall be implemented in Go
- All files but main.go shall be in single `internal` folder, there shall be no subfolders
- Design of the solution shall follow SOLID principles
  - Tracing shall be abstracted by ITracer interface, implemented by Tracer
  - All necessary interfaces shall be injected into Tracer during construction (NewTracer)
- Naming
  - Interface names shall start with I
  - Interface implementation names shall be deduced from the interface name by removing the I prefix
  - All interfaces shall be defined in a separate file interfaces.go
  - All data structures used across the application shall be defined in the models.go file
- "github.com/go-git/go-git/v5" shall be used for git operations

---

## Decisions

- RequirementSiteStatus
  - `covered` denotes the covered status
  - `uncvrd` denotes the uncovered status
  - Motivation: use short words with a high level of uniqueness for uncovered status
- Separation of the `<path-to-markdowns>` and `<path-to-sources>`
  - Paths are separated to avoid modifications of sources
- SSH URLs (like git@github.com:org/repo.git) are not supported
- Commit references
  - `main/master` is used as the default reference for file URLs instead of commit hashes
  - Motivation:
    - Simplifies maintenance by eliminating the need to track file changes
    - Enables working in branches that will be squashed
    - Provides more readable and stable URLs in documentation
