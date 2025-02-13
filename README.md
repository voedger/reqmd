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

#### Arguments:

- `<path-to-markdowns>` (Required) – Directory containing Markdown files
- `<path-to-cloned-repo>` (Optional) – Directory containing cloned source repositories

### Example Workflow

#### Markdown File (`requirements.md`)

```markdown
- APIv2 implementation shall provide a handler for POST requests. `~Post.handler~`coverage[^~Post.handler~].
```

#### Source Code (`handler.go`)

```go
// [~server.api.v2/Post.handler~impl]
func handlePostRequest(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

#### Generated Coverage Footnote

```markdown
[^~Post.handler~]: `[~server.api.v2~impl]`[pkg/handler.go:42:impl](https://github.com/repo/pkg/handler.go#L42)
```

## How It Works

The tool follows a three-stage pipeline architecture:

1. **Scan** – Build the Domain Model
   - Recursively discovers Markdown and source files
   - Extracts requirement references from Markdown files
   - Identifies coverage tags in source code
   - Builds file structures with Git metadata
   - Collects any syntax errors during parsing

2. **Analyze** – Validate and Plan Changes
   - Performs semantic validation checks
   - Ensures requirement IDs are unique
   - Determines which coverage footnotes need updates
   - Identifies requirements needing coverage annotations
   - Verifies file hashes against reqmdfiles.json
   - Generates a list of required file modifications

3. **Apply** – Update Files
   - Updates or creates coverage footnotes
   - Appends coverage annotations to requirements
   - Maintains reqmdfiles.json for file tracking
   - Ensures changes are made only when no errors exist

The system is designed using SOLID principles:
- Each component has a single, focused responsibility
- New features can be added without modifying existing code
- Components interact through well-defined interfaces
- Dependencies are injected for flexible configuration

## Design

Refer to [design.md](design.md) for detailed design decisions and architecture.

## Requirements & Dependencies

- Go 1.18+
- Git (for hashing file contents using `git hash-object`)

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
