# Implement golden data embedding to avoid separate .md_ files

## Problem

Currently, our system test architecture relies on separate golden files (with `_` suffix) to store expected outputs. This requires maintaining duplicate files and complicates the testing workflow.

## Proposed solution

`~nf/GoldenDataEmbedding~`: Embed golden data directly in the source markdown files using specially formatted comments, similar to how we already handle error expectations.

Example:
```markdown
`~REQ001~`
// golden: `~REQ001~`covered✅

Some other content
// golden: Some other expected content
```

Complete syntax

- `// golden-`: Remove the previous non-golden line
- `// golden+`: Add a line after the previous non-golden line
  - Multiple statements are allowed and processed in order
- `// golden`: Replace the previous non-golden line
- `// golden1`: Insert a line at the beginning of the file
  - Multiple statements are allowed and processed in order
- `// golden>>`: Append a line at the end of the file
  - Multiple statements are allowed and processed in order

## Addessed issues

- https://github.com/voedger/reqmd/issues/19