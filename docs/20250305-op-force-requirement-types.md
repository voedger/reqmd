# Force requirement types

## Problem

Currently there is a lack of the requirement type system.

## Proposal

`~nf/ForceRequirementTypes~`: reqmd trace --types <type list>

Requirement type is specified by the first part of the RequirementName

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
