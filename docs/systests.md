# System tests

## Architecture

- System tests (**SysTests**) are located in the `internal/sys_test.go` file
- Each SysTest has a symbolic TestId
- Each SysTest is associated with a SysTestData folder located in `testdata/systest/<TestId>` folder
  - `req`: Contains TestMarkdown files, Reqmd files, and GoldenFiles
  - `src`: Contains source files
- **TestMarkdown** files contain **NormalLines** and **GoldenAnnotations**, see below
- `reqid: nf/GoldenDataEmbedding`🏷️: GoldenAnnotations represent the expected errors or transormation of the
  - obsoleted🚫: GoldenLines represent expected errors for the previous NormalLine
- SysTestData is loaded and processed by the `internal/systest/RunSysTest` function
- `RunSysTest` uses the `parseGoldenData()` function to parse the Golden Data and return a `goldenData` struct
- `RunSysTest` uses the `actualizeGoldenData()` function to replace `{{.CommitHash}}` with the actual commit hash
- `goldenData` struct contains:
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
  - Golden data can also be embedded directly in NormalFiles using GoldenAnnotations (see Golden data embedding below)

## TestMarkdown format

File structure:

```ebnf
WS               = { " " | "\t" } .
Body             = { NormalLine | GoldenAnnotation } .
GoldenAnnotation = "//" {WS} (GoldenErrors | LineMutation) .
GoldenErrors     = "errors:" {WS} {"""" ErrRegex """" {WS}} .
LineMutation       = ("line" | "line-" | "line+" | "line" | "line>>") [{WS} ":" Content] .
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

Golden annotation syntax:

- `// line-`: Remove the previous non-GoldenAnnotation line
- `// line+`: Add a line after the previous non-GoldenAnnotation line
  - Multiple statements are allowed and processed in order
- `// line`: Replace the previous non-GoldenAnnotation line
- `// line1`: Insert a line at the beginning of the file
  - Multiple statements are allowed and processed in order
- `// line>>`: Append a line at the end of the file
  - Multiple statements are allowed and processed in order
