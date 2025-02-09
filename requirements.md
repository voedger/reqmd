# Requirements Tracing Tool Specification

## Overview

This document defines the specifications for a command-line tool that traces requirements from Markdown files to their corresponding coverage in source files. The tool establishes traceability links between requirement identifiers and coverage tags, automatically generating footnotes that link requirement identifiers to coverage tags.

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

The tool processes input text files to obtain a list of requirements and coverage tags.

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

- RequirementID
- RequirementSite
- CoverageFootnote

RequirementID is an Identifier and looks like: `~Post.handler~` (it can also be as simple as `~SomeName~`).

RequirementSite is a RequirementID with CoverageAnnotation (CoverageAnnotation is added by the reqmd tool). An example:

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
// [~server.api.v2/Post~impl]
func handlePostRequest(w http.ResponseWriter, r *http.Request) {
    // Implementation
}

// [~server.api.v2/Post~test]
func handlePostRequestTest(t *testing.T) {
    // Test
}
```

Breakdown of the `[~server.api.v2/Post~test]`:

- `server.api.v2` is the PackageID.
- `Post` is the RequirementID.
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

Command: `go install github.com/voedger/reqmd@latest`

### Tracing

Command: `reqmd trace PathToMarkdowns {PathToClonedRepo}`

Output:

- reqmdfiles.json in PathToMarkdowns and its subdirectories if FileURLs are present in the markdown files.
- RequirementSite for each RequirementID in the markdown files.
- CoverageFootnote for each RequirementSite that has matched CoverageTags
