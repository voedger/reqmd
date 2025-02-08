# Requirements Tracing Tool Specification

## Overview

This document defines the specifications for a command-line tool that traces requirements from Markdown files to their corresponding coverage in source files. The tool establishes traceability links between requirement identifiers and coverage tags, automatically generating footnotes that links requirement identifiers and coverage tags.

## Markdown elements

- **Footnote reference**: A reference to a footnote in the markdown text.
- **Footnote label**: A label that identifies a footnote.

Example:

```markdown
This is a footnote reference[^1].

[^1]: This is a footnote label.
```

## Input files

Input files are markdown files and source files.

The tool processes input text files and obtain list of requirements and coverage tags.

## Lexical elements

```ebnf
Name = Letter { Letter | Digit | "_" }

Identifier = Name {"." Name}
```

## Markdown files

MarkdownFile is a text file with `.md` extension. Markdown file contains a header and a body.

Header specifies PackageID (which is Identifier), an example:

```markdown
---
reqmd.package: server.api.v2
---
```

Markdown body is a sequence of different text elements, the tool processes:

- RequirementID
- RequirementSite
- CoverageFootnote

RequirementID is Identifier and looks like: `~Post.handler~` (of course may be just `~SomeName~`).

RequirementSite is RequirementID with CoverageAnnotation (CoverageAnnotation is added by the reqmd tool). An example:

```markdown
- APIv2 implementation shall provide a handler for POST requests. `~Post.handler~`coverage[^~Post.handler~].
```

CoverageFootnote contains CoverageFootnoteHint and optional list of Coverers. Coverer contains "[" CoverageLabel "]"  followed by "(" CoverageURL ")". An example:

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

Each InpitFile contains multiple CoverageTag in the text.

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
- `test` is the CoverageType.

CoverageType is Name

## reqmdfiles.json

This file lists 

## Processing requirements

## Tracing mechanism

### Concepts

RequirementIdentifierEntry:

- RequrementIdentifier
- Location

RequirementIdentifierEntry:

- RequrementIdentifier
- Location

Location

- Repository URL
- Path in Repository
- Hash

The tool follows these steps to establish traceability.

### Tracing

- Scan all `InputFile`s and identify `RequirementIdentifierEntry`s and `CoverageTagEntry`s


1. **Scans** all specified files to detect requirement identifiers and coverage tags.
2. For each **requirement identifier**:
   - Identifies all corresponding coverage tags.
   - Generates footnotes referencing each coverage.
   - Creates links within the footnotes pointing to the corresponding coverage locations.

## Output Format

The tool produces structured documentation with the following characteristics:

- Each requirement may have multiple footnotes.
- Each footnote corresponds to a single coverage tag.
- Footnotes contain clickable links to the respective coverage locations.

## EBNF

Core:

