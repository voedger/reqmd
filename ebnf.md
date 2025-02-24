# Syntax and structure of input files

This document defines the structure and syntax for the input files processed by the reqmd tool.

## Notation

The syntax is specified using a [variant](https://en.wikipedia.org/wiki/Wirth_syntax_notation) of Extended Backus-Naur Form (EBNF).

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
  Letter         = "A" | "B" | "C" | "D" | "E" | "F" | "G" | "H" | "I" | "J" |
                  "K" | "L" | "M" | "N" | "O" | "P" | "Q" | "R" | "S" | "T" |
                  "U" | "V" | "W" | "X" | "Y" | "Z" |
                  "a" | "b" | "c" | "d" | "e" | "f" | "g" | "h" | "i" | "j" |
                  "k" | "l" | "m" | "n" | "o" | "p" | "q" | "r" | "s" | "t" |
                  "u" | "v" | "w" | "x" | "y" | "z" .

  Digit          = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" .

  Number         = Digit { Digit } .

  Name           = Letter { Letter | Digit | "_" } .
  Identifier     = Name { "." Name } .

  WS             = { " " | "\t" } .
  NewLine        = "\n" | "\r\n" .

  AnyCharacter   = ? any character ? .

  CoverageFootnoteID = {? any character but "]" ?}
```

## Markdown files

A MarkdownFile is a text file with a `.md` extension. Each Markdown file contains a header and a body.

```ebnf
MarkdownFile   = Header Body .

Header         = "---" NewLine
                 PackageDeclaration NewLine
                 "---" NewLine .
PackageDeclaration = "reqmd.package:" WS PackageID .
PackageID      = Identifier .

Body           = { MarkdownElement } .
MarkdownElement = RequirementSite | CoverageFootnote | PlainText .

PlainText      = { AnyCharacter } .
```

### Header

The header specifies a PackageID (which is an Identifier), for example:

```markdown
---
reqmd.package: system.reqs
---
```

### Content

Content of the MarkdownFiles where reqmd.package is started with "ignoreme" is ignored.

Markdown body is a sequence of different text elements. The tool processes:

- RequirementSite, can be:
  - BareRequirementSite
  - AnnotatedRequirementSite
    - CoveredAnnotatedRequirementSite
    - UncoveredAnnotatedRequirementSite
- CoverageFootnote

### RequirementSite

```ebnf
  RequirementSite = RequirementSiteLabel [ CoverageStatusWord CoverageFootnoteReference ] [ CoverageStatusEmoji ] .
  RequirementSiteLabel = "`" RequirementSiteID "`" .
  RequirementSiteID = "~" RequirementName "~" .
  CoverageStatusWord = "covered" | "uncvrd" .
  RequirementName = Identifier .
  RequirementId = PackageID "/" RequirementName .
  CoverageFootnoteReference = "[^" CoverageFootnoteID "]" .
  CoverageStatusEmoji = "✅" | "❓" .
```

Constraints:

- Only one RequirementSite is allowed per line.
- RequirementSites are not processed inside code blocks.
- Node that code block can have identation specified by spaces or a tab.
- RequirementId shall be unique within all MarkdownFiles.

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

### CoverageFootnote

CoverageFootnote contains a CoverageFootnoteHint and an optional list of Coverers. A Coverer contains "[" CoverageLabel "]" followed by "(" CoverageURL ")". An example:

```ebnf
  CoverageFootnote = "[^" CoverageFootnoteID "]" ":" "`[" CoverageFootnoteHint "]`" [Coverers] .
  CoverageFootnoteHint = "~" PackageID "/" RequirementName "~" CoverageType .
  Coverers    = Coverer { "," Coverer } .
  Coverer     = "[" CoverageLabel "]" "(" CoverageURL ")" .
  CoverageLabel = FilePath ":" Number ":" CoverageType .
  CoverageType = Name .

  CoverageURL  = FileURL [?plain=1] "#" CoverageArea .
  FileURL = GitHubURL | GitLabURL .

  GitHubURL      = GitHubBaseURL "/blob/" CommitHash "/" FilePath .
  GitHubBaseURL  = "https://github.com/" Owner "/" Repository .

  GitLabURL      = GitLabBaseURL "/-/blob/" CommitHash "/" FilePath .
  GitLabBaseURL  = "https://gitlab.com/" Owner "/" Repository .

  Owner          = Identifier .
  Repository     = Identifier .
  CommitHash     = HexDigit { HexDigit } .
  FilePath       = { AnyCharacter - ("?" | "#") } .
  CoverageArea   = "L" Number .
```

Coverers requirements:

- Coverers shall be sorted by CoverageType, then by FilePath, then by Number, then by CoverageURL.

An example:

```markdown
[^~Post.handler~]: `[~server.api.v2/Post.handler~impl]`[folder1/filename1:line1:impl](CoverageURL1), [folder2/filename2:line2:test](CoverageURL2)...
```

Where:

- "`[~server.api.v2/Post.handler~impl]`" - CoverageFootnoteHint
- `[folder1/filename1:line1:impl](CoverageURL1)` - Coverer
  - `folder1/filename1:line1:impl` - CoverageLabel
  - `CoverageURL1` - CoverageURL
- `[folder2/filename2:line2:test](CoverageURL2)` - Coverer
  - `folder2/filename2:line2:test` - CoverageLabel
  - `CoverageURL2` - CoverageURL

CoverageURL is defined as:

```ebnf
  CoverageURL  = FileURL [?plain=1] "#" CoverageArea .
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

- `server.api.v2/Post` is RequirementId
- `server.api.v2` is the PackageID.
- `Post.handler` is the RequirementName.
- `test` is the CoverageType that is Name.

## reqmd.json

- reqmd.json contains information necessary for the tool to process the requirements
- reqmd.json files are updated by the tool during the processing of the requirements and shall be committed to the repository
- This file is create per folder if markdown files are present and contains RequirementSites
- Processing shall survive the deletion of the reqmd.json file and missing FileURLs

Structure

- FileURL2FileHash
  - Maps FileURL to FileHash
  - FileURLs shall be ordered lexically to avoid unnecessary changes and merge conflicts

Example:

```json
{
  "FileHashes" : {
    "https://github.com/voedger/voedger/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/pkg/api/handler.go": "979d75b2c7da961f94396ce2b286e7389eb73d75",
    "https://github.com/voedger/voedger/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/pkg/api/handler_test.go", "845a23c8f9d6a8b7e9c2d4f5a6b7c8d9e0f1a2b3", 
    "https://gitlab.com/myorg/project/-/blob/979d75b2c7da961f94396ce2b286e7389eb73d75/src/core/processor.ts", "123f45e6c7d8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
  }
}
```
