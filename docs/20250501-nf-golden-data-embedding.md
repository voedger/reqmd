# Golden data embedding to avoid separate .md_ files

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

More examples:

- `> delete`: Remove the previous non-golden line
- `> insert: text`: Add a line after the previous non-golden line
  - Multiple statements are allowed and processed in order
- `> replace: text`: Replace the previous non-golden line
- `> firstline: text`: Insert a line at the beginning of the file
  - Multiple statements are allowed and processed in order
- `> append: text`: Append a line at the end of the file
  - Multiple statements are allowed and processed in order
- `> deletelast`: Remove the last line of the file
  - Multiple statements are allowed and processed in order

## Addessed issues

- https://github.com/voedger/reqmd/issues/19
