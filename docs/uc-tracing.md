# Tracing

## Functional design

### SYNOPSIS

```bash
reqmd [-v] trace [ (-e | --extensions) <extensions>] [--dry-run | -n] <paths>...
```

### DESCRIPTION

Analyzes markdown requirement files and their corresponding source code implementation to establish traceability links. The command processes requirement identifiers in markdown files and maps them to their implementation coverage tags in source code.

General processing rules:

- Files that are larger than 128K are skipped
- Only source files that are tracked by git (hash can be obtained) are processed
- Each path can contain both markdown and source files
- Multiple paths can be specified to process different parts of a repository

### OPTIONS

- `-v`:
  - Enable verbose output showing detailed processing information.
- `-e`, `--extensions`:
  - Comma-separated list of source file extensions to process (e.g., ".go,.ts,.js").
  - When omitted, defaults to:
    ```text
    .go,.js,.ts,.jsx,.tsx,.java,.cs,.cpp,.c,.h,.hpp,.py,.rb,.php,.rs,.kt,.scala,.m,.swift,.fs,.md,.sql,.vsql
    ```
  - Extensions must include the dot prefix.  
- `-n`, `--dry-run`:
  - Perform a dry run without modifying files.

### ARGUMENTS

- `<paths>`:
  - One or more paths to process. Each path can contain both markdown requirement files and source code with coverage tags
  - At least one path must be provided
  - When multiple paths are provided, they are processed in sequence

### OUTPUT FILES

- Markdown files:
  - Updated with:
  - Coverage annotations for requirement sites
  - Coverage footnotes linking requirements to implementations

- Error handling
  - Files may be left in inconsistent state if error occurs, e.g.:
    - Partially updated footnotes
    - Missing coverage annotations
  - No rollback mechanism is provided

### EXIT STATUS

- 0: Success
- 1: Syntax/Semantic errors found during scan phase or other errors have occurred

### EXAMPLES

Process a single directory containing both markdown and source files:

```bash
reqmd trace project/
```

Process multiple directories with mixed content:

```bash
reqmd trace docs/ src/ tests/
```

Process only Go and TypeScript files in multiple directories:

```bash
reqmd trace -e .go,.ts docs/ src/ tests/
```

Process with verbose output:

```bash
reqmd trace -v docs/ src/ tests/
```
