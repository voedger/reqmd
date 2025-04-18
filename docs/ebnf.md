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

CoverageFootnote contains a CoverageFootnoteHint and an optional list of Coverers. A Coverer contains "[" CoverageLabel "]" followed by "(" CoverageURL ")".

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

An example of CoverageURL:

```text
https://github.com/voedger/voedger/blob/main/pkg/sys/sys.vsql#L4
```

Where:

- `https://github.com/voedger/voedger/blob/main/pkg/sys/sys.vsql` - FileURL
- `pkg/sys/sys.vsql` - FilePath
- `main` - CommitRef
- `L4` - CoverageArea  

Syntax:
```ebnf
  CoverageFootnote = "[^" CoverageFootnoteID "]" ":" "`[" CoverageFootnoteHint "]`" [Coverers] .
  CoverageFootnoteHint = "~" PackageID "/" RequirementName "~" CoverageType .
  Coverers    = Coverer { "," Coverer } .
  Coverer     = "[" CoverageLabel "]" "(" CoverageURL ")" .
  CoverageLabel = FilePath ":" Number ":" CoverageType .
  CoverageType = Name .

  CoverageURL  = FileURL [?plain=1] "#" CoverageArea .
  FileURL = GitHubURL | GitLabURL .

  GitHubURL      = GitHubBaseURL "/blob/" CommitRef "/" FilePath .
  GitHubBaseURL  = "https://github.com/" Owner "/" Repository .

  GitLabURL      = GitLabBaseURL "/-/blob/" CommitRef "/" FilePath .
  GitLabBaseURL  = "https://gitlab.com/" Owner "/" Repository .

  Owner          = Identifier .
  Repository     = Identifier .
  CommitRef      = "main" | "master" | BranchName | CommitHash .
  BranchName     = Name { Name | Digit | "-" | "_" | "/" } .
  CommitHash     = HexDigit { HexDigit } .
  HexDigit       = Digit | "a" | "b" | "c" | "d" | "e" | "f" | "A" | "B" | "C" | "D" | "E" | "F" .
  FilePath       = { AnyCharacter - ("?" | "#") } .
  CoverageArea   = "L" Number .
```

**Note:** When generating URLs, the tool uses `main/master` as the default CommitRef. This ensures that links remain valid even when branches are squashed and simplifies tracking across repositories.

Requirements:

- Coverers shall be sorted by CoverageType, then by FilePath, then by Number, then by CoverageURL
- If the generated footnote is the first one in the file, an empty line shall be added before it
- CoverageFootnoteId for new footnotes are calculated starting from maxFootnoteIntId
- maxFootnoteIntId is the maximum integer value of all CoverageFootnoteIds mentioned in RequirementSites and CoverageFootnotes
- New CoverageFootnotes shall be added in the order of the appearance of the appropriate RequirementSites

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

```ebnf
  (* 
    Source Files 
    ------------ 
    A source file is a text file that may contain one or more CoverageTags. 
    A CoverageTag is written as a bracketed expression which links a requirement (by its package and name) 
    to a coverage type. 
    
    For example: 
        [~server.api.v2/Post.handler~test] 
    Here, PackageID is "server.api.v2", RequirementName is "Post.handler", and CoverageType is "test". 
  *)

  SourceFile   = { SourceElement } ;
  SourceElement = CoverageTag | PlainText ;
  CoverageTag  = "[" "~" PackageID "/" RequirementName "~" CoverageType "]";
```