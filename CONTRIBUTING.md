# Contributing

## Pull request titles

Pull requests are squash-merged, so the PR title becomes the commit message on `main` and decides
whether a release is published. Titles follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): summary
```

The scope is optional, and a CI check rejects titles that do not match.

| Type | Use it for | Release |
| --- | --- | --- |
| `feat` | A new capability | minor |
| `fix`, `perf`, `revert` | A bug fix, speedup, or revert | patch |
| `docs`, `refactor`, `test`, `build`, `ci`, `style`, `chore` | Anything else | none |

Add `!` after the type or scope for a breaking change and describe what breaks in the PR body. It
releases a new major version.

## Releases

Merging a `feat`, `fix`, `perf`, `revert`, or breaking change to `main` tags `vX.Y.Z` and publishes a
GitHub release with archives, `checksums.txt`, build provenance, and Homebrew formulas. The version is
chosen from the commit messages since the last tag, and the first release is 0.1.0.

## Local checks

```
gofmt -l .
go vet ./...
go test ./...
GOOS=windows go vet ./...
```

## Profilers

What a profiler must do is in [CONTRACT.md](CONTRACT.md). `CoreSchema` in `profile.go` is the core
schema version this release targets; change it together with the profilers' supported versions.

## Maintainer setup

- Merge with "Squash and merge", using the PR title as the commit message.
- Require the `title` and `verify` checks on `main`.
- Copy `rc.rb` and `rc-profiler-go.rb` from each release into `runtimeconditions/homebrew-tools`.
