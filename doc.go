// Package goanim is a Go library for producing animated videos of
// system-design and mathematical / educational content. The visual
// style ranges from the hand-drawn aesthetic of Excalidraw to clean
// AWS-architecture-diagram crispness, all controlled by an orthogonal
// style system that lets callers set the look once and stamp the same
// diagram in any aesthetic.
//
// goanim is conceptually adjacent to Manim, but with a different
// visual style, an explicit layout-helpers stance (compose, do not
// compute coordinates), and pure-Go LaTeX rendering — no external
// TeX installation required.
//
// # Package layout
//
//   - core/scene      — Scene, the timeline/camera/mobject root.
//   - core/mobject    — Mobject interface plus primitive shapes.
//   - core/animation  — Animation primitives and combinators.
//   - core/render     — Rasterizer backends and the MP4 encoder.
//   - core/layout     — HBox/VBox/Grid/Stack/Padding/AlignTo helpers.
//   - core/direction  — Camera, laser pointer, pulse, spotlight, annotations.
//   - core/style      — Five preset stylesheets (Excalidraw, Sketchy, Crisp, ...).
//   - core/rough      — Hand-drawn / sketchy stroke engine (temporally stable).
//   - core/geometry   — Pt, BBox, transforms.
//   - core/icon       — Attachable contract for arrow routing.
//   - mobjects/icons  — 39 system-design icons (compute, storage, network, ...).
//   - mobjects/mathx  — Equation, NumberLine, Axes, Graph (pure-Go LaTeX).
//   - mobjects/systemdesign — Higher-level system-design composites.
//   - mobjects/netgraph     — Network topology layouts.
//
// # Getting started
//
// See the examples/ directory for runnable programs that demonstrate
// scene construction, animation composition, and MP4 output. The
// smallest end-to-end example is examples/animated_diagram.
//
// # External tools
//
// MP4 encoding shells out to ffmpeg (libx264). LaTeX rendering is
// fully in-process via tdewolff/canvas + star-tex.org/x/tex, no host
// TeX install required.
package goanim
