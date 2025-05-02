# reqmd

[![Go Report Card](https://goreportcard.com/badge/github.com/voedger/reqmd)](https://goreportcard.com/report/github.com/voedger/reqmd)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.txt)

**reqmd** is a command-line tool that traces requirements from Markdown documents to their coverage in source code. It automatically generates traceability links, ensuring seamless tracking between specifications and implementation.

## Features

- Extracts requirement references from Markdown files
- Scans source files for coverage tags
- Generates and updates coverage footnotes in Markdown
- Uses branch references (main/master) for stable file URLs
- Fast & scalable – uses Go concurrency to process files efficiently
- Supports multiple paths with mixed markdown and source files

## Installation

Install `reqmd` via Go:

```sh
go install github.com/voedger/reqmd@latest
```

or build from source:

```sh
git clone https://github.com/voedger/reqmd.git
cd reqmd
go build -o reqmd .
```

## Usage

### Tracing requirements

Scan directories containing both Markdown files and source code to generate coverage mapping:

```sh
reqmd [-v] trace [ (-e | --extensions) <extensions>] [--dry-run | -n] <paths>...
```

#### Options

- `-v`: Enable verbose output showing detailed processing information
- `-e`, `--extensions`: Comma-separated list of source file extensions to process (e.g., ".go,.ts,.js")
- `-n`, `--dry-run`: Perform a dry run without modifying files

#### Arguments

- `<paths>`: One or more paths to process. Each path can contain both markdown requirement files and source code with coverage tags
  - At least one path must be provided
  - When multiple paths are provided, they are processed in sequence

### Examples

Process a single directory containing both markdown and source files:

```sh
reqmd trace project/
```

Process multiple directories with mixed content:

```sh
reqmd trace docs/ src/ tests/
```

Process only Go and TypeScript files in multiple directories:

```sh
reqmd trace -e .go,.ts docs/ src/ tests/
```

Process with verbose output:

```sh
reqmd trace -v docs/ src/ tests/
```

### Example files

`requirements.md`

```markdown
- APIv2 implementation shall provide a handler for POST requests. `~Post.handler~`coverage[^~Post.handler~].
```

`handler.go`

```go
// [~server.api.v2/Post.handler~impl]
func handlePostRequest(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

Generated coverage footnote for `requirements.md`:

```markdown
[^~Post.handler~]: `[~server.api.v2~impl]`[pkg/http/handler.go:42:impl](https://github.com/repo/pkg/http/handler.go#L42)
```

## Output files

### reqmdfiles.json

This file is created or updated in each directory containing markdown files when FileURLs are present. It maps FileURLs to their git hashes.

### Markdown files

Markdown files are updated with:

- Coverage annotations for requirement sites
- Coverage footnotes linking requirements to implementations

## Design

Refer to [design.md](docs/architecture.md) for detailed design decisions and architecture.

## Requirements & dependencies

- Go 1.23+

## Contributing

We welcome contributions. To get started:

1. Fork the repository
2. Clone your fork locally
3. Create a branch for your changes
4. Commit and push
5. Open a Pull Request

```sh
git clone https://github.com/yourusername/reqmd.git
cd reqmd
git checkout -b feature-new-enhancement
```

## License

reqmd is released under the Apache 2.0 license. See [LICENSE.txt](LICENSE.txt)

## Acknowledgments

Notation is inspired by [OpenFastTrace](https://github.com/itsallcode/openfasttrace).
