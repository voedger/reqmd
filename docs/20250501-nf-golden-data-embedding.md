# Golden data embedding to avoid separate .md_ files

## Problem

Currently, our system test architecture relies on separate golden files (with `_` suffix) to store expected outputs. This requires maintaining duplicate files and complicates the testing workflow.

## Proposed solution

`~nf/GoldenDataEmbedding~`: Embed golden data directly in the source markdown files using specially formatted comments, similar to how we already handle error expectations.

Example:
```markdown
`~REQ001~`
> replace `~REQ001~`covered✅

Some other content
> replace Some other expected content
```

## Addessed issues

- https://github.com/voedger/reqmd/issues/19
