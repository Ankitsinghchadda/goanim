# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

While the major version is `0`, the public API may change between
minor versions. The library will commit to backwards-compatible
minor releases starting at `v1.0.0`.

## [Unreleased]

## [0.1.0] — 2026-05-24

Initial public release.

### Added — Library core (`core/`)

- `core/scene` — Scene container, frame stepper, player.
- `core/mobject` — Mobject interface, primitive shapes, Group, Text,
  attachable contract, rough-geometry cache.
- `core/animation` — Animation interface, primitives (FadeIn/Out,
  DrawOn, MoveTo, MoveAlong, PopIn, Flash, Wait, SetStyle,
  MorphStyle), combinators (Sequence, Parallel, Stagger), easing
  (incl. Spring).
- `core/render` — `CanvasRenderer` (tdewolff/canvas backend), video
  encoder piping RGB frames to ffmpeg / libx264 at crf=18. Bundled
  Excalifont and Inter fonts.
- `core/layout` — `HBox`, `VBox`, `Grid`, `Stack`, `Padding`,
  `AlignTo`. Lazy bounds-based, recursive.
- `core/direction` — Pause, Camera (Pan/Zoom/Focus/Reset/FocusAt),
  LaserPointer, Pulse, Spotlight, UnderlineOn / CircleAround /
  Callout / Caption / LabelNear / Replay. Scene patterns:
  TitleSlide, ChapterIntro, EstablishingShot, DetailFocus.
- `core/style` — Five style presets (Excalidraw, Sketchy, Crisp,
  Blueprint, Notebook), Style override system with attribute-level
  composition, Tokens, Context.
- `core/rough` — rough.js-style hand-drawn stroke engine with stable
  geometry caching.
- `core/geometry` — Point, Path, Rect, Transform.
- `core/icon` — IconBase + Attachable contract for arrow routing.

### Added — Higher-level mobjects (`mobjects/`)

- `mobjects/icons` — 39 system-design icons across compute, storage,
  network, messaging, observability, dataflow, and endpoint
  categories. Every icon ships in all three sloppiness levels.
- `mobjects/mathx` — Equation, NumberLine, Axes, Graph with pure-Go
  LaTeX rendering via `tdewolff/canvas.ParseLaTeX` (no host TeX
  installation required). Disk-cached for fast iteration.
- `mobjects/systemdesign` — Arrow (with smart routing — straight /
  orthogonal / curved), Packet, composite shapes.
- `mobjects/netgraph` — Network topology layouts.
- `mobjects/mascot` — Mascot character primitives.
- `mobjects/gridbg`, `mobjects/packet`, `mobjects/nqueens` —
  helper mobjects used by the longer-form examples.

### Added — Tooling and examples

- `cmd/bench` — End-to-end and component benchmark suite with
  comparison-vs-baseline reporting.
- `examples/` — Runnable demo programs covering single PNGs, MP4
  diagrams, multi-style choreography, the direction-layer
  walkthroughs, math reveals, and the long-form architecture
  explainers (`bgmi_*`, `gmail_synced`, `nqueens_synced`,
  `tenx_engineer`).

### Performance

- Per-frame allocations reduced ~99% from initial implementation
  (background composite via `image.Uniform` + `draw.Src`).
- Downsample filter switched to BiLinear for ~3× lower CPU vs
  CatmullRom on hand-drawn content.
- `Options.Supersample` defaults to `1` (was `2`); explicit opt-in
  for print-quality stills.
- Pause-frame caching: `direction.Pause` and nested Sequences
  rasterize once per pause and reuse the frame N times.

### Notes

- LaTeX support targets inline math; complex documents,
  `\begin{align}` blocks, and custom packages are out of scope.
- `icon.LoadSVG` is stubbed; `icon.LoadPNG` is the supported path
  for branded logos.
- The `internal/` packages are off-limits to consumers per Go's
  import rules.

[Unreleased]: https://github.com/ankitsinghchadda/goanim/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ankitsinghchadda/goanim/releases/tag/v0.1.0
