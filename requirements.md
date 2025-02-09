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
reqmd.package: server.api.v2
---
```

Markdown body is a sequence of different text elements. The tool processes:

- BareRequirementName
- RequirementSite
- CoverageFootnote

RequirementName is an Identifier and looks like: `~Post.handler~` (it can also be as simple as `~SomeName~`).

RequirementID is formed as PackageID "/" RequirementName.

RequirementID shall be unique within all MarkdownFiles.

```markdown

BareRequirementName is a RequirementName without CoverageAnnotation.

RequirementSite is a RequirementName with CoverageAnnotation (CoverageAnnotation is added by the reqmd tool). An example:

```markdown
- APIv2 implementation shall provide a handler for POST requests. `~Post.handler~`coverage[^~Post.handler~].
```

CoverageFootnote contains a CoverageFootnoteHint and an optional list of Coverers. A Coverer contains "[" CoverageLabel "]" followed by "(" CoverageURL ")". An example:

```markdown
[^~Post.handler~]: `[~server.api.v2~impl]`[folder1/filename1:line1:impl](CoverageURL1), [folder2/filename2:line2:test](CoverageURL2)...

Where:
- [~server.api.v2~impl] - CoverageFootnoteHint
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

## Input Files

Each InputFile may contain multiple CoverageTags in its text.

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

Requirement: The tool shall support tracing of requirements and generation of coverage mapping.  
Command:

```bash
reqmd trace <path-to-markdowns> [<path-to-cloned-repo>...]
```

Where `<path-to-markdowns>` and `<path-to-cloned-repo>` are the InputFiles.

Arguments:

- `<path-to-markdowns>` (Required): Directory containing the markdown files to be processed.
- `<path-to-cloned-repo>` (Optional): Path to a local clone of the repository for additional coverage analysis.

Upon execution, the tool shall:

- Create or update `reqmdfiles.json` in the `<path-to-markdowns>` and its subdirectories
  - `reqmdfiles.json` is created/updated only if markdown files are found in the directory and they contain FileURLs.
- Convert all BareRequirementNames into corresponding RequirementSites by appending the appropriate CoverageAnnotation.
- Generate/Update CoverageFootnotes for each RequirementSite that possesses matching CoverageTags.
  - Coveres are updated only if the file addressed by the correspoding FileURL has hash different from the one in `reqmdfiles.json`.

#### Processing requirements

Concepts:

- Action:
  - Type: Add, Update, Delete
  - What: reqmdfiles.json, RequirementSite, CoverageFootnote
  - FilePath: path of the file where the action is performed
  - Line: line number where the action is performed
  - Data: New data
- SyntaxError: currently syntax errors are not defined (empty list)
- SemanticError:
  - requirement id shall be unique within all MarkdownFiles

Phases:

- Scan
  - Parse all InputFiles and generate FileStructures and the list of SyntaxErrors.
  - InputFiles shall be processed per-subfolder by the goroutines pool.
- Analyze
  - If there are SyntaxErrors the processing is stopped
  - Parse all FileStructures and generate list of SemanticErrors and list of Actions.
- Apply
  - If there are SemanticErrors the processing is stopped
  - Apply all Actions to the InputFiles.

## Construction requirements

- The tool shall be implemented in Go.
- All files but main.go shall be in `internal` folder
- File hashes shall be calculated using `git hash-object`
- Design of the solution shall follow SOLID principles
  - Tracing shall be abstracted by ITracer interface, implemented by tracer
  - All necessary intarfaces shall be injected into Tracer during construction (NewTracer)
- Naming
  - Interface names shall start with I
  - Interface implementation names shall be deduced from the interface name by removing the I prefix
  - All interfaces shall be defined in a separate file interfaces.go
  - All data structures used across the application shall be defined in thw models.go file.
