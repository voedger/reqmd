# System tests

## Architecture

- System tests (SysTests) are located in the `internal/sys_test.go` file
- Each SysTest has a symbolic TestId
- Each SysTest is associated with a SysTestData folder located in `testdata/systest/<TestId>` folder
  - `req`: Contains TestMarkdown files, Reqmd files, and GoldenFiles
  - `src`: Contains source files
- TestMarkdown files contain NormalLines and GoldenLines, see below
- `reqid: nf/GoldenDataEmbedding`🏷️: GoldenLines represent the expected errors for the previous line
  - 🚫: GoldenLines represent expected errors for the previous NormalLine
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

## TestMarkdown format

File structure:

```ebnf
WS           = { " " | "\t" } .
Body         = { NormalLine | GoldenLine } .
GoldenLine   = "//" {WS} (GoldenErrors) .
GoldenErrors = "errors:" {WS} {"""" ErrRegex """" {WS}} .
```

Specification:

- GoldenErrors represent the expected errors for the previous line
- ErrRegex is a regular expression that matches the error message
  - If a line is related to multiple errors, then multiple ErrRegexes are used
  - Line errors and ErrRegexes have a one-to-one relationship
