# reqmd requirements

## Overview

This document defines the requirements for a command-line tool that traces requirements from Markdown files to their corresponding coverage in source files. The tool establishes traceability links between requirement identifiers and coverage tags, automatically generating footnotes that link requirement identifiers to coverage tags.

## Markdown elements

- **Footnote reference**: A reference to a footnote in the markdown text.
- **Footnote label**: A label that identifies a footnote.

Example:

```markdown
This is a footnote reference[^1].

[^1]: This is a footnote label.
```

## Input files

Input files consist of markdown files and source files.

## Lexical elements

```ebnf
Name = Letter { Letter | Digit | "_" }

Identifier = Name {"." Name}

```

## Markdown files

A MarkdownFile is a text file with a `.md` extension. Each Markdown file contains a header and a body.

The header specifies a PackageID (which is an Identifier), for example:

```markdown
---
reqmd.package: system.reqs
---
```

Markdown body is a sequence of different text elements. The tool processes:

- RequirementSite, can be:
  - BareRequirementSite
  - AnnotatedRequirementSite
    - CoveredAnnotatedRequirementSite
    - UncoveredAnnotatedRequirementSite
- CoverageFootnote

```ebnf
RequirementSite = RequirementSiteID [( "covered" | "uncvrd" ) CoverageFootnoteReference] ("✅" | "❓")

RequirementSiteLabel = "`" RequirementSiteID  "`"

RequirementSiteID = "~" RequirementName "~"

RequirementName = Identifier

CoverageFootnoteReference = "[^" RequirementSiteID "]

RequirementID = PackageID "/" RequirementName

```

RequirementID shall be unique within all MarkdownFiles.

Example of a BareRequirementSite:

```markdown
- The system shall handle incoming POST requests and return an HTTP 200 response upon successful processing. `~Post.handler~`.
```

Example of an CoveredAnnotatedRequirementSite:

```markdown
- The system shall handle incoming POST requests and return an HTTP 200 response upon successful processing. `~Post.handler~`covered[^~Post.handler~]✅.
```

Example of an UncoveredAnnotatedRequirementSite:

```markdown
- The system shall handle incoming POST requests and return an HTTP 200 response upon successful processing. `~Post.handler~`uncvrd[^~Post.handler~]❓.
```

CoverageFootnote contains a CoverageFootnoteHint and an optional list of Coverers. A Coverer contains "[" CoverageLabel "]" followed by "(" CoverageURL ")". An example:

```ebnf
CoverageFootnote = "[^" RequirementSiteID "]" ":" "`[" CoverageFootnoteHint "]`" [Coverer {"," Coverer}]

Coverer = "[" CoverageLabel "]" "(" CoverageURL ")"

```

An example:

```markdown
[^~Post.handler~]: `[~server.api.v2/Post.handler~impl]`[folder1/filename1:line1:impl](CoverageURL1), [folder2/filename2:line2:test](CoverageURL2)...
```

Where:

- `~server.api.v2~impl` - CoverageFootnoteHint
- `[folder1/filename1:line1:impl](CoverageURL1)` - Coverer
  - `folder1/filename1:line1:impl` - CoverageLabel
  - `CoverageURL1` - CoverageURL
- `[folder2/filename2:line2:test](CoverageURL2)` - Coverer
  - `folder2/filename2:line2:test` - CoverageLabel
  - `CoverageURL2` - CoverageURL

CoverageURL is defined as:

```ebnf
CoverageURL  = FileURL [?plain=1] "#" CoverageArea;
```

FileURL:

```ebnf

FileURL = GitHubURL | GitLabURL

/* GitHub URL Structure */
GitHubURL           = GitHubBaseURL "/blob/" CommitHash "/" FilePath
GitHubBaseURL       = "https://github.com/" Owner "/" Repository

/* GitLab URL Structure */
GitLabURL           = GitLabBaseURL "/-/blob/" CommitHash "/" FilePath
GitLabBaseURL       = "https://gitlab.com/" Owner "/" Repository
```

An example:

```text
https://github.com/voedger/voedger/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/pkg/sys/sys.vsql#L4
```

Where:

- `https://github.com/voedger/voedger/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/pkg/sys/sys.vsql` - FileURL
- `pkg/sys/sys.vsql` - FilePath
- `979d75b2c7da961f94396ce2b286e7389eb73d75` - CommitHash
- `L4` - CoverageArea

