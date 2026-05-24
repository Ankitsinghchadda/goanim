# Contributing to goanim

Thanks for your interest in goanim. This document covers how to get
a working dev environment, how the codebase is organized, and what
to expect from the review process.

## Prerequisites

- Go 1.22 or later
- `ffmpeg` on PATH for MP4 output (PNG-only workflows do not need it)
- macOS, Linux, or Windows (CI tests Linux + macOS)

## Setup

```bash
git clone https://github.com/ankitsinghchadda/goanim
cd goanim
go mod download
go test ./...
```

## Repository layout

Public packages live under `core/` and `mobjects/`. Anything inside
`internal/` is not part of the public API and may change without
notice. `examples/` contains runnable demo programs; `cmd/bench`
holds the benchmark runner.

See the [README](README.md#repository-layout) for the full tree.

## Code style

- Run `gofmt -w .` (or `goimports`) before submitting.
- Run `go vet ./...` — CI runs it.
- New exported symbols need godoc comments. Keep them short: the
  WHAT can live in the type name; comments should add the WHY, the
  invariants, or anything surprising.
- Don't add inline comments that just restate the code. Prefer
  removing or renaming over commenting.
- The library has no per-package `init()` side effects; please keep
  it that way.

## Tests

- Run `go test ./...` before sending a PR.
- New behavior needs a test in the same package.
- Determinism is a load-bearing property of this library. Anything
  that introduces non-determinism (concurrent ordering, time-based
  randomness, etc.) needs an explicit seam or it will break the
  golden tests.

## Performance work

Performance changes need a benchmark.

```bash
# Before:
go run ./cmd/bench --runs 3 --output /tmp/before.json

# After your change:
go run ./cmd/bench --runs 3 --compare /tmp/before.json
```

Regressions >5% on existing scenes block the PR until investigated.
See [`docs/performance.md`](docs/performance.md) for the profiling
guide.

## Pull requests

- One logical change per PR. Splitting unrelated cleanups into
  separate PRs makes review and revert both faster.
- Update [`CHANGELOG.md`](CHANGELOG.md) under `[Unreleased]` for any
  user-visible change.
- Link the issue you're closing in the PR description (`Fixes #N`).

## Reporting bugs

Open an issue with:

- A minimal reproduction (a single `main.go` is ideal).
- The expected output and the actual output.
- Your Go version (`go version`) and OS.
- The relevant frame / PNG if the issue is visual.

## Asking questions

Open a Discussion (preferred) or an issue tagged `question`. Please
search both before opening a duplicate.
