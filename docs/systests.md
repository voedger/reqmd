# System tests

## Definitions

- **GoldenData**: A set of expected errors and lines that are used to validate the output of a SysTest

## Architecture

- System tests (**SysTests**) are located in the `internal/sys_test.go` file
- Each SysTest has a symbolic TestId
- Each SysTest is associated with a SysTestData folder located in `testdata/systest/<TestId>` folder
  - `req`: Contains TestMarkdown files, Reqmd files, and GoldenFiles
  - `src`: Contains source files
- **TestMarkdown** files contain **NormalLines** and **GoldenAnnotations**
- GoldenAnnotations represent the expected errors or transformation of TestMarkdown lines to GoldenData lines
  - `covtag: nf/GoldenDataEmbedding`🏷️
  - obsoleted🚫: GoldenLines represent expected errors for the previous NormalLine
- **SysTestData** is loaded and processed by the `~RunSysTest~` function of the internal/systest package
- `~RunSysTest~`: Function
  - `~parseGoldenData~`: Parses the Golden Data and returns a `goldenData` struct
  - `~actualizeGoldenData~`: Replaces `{{.CommitHash.}}` with the actual commit hash
  - `~applyGoldenAnnotations~`: Applies GoldenAnnotations to the NormalLines
    - `covtag: nf/GoldenDataEmbedding`🏷️
- `~goldenData~`: A struct:
  - `errors map[Path]map[int][]*regexp.Regexp` - expected errors (compiled regexes)
  - `lines map[Path][]string` - expected lines
- `parseGoldenData`:
  - Definitions:
    - GoldenFile is a file whose path ends with "_", e.g., `req.md_`
    - NormalFile is a file whose path does not end with "_"
    - NormalizedPath is the path with "_" removed
  - Takes the path to the `req` folder as a parameter
  - NormalFiles that end with ".md" are processed to extract GoldenErrors (see below)
  - NormalFiles that do not have a GoldenFile counterpart are loaded to goldenData.lines
  - GoldenFiles are loaded to goldenData.lines, path is normalized ("_" is removed)
  - Processing of goldenData.lines:
    - For each NormalFile without a GoldenFile counterpart, read the file content line by line and store in goldenData.lines[normalizedPath]
    - For each GoldenFile, read the file content line by line and store in goldenData.lines[normalizedPath]
    - The lines are stored in the same order as they appear in the file
    - Empty lines and whitespace are preserved exactly as they appear in the files

## TestMarkdown format

File structure:

```ebnf
WS               = { " " | "\t" } .
Body             = { NormalLine | GoldenAnnotation } .
GoldenAnnotation = ">" {WS} (GoldenErrors | LineMutation) .
GoldenErrors     = "errors:" {WS} {"""" ErrRegex """" {WS}} .
LineMutation       = ("replace" | "delete" | "insert" | "firstline" | "lastilne" ) [{WS} ":" Content] .
```

Specification:

- GoldenErrors represent the expected errors for the previous line
- ErrRegex is a regular expression that matches the error message
  - If a line is related to multiple errors, then multiple ErrRegexes are used
  - Line errors and ErrRegexes have a one-to-one relationship

## Golden data embedding

Instead of maintaining separate golden files (with `_` suffix), golden data can be embedded directly in the source markdown files using specially formatted comments:

```markdown
`~REQ001~`
// line: `~REQ001~`covered✅

This line is expected to be removed
// line-
```

LineMutations:

- `delete`: Delete the previous non-GoldenAnnotation line
- `deletelast`: Delete the last non-GoldenAnnotation line
- `insert`: Add a line after the previous non-GoldenAnnotation line
  - Multiple statements are allowed and processed in order
- `replace`: Replace the previous non-GoldenAnnotation line
- `firstline`: Insert a line at the beginning of the file
  - Multiple statements are allowed and processed in order
- `append`: Append a line at the end of the file
  - Multiple statements are allowed and processed in order