## Source Files

Each SourceFile may contain multiple CoverageTags in its text.

CoverageTag is specified as explained in the following example:

```go
// [~server.api.v2/Post.handler~impl]
func handlePostRequest(w http.ResponseWriter, r *http.Request) {
    // Implementation
}

// [~server.api.v2/Post.handler~test]
func handlePostRequestTest(t *testing.T) {
    // Test
}
```

Breakdown of the `[~server.api.v2/Post.handler~test]`:

- `server.api.v2/Post` is RequirementID
- `server.api.v2` is the PackageID.
- `Post.handler` is the RequirementName.
- `test` is the CoverageType that is Name.

## reqmdfiles.json

This file maps FileURL to FileHash for all FileURLs found in markdown files in the folder. FileURLs must be ordered lexically to avoid merge conflicts.

Example:

```json
{
    "https://github.com/voedger/voedger/blob/main/pkg/api/handler.go": "979d75b2c7da961f94396ce2b286e7389eb73d75",
    "https://github.com/voedger/voedger/blob/main/pkg/api/handler_test.go": "845a23c8f9d6a8b7e9c2d4f5a6b7c8d9e0f1a2b3",
    "https://gitlab.com/myorg/project/-/blob/main/src/core/processor.ts": "123f45e6c7d8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
}
```

## Use Cases

### Installation

To install the tool:

```bash
go install github.com/voedger/reqmd@latest
```

### Tracing

#### SYNOPSIS

```bash
reqmd trace [-v] [-e <extensions>] <path-to-markdowns> [<path-to-sources>...]
```

#### DESCRIPTION

Analyzes markdown requirement files and their corresponding source code implementation to establish traceability links. The command processes requirement identifiers in markdown files and maps them to their implementation coverage tags in source code.

General processing rules:

- Files that are larger than 128K are skipped.
- Only source files that are tracked by git (hash can be obtained) are processed.

#### OPTIONS

- `-v`:
  - Enable verbose output showing detailed processing information.
- `-e`, `--extensions`:
  - Optional. Comma-separated list of source file extensions to process (e.g., ".go,.ts,.js").
  - When omitted, defaults to:
    ```text
    .go,.js,.ts,.jsx,.tsx,.java,.cs,.cpp,.c,.h,.hpp,.py,.rb,.php,.rs,.kt,.scala,.m,.swift,.fs,.md,.sql,.vsql
    ```
  - Extensions must include the dot prefix.  

#### ARGUMENTS

- `<path-to-markdowns>`:
  - Required. Directory containing markdown requirement files to process.

- `<path-to-sources>`:
  - Optional. One or more paths to local git repository clones containing source code with coverage tags. When omitted, only markdown parsing is performed.

#### OUTPUT FILES

- `reqmdfiles.json`:
  - Created or updated in `<path-to-markdowns>` directories when FileURLs are present. Maps FileURLs to their git hashes.

- Markdown files:
  - Updated with:
  - Coverage annotations for requirement sites
  - Coverage footnotes linking requirements to implementations

#### EXIT STATUS

- 0: Success
- 1: Syntax/Semantic errors found during scan phase or other errors have occurred

#### EXAMPLES

Process markdown files only:

```bash
reqmd trace docs/requirements/
```

Process only Go and TypeScript files:
```bash
reqmd trace -e .go,.ts docs/requirements/ src/backend/
```

Process markdown with coverage from multiple source directories:

```bash
reqmd trace docs/requirements/ src/backend/ src/frontend/
```

Process with verbose output:

```bash
reqmd trace -v docs/requirements/ src/impl/
```

#### Processing requirements

Concepts:

- Action:
  - Type: Add, Update, Delete
  - What: reqmdfiles.json, RequirementSite, CoverageFootnote
  - FilePath: path of the file where the action is performed
  - Line: line number where the action is performed
  - Data: New data
- SyntaxError:
  - See [internal/errors.go](internal/errors.go)
  - RequirementName shall be an identifier
  - RequirementSiteID shall be equals in the RequirementSite and CoverageFootnote
  - CoverageStatusWord shall be "covered" or "uncvrd"
  - // TODO One RequirementSite per line
- SemanticError:
  - RequirementID shall be unique within all MarkdownFiles
  - PackageID shall be defined if there are RequirementSites
- FileStructure:
  - Path
  - Merge of the MarkdownFile and InputFile properties

Phases:

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
