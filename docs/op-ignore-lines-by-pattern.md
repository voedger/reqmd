# Ignore lines by pattern

The `trace` command must support excluding specific lines from processing using the `--ignore-lines` option with line patterns.

## Motivation

Ignore lines in system tests with the `// line:` prefix:

```markdown
`~func1~`uncvrd[^1]❓
`~func2~`
// line: `~func2~`uncvrd[^2]❓
```

## Solution

`~nf/IgnoreLinesByPattern~`: reqmd trace (--ignore-lines <lines-pattern>)...

- `lines-pattern` is a RE2-expression that matches lines to be ignored. With more than one --ignore-lines, lines that match any of the patterns are ignored.

## Prompts

- Suggest an implementation plan to implement # Ignore lines by pattern. Do not generate any code yet. Update requirement and design document first

## Implementation plan

- Update documentation to clearly describe the `--ignore-lines` option
- Add `ignoreLines` field to scanner configuration
- Modify the file parsing logic to check each line against ignore patterns
- Update the CLI command parser to accept multiple `--ignore-lines` flags
- Add unit tests for line pattern matching
- Add system tests with golden data to verify ignored lines
- Update help text and examples in README.md
