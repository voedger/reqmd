# reqmd

[![Go Report Card](https://goreportcard.com/badge/github.com/voedger/reqmd)](https://goreportcard.com/report/github.com/voedger/reqmd)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Build Status](https://github.com/voedger/reqmd/actions/workflows/build.yml/badge.svg)](https://github.com/voedger/reqmd/actions)

**reqmd** is a command-line tool that traces requirements from Markdown documents to their coverage in source code. It automatically generates traceability links, ensuring seamless tracking between specifications and implementation.

## Features

- Extracts requirement references from Markdown files
- Scans source files for coverage tags
- Generates and updates coverage footnotes in Markdown
- Maintains a reqmdfiles.json file for tracking file hashes
- Fast & scalable – uses Go concurrency to process files efficiently

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

### Tracing Requirements

Scan Markdown files and source repositories to generate coverage mapping:

```sh
reqmd trace <path-to-markdowns> [<path-to-cloned-repo>...]
```

#### Arguments

- `<path-to-markdowns>` (Required) – Directory containing Markdown files
- `<path-to-cloned-repo>` (Optional) – Directory containing cloned source repositories

### Examples

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

## Design

Refer to [design.md](design.md) for detailed design decisions and architecture.

## Requirements & Dependencies

- Go 1.18+

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

This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Inspired by the need for better traceability between requirements and implementation in modern software projects.
