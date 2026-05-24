# Performance

This document covers goanim's performance characteristics, how to
benchmark, how to profile, and the architectural choices that govern
throughput.

## TL;DR

A small (5 s, 3-icon) sketchy scene at 1920×1080 renders in ~5.5 s
(approaching realtime). A medium 30 s × 10-icon scene renders in
~45 s.

The single biggest performance lever is **`Options.Supersample`**.
The default is `1`. Callers that need maximum stillframe quality
should pass `Supersample: 2` explicitly.

## Running benchmarks

The bench suite has two parts.

### Scene benchmarks — end-to-end

Build a scene, animate it, encode to MP4.

```bash
# Full suite (all scenes × both styles, 3 runs each):
go run ./cmd/bench

# Just the small scene, one run, comparing against the committed baseline:
go run ./cmd/bench --runs 1 --only small --compare bench_baseline.json

# Producing the JSON for a new baseline:
go run ./cmd/bench --runs 3 --output bench_baseline.json
```

Flags:

- `--output PATH` — JSON report destination (default `bench_report.json`).
- `--compare PATH` — JSON report to compare against. Prints a delta table.
- `--runs N` — runs per (scene, style); best-of-N reported. Default 3.
- `--only SUBSTR` — only run scenes whose `name-style` contains this.
- `--notes TEXT` — free-text note recorded in the report.

The scene catalog and durations are documented in
`cmd/bench/scenes/scenes.go`.

### Component benchmarks

Micro-benchmarks for isolated paths (rough geometry, layout, camera
transform, single-frame render). Standard `testing.B`:

```bash
# Full component suite:
go test ./internal/benchmarks/... -bench=. -benchmem -count=3

# Just the single-frame render benches, with CPU profile:
go test ./internal/benchmarks/... -run=NONE -bench=BenchmarkSingleFrameRender \
    -benchmem -benchtime=3s -cpuprofile=cpu.prof
```

## Profiling

Capture profiles:

```bash
# CPU profile (3 seconds of sampling):
go test ./internal/benchmarks/... -run=NONE \
    -bench=BenchmarkSingleFrameRender_Sketchy \
    -benchtime=3s -cpuprofile=/tmp/cpu.prof

# Memory profile (allocations):
go test ./internal/benchmarks/... -run=NONE \
    -bench=BenchmarkSingleFrameRender_Sketchy \
    -benchtime=3s -memprofile=/tmp/mem.prof
```

Read profile output:

```bash
# Top 20 functions by cumulative CPU time:
go tool pprof -top -cum /tmp/cpu.prof | head -25

# Interactive flame-graph in a browser:
go tool pprof -http=:8080 /tmp/cpu.prof

# Allocator sites by bytes allocated:
go tool pprof -alloc_space -top /tmp/mem.prof | head -25

# Allocator sites by allocation count:
go tool pprof -alloc_objects -top /tmp/mem.prof | head -25
```

Reading the `cum`/`flat` columns:

- **flat** is time spent INSIDE this function.
- **cum** is time spent in this function PLUS time in any function
  it called. The top of `pprof -top -cum` is usually framework /
  test machinery; the interesting hot paths are functions where
  `flat%` is high.

## Architecture for performance

The library's profile is dominated by per-frame work on the rendering
pipeline. The key optimizations baked into the defaults:

### Background composite

`CanvasRenderer.Image` fills the frame background via
`draw.Draw(dst, …, &image.Uniform{C: bg}, …, draw.Src)`, which uses
an internal fast-path on the RGBA pixel buffer — ~100× faster than
a per-pixel `dst.Set(x, y, bg)` double-loop.

### BiLinear downsample

The supersample → output downsample uses BiLinear (2-tap linear)
instead of CatmullRom (4-tap cubic). On hand-drawn content where the
supersample step did most of the anti-aliasing, BiLinear output is
visually equivalent at ~3× lower CPU.

### Supersample default

`Options.Supersample` defaults to `1`. The rasterizer already
produces anti-aliased output at DPMM=1; `Supersample: 2` contributes
~50% of per-frame CPU for a barely-perceptible quality gain. Opt
back into 2 explicitly if needed.

### Pause-frame caching

`direction.Pause` is a no-op animation. The scene player detects
`animation.ConstantAnim` (a marker interface implemented by Pause)
and renders ONCE per pause, writing the same image bytes N times to
the encoder. Works through `animation.Sequence` via the
`SequenceLike` interface — Pauses inside Sequences benefit too.

## Regression testing

The committed `bench_baseline.json` captures the current numbers. Run
with `--compare bench_baseline.json` to see delta vs baseline. Major
regressions show up as `+N%` highlighted in the comparison table;
improvements show as `-N%` bold.

Standard regression-test loop:

```bash
# Before any change: baseline number is in bench_baseline.json.
# After your change:
go run ./cmd/bench --runs 1 --only medium --compare bench_baseline.json
```

If a change regresses an existing benchmark by >5%, investigate
before merging. The comparison output is suitable for pasting into
PRs.

## Remaining bottlenecks

The dominant per-frame cost is rasterization itself (tdewolff/canvas's
path → raster conversion). Possible future wins:

- **Parallel frame rendering** — Highest-impact unshipped optimization.
  Per-frame work is independent; with proper state-snapshotting an
  M-core machine could render at M× speed. The challenge is that
  `animation.Apply` mutates mobject state, so the current single-thread
  architecture inherently serializes the pipeline.
- **Rasterizer choice** — Swapping `tdewolff/canvas`'s rasterizer for
  a SIMD/vectorized path (rasterx, `golang.org/x/image/vector` with
  explicit SIMD) could shave another ~20% off frame time.
- **Per-mobject opacity propagation through Arrow labels** — Arrow
  labels don't honor parent `Arrow.Opacity`. Not a perf issue per
  se, but related — fixing it would let `clearScene` be removed from
  the URL-shortener pattern.

None of these are blockers; the current numbers exceed targets for
small and medium scenes by comfortable margins.
