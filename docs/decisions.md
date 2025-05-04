# Decisions

- RequirementSiteStatus
  - `covrd/covered` denotes the covered status
  - `covrd` is used by the tool as CoverageStatusWord
  - `covered` is kept for backward compatibility
  - `uncvrd` denotes the uncovered status
  - Motivation: use short words with a high level of uniqueness for covered/uncovered status
- Separation of the `<path-to-markdowns>` and `<path-to-sources>`
  - Paths are separated to avoid modifications of sources
- SSH URLs (like git@github.com:org/repo.git) are not supported
- Commit references
  - `main/master` is used as the default reference for file URLs instead of commit hashes
  - Motivation:
    - Simplifies maintenance by eliminating the need to track file changes
    - Enables working in branches that will be squashed
    - Provides more readable and stable URLs in documentation
