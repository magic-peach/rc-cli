# rc

The Runtime Conditions command line, built with [Cobra](https://github.com/spf13/cobra).

```sh
rc profile <language> [args...]
rc bindings <command> [args...]
rc doctor
rc --version
```

`rc bindings` and `rc profile go` run inside `rc`: the bindings come from
[`runtimeconditions/extensions`](https://github.com/runtimeconditions/extensions) and the Go profiler from
[`go-rc-profiler`](https://github.com/runtimeconditions/go-rc-profiler), both linked as Go modules, so no other
Runtime Conditions executable is needed. `rc` does not check for a Go toolchain; if the Go profiler needs one and
cannot find it, its own error and exit status reach you unchanged. `RC_GO_PROFILER_BIN` runs an external Go
profiler instead.

Every other language is a plugin: an executable named `rc-profiler-<language>`. `rc profile <language> [args...]`
passes the arguments to it unchanged and exits with its status. `rc` looks for the executable named by
`RC_<LANGUAGE>_PROFILER_BIN`, then `rc-profiler-<language>` on `PATH`; `python` and `typescript` are also found as
`runtimeconditions-python-profiler` and `runtimeconditions-typescript-profiler`.

```sh
brew install runtimeconditions/tools/rc-profiler-typescript
rc profile typescript generate --project . --name orders --workload-uri https://services.example.com/orders
```

A plugin runs only when it is installed and its language runtime (for example `node`) is on `PATH`; there is no
version check. When either is missing, `rc` exits with an error that names what is missing and how to install it,
both in one message if both are missing, and does not run the plugin.

`rc` knows each plugin's runtime without running the plugin:

| Plugin | Runtime command | Install the plugin |
| --- | --- | --- |
| `typescript` | `node` | `brew install runtimeconditions/tools/rc-profiler-typescript` |
| `python` | `python3` | `pip install runtimeconditions-profiler` (provides `runtimeconditions-python-profiler`) |
| `java` | `java` | not released yet |

A plugin can declare its runtime in a manifest, which overrides this table. For a plugin found at
`<prefix>/bin/rc-profiler-<language>`, `rc` reads `<prefix>/share/rc/plugins/<language>.json`:

```json
{"language": "typescript", "runtime": {"command": "node", "name": "Node.js", "install": "brew install node"}}
```

`command` is required; `name` defaults to `command` and `install` is optional. A missing manifest falls back to the
table. A malformed one falls back with a warning on standard error. A language that is neither in the table nor in a
manifest is run without a runtime check.

Before running a plugin, `rc` checks that it supports the core schema `rc` targets. The requirements on a profiler are
in [CONTRACT.md](CONTRACT.md).

`rc doctor` lists the bindings and the Go profiler as built in, and every plugin it can find with its version and
supported core schemas. It exits non-zero when a profiler does not support the core schema `rc` targets.

## Install

Every merge to `main` with a releasable change publishes a GitHub release with archives for
macOS, Linux and Windows, `checksums.txt`, build provenance attestations
(`gh attestation verify <archive> --repo runtimeconditions/go-rc-cli`), and a generated Homebrew
formula, `rc.rb`. Titles follow [Conventional Commits](https://www.conventionalcommits.org/):
`feat` releases a minor version, `fix`, `perf` and `revert` a patch, and `!` a major.

### Homebrew tap (maintainers)

Each release runs the `update-formulas` workflow in `runtimeconditions/homebrew-tools`, which updates `rc`. Create a
GitHub App with Actions write access to `homebrew-tools`, install it on this repository, and set the variable
`RC_RELEASE_APP_ID` and the secret `RC_RELEASE_APP_KEY` here. Without them the step is skipped, and the tap catches up
on its next scheduled run.

## Build

```sh
go test ./...
go build -ldflags "-X main.version=0.1.0 -X github.com/runtimeconditions/go-rc-profiler/extensioncheck.installedCoreSchemaBase64=$(base64 < runtimeconditions.profile.schema.yaml | tr -d '\n')" -o rc ./cmd/rc
```

The Go profiler has the core profile schema linked in at build time and refuses to run without it. Download the
schema for `CoreSchema` from the [spec releases](https://github.com/runtimeconditions/spec/releases), and pass its
base64 as above. The release workflow does this after checking the schema's SHA-256. Without it, the Go profiler cannot
validate a profile and fails with `installed profiler has no verified core profile schema`.

To work against unreleased changes in `extensions` or `go-rc-profiler`, put clones of them next to this repository
and use a `go.work` file outside the repository.

MIT licensed.
