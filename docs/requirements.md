# Requirements for the reqmd tool

This document defines the requirements for a command-line tool that traces requirements from markdown files to their corresponding coverage in source files. The tool establishes traceability links between requirement identifiers and coverage tags, automatically generating footnotes that link requirement identifiers to coverage tags.

---

## Syntax and structure of input files

Ref. [ebnf.md](ebnf.md)

---

## Use cases

- [Installation](uc-installation.md)
- [Tracing](uc-tracing.md)

## Options

- [Ignore `.*` folders](op-ignore-dot-folders.md)
- [Ignore paths by pattern](op-ignore-paths-by-pattern.md)

---

## Syntax/semantic errors

- [Handle inconsistency between Footnote and PackageId](err-inconsistency-between-footnote-and-packageid.md)
- [internal/errors_syn.go](../internal/errors_syn.go)
- [internal/errors_sem.go](../internal/errors_sem.go)

---

## Non-functional

- [System tests](systests.md)
  - [Golden data embedding to avoid separate .md_ files](20250501-nf-golden-data-embedding.md)

## Processing requirements

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
- All files but main.go shall be in `internal` folder and its subfolders
- Design of the solution shall follow SOLID principles
  - Tracing shall be abstracted by ITracer interface, implemented by Tracer
  - All necessary interfaces shall be injected into Tracer during construction (NewTracer)
- Naming
  - Interface names shall start with `I`
  - Interface implementation names shall be deduced from the interface name by removing the I prefix and possibly lowercasing the first letter
  - All interfaces shall be defined in a separate file `interfaces.go`
  - All data structures used across the application shall be defined in the `models.go` file
- "github.com/go-git/go-git/v5" shall be used for git operations

---

## Decisions

- RequirementSiteStatus
  - `covrd/covered` denotes the covered status
  - `covrd` is used by the tool as CoverageStatusWord
  - `covered` is kept for backward compatibility
  - `uncvrd` denotes the uncovered status
  - Motivation: use short words with a high level of uniqueness for covered/uncovered status
- Separation of the `<path-to-markdowns>` and `<path-to-sources>`
  - Paths are separated to avoid modifications of sources
- SSH URLs (like git@github.com:org/repo.git) are not supported
- Commit references
  - `main/master` is used as the default reference for file URLs instead of commit hashes
  - Motivation:
    - Simplifies maintenance by eliminating the need to track file changes
    - Enables working in branches that will be squashed
    - Provides more readable and stable URLs in documentation
