---
reqmd.package: correct.package
---

# Package ID Mismatch

- `~REQ001~`uncvrd[^1]❓

[^1]: `[~wrong.package/REQ001~impl]`[file.go:10:impl](https://example.com/file.go#L10)
// errors: "PackageID in CoverageFootnoteHint.*wrong.package.*does not match PackageID in PackageDeclaration.*correct.package" 