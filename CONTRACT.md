# Profiler contract

`rc profile <language> [args...]` runs a language profiler as a separate executable. This is
what a profiler must do so `rc` can run it safely. Everything else about its arguments belongs
to the profiler, and `rc` passes them through unchanged.

## Finding the profiler

`RC_<LANGUAGE>_PROFILER_BIN`, then `rc-profiler-<language>` on `PATH`. The published
executables `go-rc-profiler`, `runtimeconditions-python-profiler` and
`runtimeconditions-typescript-profiler` are also found.

## Required

**`--version --json`** prints one JSON object to standard output and exits 0:

```json
{"name": "@runtimeconditions/typescript-profiler", "version": "0.1.0", "coreSchemas": ["0.4.0"]}
```

`coreSchemas` lists every core profile schema version the profiler can validate against.
Before running a command, `rc` calls this and stops with a clear error when `coreSchemas`
does not include the version `rc` targets (`CoreSchema` in `profile.go`, currently `0.4.0`).
Version queries such as `--version` and `--help` are not checked.

**Failure writes nothing.** A failed run exits non-zero, writes no output file, and leaves an
existing output file untouched. Success exits 0.

## Recommended

- `generate` produces a profile, and `verify-bindings` checks the installed binding packages.
- `--json` makes a failure print one object to standard error and nothing else:
  `{"error": {"stage": "extract", "message": "...", "location": "src/app.ts:3:16"}}`.
  `location` is optional.

## Current profilers

| Profiler | `--version --json` | `--json` errors | `verify-bindings` | Core schemas |
| --- | --- | --- | --- | --- |
| TypeScript 0.1.0 | yes | yes | yes | 0.4.0 |
| Python 0.4.0 | no (no `--version`) | no | yes | 0.4.0 |
| Go 0.6.0 | no (plain text `--version`) | no | no | 0.4.0, and 0.2.0 input |

When a profiler does not answer `--version --json`, `rc` prints a note to standard error and
runs it without the compatibility check.
