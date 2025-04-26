# Ignore `.*` folders

Subfolders of the `paths` passed to the `trace` command  with names that start with a `.` are ignored.

Example:

```bash
reqmd trace voedger-internals voedger-internals/reqman/.work/repos/voedger
```

`.work` folder must be ignored.
