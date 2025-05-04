# reqmd documentation

This directory contains detailed technical documentation for the reqmd tool, covering requirements, design, syntax specifications.

## Syntax and structure of input files

Ref. [ebnf.md](ebnf.md)

## Use cases

- [Installation](uc-installation.md)
- [Tracing](uc-tracing.md)

## Options

- [Ignore `.*` folders](op-ignore-dot-folders.md)
- [Ignore paths by pattern](op-ignore-paths-by-pattern.md)
- [Ignore lines by pattern](op-ignore-lines-by-pattern.md)
- [Force requirement types](op-force-requirement-types.md)

## Syntax/semantic errors

- [Handle inconsistency between Footnote and PackageId](err-inconsistency-between-footnote-and-packageid.md)

See also:

- [internal/errors_syn.go](../internal/errors_syn.go)
- [internal/errors_sem.go](../internal/errors_sem.go)  

## Test requirements

- [System tests](test-systests.md)
  - [Golden data embedding to avoid separate .md_ files](test-nf-golden-data-embedding.md)
- [High-volume file processing test](../tasks/T0022.md)  

## Other

- [Architecture](architecture.md) - Architecture and implementation design
- [Construction requirements](construction.md)
- [Processing phases](processing-phases.md)
- [Decisions](decisions.md)

## Research

- [Research Notes](rsch.md) - Research and study notes
