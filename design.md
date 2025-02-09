# Design

Design of the reqmd tool.

## Overview

Below is the desgin that follows **SOLID** principles. The solution is split into a CLI entry point (`main.go`) and an internal package (`internal/`). Within `internal/`, each file, structure, and function has a focused responsibility. All data structures are centralized in `models.go`, and all interfaces in `interfaces.go`. Implementations are named by removing the `I` prefix.

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

1. **main.go**  
   - **Purpose**: CLI entry point.  
   - **Responsibilities**:  
     - Parse command-line arguments (e.g., `reqmd trace <path-to-markdowns> [<path-to-cloned-repo>...]`).  
     - Construct a `Tracer` (using `NewTracer`) with the required dependencies.  
     - Invoke the high-level operations (`Scan`, `Analyze`, `Apply`).  
     - Handle application-level logging/errors.  

2. **interfaces.go**  
   - **Purpose**: Define all interfaces (start with `I`).  
   - **Responsibilities**:  
     - `ITracer`: Orchestrates the main steps (Scan, Analyze, Apply).  
     - `IScanner`: Scans input directories/files to build a high-level model (`FileStructure` objects, coverage tags, etc.).  
     - `IAnalyzer`: Performs semantic checks, identifies required transformations, and generates a list of `Actions`.  
     - `IApplier`: Applies transformations (e.g., updating coverage footnotes, generating `reqmdfiles.json`).  
     - Any additional smaller interfaces if needed (e.g., `IMarkdownParser`, `ISourceCoverageParser`), or you can keep these as separate or internal to the scanner if you wish.  

3. **models.go**  
   - **Purpose**: Define all data structures shared across the application.  
   - **Key structures**:  
     - `FileStructure`: represents an input file, including path and parsed details (e.g., for Markdown or source).  
     - `Action`: describes a transformation: type (`Add`, `Update`, `Delete`), target file, line number, and new data.  
     - `SyntaxError`: structure containing details about syntax errors (if any).  
     - `SemanticError`: structure describing higher-level semantic issues (e.g., requirement ID collisions).  
     - `CoverageTag`: stores found coverage annotation details (e.g., requirement ID, coverage type).  
     - `CoverageFootnote`: stores the requirement footnote details including coverage references.  
     - `ReqmdfilesMap`: representation of the `reqmdfiles.json` data (mapping from `FileURL` -> `FileHash`).  
     - (Optional) Additional smaller structs like `RequirementID`, `PackageID`, `RequirementName`, etc., if you want to model them more explicitly.  

4. **tracer.go**  
   - **Purpose**: Implements `ITracer`. This is the **facade** that orchestrates the scanning, analyzing, and applying phases.  
   - **Key functions**:  
     - `NewTracer(scanner IScanner, analyzer IAnalyzer, applier IApplier) *Tracer`: constructor to inject dependencies.  
     - `Scan(paths []string) ([]FileStructure, []SyntaxError)`: orchestrates scanning by delegating to `IScanner`.  
     - `Analyze(files []FileStructure) ([]Action, []SemanticError)`: delegates to `IAnalyzer`.  
     - `Apply(actions []Action) error`: delegates to `IApplier`.  
   - **Responsibilities**:  
     - High-level workflow control.  
     - Enforce the steps: if syntax errors exist, abort; if semantic errors exist, abort; otherwise apply actions.  
     - Depend on **abstractions** (`IScanner`, `IAnalyzer`, `IApplier`), not on concrete implementations.  

5. **scanner.go**  
   - **Purpose**: Implements `IScanner`.  
   - **Key functions**:  
     - `Scan(paths []string) ([]FileStructure, []SyntaxError)`:  
       - Recursively discover Markdown and source files.  
       - Delegate parsing to specialized components (`mdparser.go`, `srccoverparser.go`).  
       - Build a unified list of `FileStructure` objects for each file.  
       - Collect any `SyntaxError`s.  
   - **Responsibilities**:  
     - Single responsibility: collecting raw data (files, coverage tags, requirement references) and building the domain model.  
     - Potential concurrency (goroutines) for scanning subfolders.  

6. **analyzer.go**  
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

7. **applier.go**  
   - **Purpose**: Implements `IApplier`.  
   - **Key functions**:  
     - `Apply(actions []Action) error`:  
       - For each `Action`, open the target file, apply the transformation (insert footnotes, update coverage lines, etc.).  
       - Update or create `reqmdfiles.json`.  
     - Ensures that changes are made **only** when no semantic or syntax errors are present.  
   - **Responsibilities**:  
     - Single responsibility: physically write changes (like footnote references or coverage annotations) to the files.  

8. **mdparser.go** (optional name—could be integrated into `scanner.go` if you prefer fewer files)  
   - **Purpose**: Specialized logic for parsing Markdown files (headers, footnote references, requirement names, etc.).  
   - **Key functions** (example):  
     - `ParseMarkdown(path string) (MarkdownFileStructure, []SyntaxError)`: parse package headers, requirement names, footnotes.  
   - **Responsibilities**:  
     - Single responsibility: read a Markdown file, produce domain objects (`FileStructure` details) or syntax errors.  

9. **srccoverparser.go** (optional name—could be integrated into `scanner.go` if you prefer fewer files)  
   - **Purpose**: Specialized logic for parsing coverage tags from source files.  
   - **Key functions**:  
     - `ParseSourceCoverage(path string) (SourceFileStructure, []SyntaxError)`: parse coverage tags like `[~server.api.v2/Post.handler~test]`.  
   - **Responsibilities**:  
     - Single responsibility: read a source file, find coverage tags, produce domain objects.  

10. **filehash.go**  
    - **Purpose**: Query Git for file hashes using `git hash-object`.  
    - **Key functions**:  
      - `GetFileHash(path string) (string, error)`: calculates the hash.  
    - **Responsibilities**:  
      - Encapsulate how file hashing is performed so it’s easy to maintain or replace.  

11. **utils.go**  
    - **Purpose**: Collect small helper functions that do not belong in the main business logic (string manipulations, logging helpers, sorting, etc.).  
    - **Responsibilities**:  
      - Common or cross-cutting concerns without creating cyclical dependencies.  

---

## How SOLID principles are applied

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
