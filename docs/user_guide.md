# reqmd user guide

## Introduction

reqmd is a command-line tool that traces requirements from Markdown files to their corresponding coverage in source files. It establishes traceability links between requirement identifiers and coverage tags, automatically generating footnotes that link requirement identifiers to coverage tags.

## Installation

To install reqmd:

```bash
go install github.com/voedger/reqmd@latest
```

## Basic usage

The basic command syntax is:

```bash
reqmd [-v] trace [ (-e | --extensions) <extensions>] [--dry-run | -n] <paths>...
```

### Command options

- `-v`: Enable verbose output showing detailed processing information
- `-e`, `--extensions`: Comma-separated list of source file extensions to process (e.g., ".go,.ts,.js")
- `-n`, `--dry-run`: Perform a dry run without modifying files

### Path arguments

The tool accepts one or more paths to process. Each path can contain both markdown requirement files and source code with coverage tags. When multiple paths are provided, they are processed in sequence.

## Examples

### Process a single directory

To process a directory containing both markdown and source files:

```bash
reqmd trace project/
```

### Process multiple directories

To process multiple directories with mixed content:

```bash
reqmd trace docs/ src/ tests/
```

### Process specific file types

To process only Go and TypeScript files in multiple directories:

```bash
reqmd trace -e .go,.ts docs/ src/ tests/
```

### Verbose output

To see detailed processing information:

```bash
reqmd trace -v docs/ src/ tests/
```

### Dry run

To preview changes without modifying files:

```bash
reqmd trace -n docs/ src/ tests/
```

## Output files

### reqmdfiles.json

This file is created or updated in each directory containing markdown files when FileURLs are present. It maps FileURLs to their git hashes.

### Markdown files

Markdown files are updated with:
- Coverage annotations for requirement sites
- Coverage footnotes linking requirements to implementations

## Error handling

Files may be left in an inconsistent state if an error occurs, for example:
- Partially updated footnotes
- Missing coverage annotations

No rollback mechanism is provided.

## Exit status

- 0: Success
- 1: Syntax/Semantic errors found during scan phase or other errors have occurred 