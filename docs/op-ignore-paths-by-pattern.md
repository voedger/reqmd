# Ignore paths by pattern

The `trace` command must support excluding specific files and folders from processing using the `--ignore` option with path patterns.

Potential use cases:

- Skip test data directories
- Exclude build artifacts
- Ignore source control folders

## Motivation

Attempt to trace reqmd by reqmd:

```txt
reqmd trace -v .

2025/04/26 21:25:09 Starting processing with multi-path approach 
	wd: C:\workspaces\work\reqmd, 
	paths: [.]
2025/04/26 21:25:09 Absolute path: C:\workspaces\work\reqmd
2025/04/26 21:25:09 Skipping folder 
	path: C:\workspaces\work\reqmd\.cursor
2025/04/26 21:25:09 Skipping folder 
	path: C:\workspaces\work\reqmd\.git
2025/04/26 21:25:09 Skipping folder 
	path: C:\workspaces\work\reqmd\.github
2025/04/26 21:25:09 Skipping large file 
	path: C:/workspaces/work/reqmd/reqmd.exe, 
	size: 19.0 MB
2025/04/26 21:25:09 Skipping folder 
	path: C:\workspaces\work\reqmd\internal\.testdata
2025/04/26 21:25:09 Skipping folder 
	path: C:\workspaces\work\reqmd\internal\experiments\.testdata
2025/04/26 21:25:09 Skipping folder 
	path: C:\workspaces\work\reqmd\internal\hvgen\.testdata
2025/04/26 21:25:09 Skipping folder 
	path: C:\workspaces\work\reqmd\internal\systrun\.testdata
2025/04/26 21:25:09 Skipping folder 
	path: C:\workspaces\work\reqmd\internal\testdata\systest\justreqs\req\.ignored
2025/04/26 21:25:09 Scan complete (multi-path) 
	processed files: 90, 
	processed size: 274.3 kB, 
	skipped files: 1, 
	skipped size: 19.0 MB, 
	duration: 15.9622ms
C:/workspaces/work/reqmd/tasks/worklog.md:628: only one RequirementSite is allowed per line: `~REQ001~`,  `~REQ001~`covered[^~REQ001~]âœ…
C:/workspaces/work/reqmd/internal/testdata/systest/errors_syn/err_reqident.md:7: RequirementName shall be an identifier
C:/workspaces/work/reqmd/internal/testdata/systest/errors_syn/err_unmatchedfence.md:9: opening code block fence at line 9 has no matching closing fence
C:/workspaces/work/reqmd/internal/testdata/systest/errors_syn/err_pkgident.md:2: PackageID shall be an identifier: 11com.example.basic
C:/workspaces/work/reqmd/internal/testdata/systest/errors_syn/err_urlsyntax.md:5: CoverageStatusWord shall be 'covered' or 'uncvrd': covrd
C:/workspaces/work/reqmd/internal/testdata/systest/errors_syn/err_urlsyntax.md:7: URL shall adhere to a valid syntax: ://github.com/voedger/example/blob/main/reqsrcfootnote.go#L5
C:/workspaces/work/reqmd/internal/testdata/systest/errors_syn/err_covstatus.md:7: CoverageStatusWord shall be 'covered' or 'uncvrd': cov
C:/workspaces/work/reqmd/internal/systrun/testdata/err_unexpected/req/err_pkgident.md:2: PackageID shall be an identifier: 11com.example.basic
C:/workspaces/work/reqmd/internal/testdata/systest/reqsrc/footnote.md:7: CoverageStatusWord shall be 'covered' or 'uncvrd': covrd
C:/workspaces/work/reqmd/internal/systrun/testdata/err_matchedunmatched/req/err_pkgident.md:2: PackageID shall be an identifier: 11com.example.basic
C:/workspaces/work/reqmd/internal/systrun/testdata/err_matchedunmatched/req/err_pkgident.md:8: RequirementName shall be an identifier
```
