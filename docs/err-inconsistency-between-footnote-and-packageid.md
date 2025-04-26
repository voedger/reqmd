# Handle inconsistency between Footnote and PackageId

The tool must detect semantic inconsistency between the `reqmd.package` header and the package used in coverage footnotes.

## Example

```markdown
---
reqmd.package: pkgmismatch
---

`~req~`covrd[^1]✅

[^1]: `[~other.package/req~impl]` [file.go:10:impl](https://github.com/org/repo/blob/main/file.go#L10)
```

In this example, the requirement site uses package `pkgmismatch` (from the header), but the coverage footnote references `other.package`, which creates an inconsistency that must be detected and reported as an error.

## Addressed issues

- [Handle inconsistency between Footnote and PackageId](https://github.com/voedger/reqmd/issues/8)
