# Profiler contract

`rc profile <language> [args...]` runs a language profiler as a separate executable, called a plugin. The
Go profiler is built into `rc` and is not a plugin. This is what a plugin must do so `rc` can run it safely. Everything else about its arguments belongs
to the profiler, and `rc` passes them through unchanged.

## Finding the profiler

`RC_<LANGUAGE>_PROFILER_BIN`, then `rc-profiler-<language>` on `PATH`. The published
executables `runtimeconditions-python-profiler` and `runtimeconditions-typescript-profiler` are also found.

## Runtime and manifest

A plugin runs only if it is installed and its language runtime command is on `PATH`; `rc` checks both before
it runs the plugin, with no version check, and reports a missing plugin, a missing runtime, or both in one error.
`rc` learns the runtime from a manifest at `<prefix>/share/rc/plugins/<language>.json` for a plugin installed at
`<prefix>/bin/rc-profiler-<language>`, else from its built-in table for `typescript` (`node`), `python`
(`python3`) and `java` (`java`):

```json
{"language": "typescript", "runtime": {"command": "node", "name": "Node.js", "install": "brew install node"}}
```

`language` and `runtime.command` are required; other fields are ignored. A malformed manifest is ignored with
a warning on standard error.

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

The Go profiler is built into `rc`; its `--version --json` answers `rc doctor`.

When a profiler does not answer `--version --json`, `rc` prints a note to standard error and
runs it without the compatibility check.
