# Force requirement types

## Problem

The current system lacks a structured requirement type system, making it difficult to categorize and filter requirements by their intended purpose.

## Proposal

`~nf/ForceRequirementTypes~`: reqmd trace --types <type list>

Requirement type is determined by the first segment of the RequirementName. Examples of requirement types:

- `~it.TestQueryProcessor2_CDocs~`: Integration Test
- `~cmp.cdocsHandler~`: Component
- `~utest.isequencer.mockISeqStorage~`: Unit Test

## Background

`RequirementSite` is defined as follows:

```ebnf
  Identifier     = Name { "." Name } .
  RequirementSite = RequirementSiteLabel [ CoverageStatusWord CoverageFootnoteReference ] [ CoverageStatusEmoji ] .
  RequirementSiteLabel = "`" RequirementSiteId "`" .
  RequirementSiteId = "~" RequirementName "~" .
  CoverageStatusWord = "covered" | "uncvrd" | "covrd" .
  RequirementName = Identifier .
  RequirementId = PackageId "/" RequirementName .
  CoverageFootnoteReference = "[^" CoverageFootnoteId "]" .
  CoverageStatusEmoji = "✅" | "❓" .
```

## Addressed issues

- https://github.com/voedger/reqmd/issues/21
