# rc

The Runtime Conditions command line, built with [Cobra](https://github.com/spf13/cobra).

```sh
rc profile <language> [args...]
rc bindings <command> [args...]
rc doctor
rc --version
```

`rc profile` runs an installed language profiler, passes the remaining arguments to it
unchanged, and exits with the profiler's status. It looks for the executable named by
`RC_<LANGUAGE>_PROFILER_BIN`, then `rc-profiler-<language>` on `PATH`. The `go`, `python`
and `typescript` profilers are also found as `go-rc-profiler`,
`runtimeconditions-python-profiler` and `runtimeconditions-typescript-profiler`.

```sh
npm install --save-dev @runtimeconditions/typescript-profiler
rc profile typescript generate --project . --name orders --workload-uri https://services.example.com/orders
```

A profiler for a new language needs no change here: put an executable named
`rc-profiler-<language>` on `PATH`.

`rc` checks that the profiler supports the core schema it targets before running it. The
requirements on a profiler are in [CONTRACT.md](CONTRACT.md).

## Install

Every merge to `main` with a releasable change publishes a GitHub release with archives for
macOS, Linux and Windows, `checksums.txt`, build provenance attestations
(`gh attestation verify <archive> --repo runtimeconditions/go-rc-cli`), and a generated Homebrew
formulas, `rc.rb` and `rc-profiler-go.rb` (built from the latest Go profiler release). Titles follow [Conventional Commits](https://www.conventionalcommits.org/):
`feat` releases a minor version, `fix`, `perf` and `revert` a patch, and `!` a major.

### Homebrew tap (maintainers)

Create `runtimeconditions/homebrew-tools` and add a repository secret `HOMEBREW_TAP_TOKEN` here: a
fine-grained token with contents write access to that repository only. Each release then commits
`Formula/rc.rb` and `Formula/rc-profiler-go.rb` to it. Without the secret the step is skipped, and the
formulas stay attached to the release for copying by hand.

## Build

```sh
go build -ldflags "-X main.version=0.1.0" -o rc ./cmd/rc
go test ./...
```

`rc bindings` works the same way: it runs `rc-bindings` (or `RC_BINDINGS_BIN`) and passes every
argument through. That tool is built from
[`runtimeconditions/extensions`](https://github.com/runtimeconditions/extensions), whose Go modules
are not published, so it stays a separate executable instead of being linked into `rc`.

`rc doctor` lists the bindings tool and every profiler it can find with its version and supported core
schemas, and exits non-zero when a profiler does not support the core schema `rc` targets.

The Go profiler is not linked into `rc` either. It is a separate program, released with its core schema
built in, and a Go `main` package cannot be imported. The Homebrew formula for `rc` depends on a formula
for it, so `brew install rc` installs both, and `rc profile go` finds `go-rc-profiler`.

MIT licensed.
