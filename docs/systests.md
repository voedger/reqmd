# System tests

## Architecture

- System tests (SysTests) are located in the `internal/sys_test.go` file
- Each SysTest has a symbolic TestId
- Each SysTest is associated with a  SysTestData folder located in `testdata/systest/<TestId>` folder
  - `req`: TestMarkdown-s, Reqmd-s, GoldenFile-s
  - `src`: Source files
- TestMarkdown file contains NormalLines and GoldenLines (for errors), see below
- GoldenLines represent expected errors for the previous NormalLine
- SysTestData are loaded and processed by the `internal/systest/RunSysTest` function
- `RunSysTest` uses `parseGoldenData()` function to parse the Golden Data and returns `goldenData` struct
- `RunSysTest` uses `actualizeGoldenData()` function to replace `{{.CommitHash}}` with actual commit hash
- `goldenData` struct
  - `errors map[Path]map[int][]*regexp.Regexp` - expected errors (compiled regexes)
  - `lines map[Path][]string` - expected lines
- `parseGoldenData`
  - Definitions
    - GoldenFile is a file whose path ends with "_", e.q. `req.md_`
    - NormalFile is a file whose path does not end with "_"
    - NormalizedPath is the path with "_" removed from the path
  - Takes the path to the `req` folder as a parameter
  - NormalFiles that ends with ".md" are processed to extract GoldenErrors (see below)
  - NormalFiles that do not have GoldenFile counterpart are loaded to goldenData.lines
  - GoldenFiles are loaded to goldenData.lines, path is normalized ("_" is removed)
  - Processing of goldenData.lines:
    - For each NormalFile without a GoldenFile counterpart, read the file content line by line and store in goldenData.lines[normalizedPath]
    - For each GoldenFile, read the file content line by line and store in goldenData.lines[normalizedPath]
    - The lines are stored in the same order as they appear in the file
    - Empty lines and whitespace are preserved exactly as they appear in the files

## TestMarkdown format

File structure:

```ebnf
WS       = { " " | "\t" } .
Body     = { NormalLine | GoldenLine} .
GoldenLine = "//" {WS} (GoldenErrors) .
GoldenErrors = "errors:" {WS} {"""" ErrRegex """" {WS}} .
```

Specification:

- GoldenErrors represent the expected errors for the previous line
- ErrRegex is a regular expression that matches the error message
  - If line is related to multiple errors then multiple ErrRegexes are used
  - errors and ErrRegexes are related one-to-one

