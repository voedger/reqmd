# Force requirement types

## Problem

The current system lacks a structured requirement type system, making it difficult to categorize and filter requirements by their intended purpose.

## Proposal

- `~op.ForceRequirementTypes~`: `reqmd trace --types <type list>`
  - Example: `reqmd trace --types it,cmp,utest`
  - Order is important and will be used to track coverage and generate reports
  - If MarkdownFile contains requirements of types not mentioned in the `--types` option, a syntax error `reqtype` is reported: `Requirement type must be one of %v: %v`

The RequirementType of the RequirementSite is determined by the first segment of the RequirementName. Examples of requirement types:

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

## Technical design

### internal/typeregistry.go

```go
type RequirementType struct {
    Identifier string // The prefix that identifies this type (e.g., "it", "cmp", "utest")
    OrderNo   int     // For ordering in reports and coverage analysis
}

type TypeRegistry struct {
    Types map[string]RequirementType  // Map of type identifiers to their definitions
    Identifiers []string              // Ordered list of type identifiers
}
```

- `NewTypeRegistry(typeDefs []RequirementType) *TypeRegistry`
  - Creates a new registry with the specified requirement types
  - Validates that no duplicate identifiers exist

- `(r *TypeRegistry) Type(identifier string) (RequirementType, bool)`
  - Retrieves a requirement type by its identifier
  - Returns the type and whether it was found

- `ExtractTypeFromRequirement(requirementName string) string`
  - Extracts the type identifier from a requirement name (first segment before the dot)
  - Example: "it.TestQueryProcessor2_CDocs" → "it"

- `ParseTypeList(typeList string) ([]string, error)`
  - Parses a comma-separated list of type identifiers
  - Validates that all identifiers are unique and match the Name rule
  - Returns an ordered list of type identifiers  

### Scanner

- ScannerConfig that includes *TypeRegistry and existing ignorePatterns
- ScannerContext that includes *TypeRegistry and existing ignorePatterns
- parseRequirements
  - Takes `sctx *ScannerContext` as a param
  - Validate the RequirementType part of the RequirementNames

### internal/cmd.go

- newTraceCmd.RunE
  - scfg ScannerConfig

### System tests  

- Test for the `reqtype` error

## Addressed issues

- https://github.com/voedger/reqmd/issues/21
