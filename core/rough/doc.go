// Package rough is goanim's roughness engine: a deterministic
// reimplementation of the algorithms from rough.js
// (https://github.com/rough-stuff/rough), adapted for Go and for the
// temporal-stability requirements of an animation system.
//
// Determinism is the load-bearing property of this package. Given the
// same Options.Seed and the same geometric inputs, every function in
// this package produces byte-identical output across runs, OSes, and
// architectures. No global PRNG state is used; each call constructs a
// Lehmer (Park-Miller) random sequence from the seed and consumes
// draws in a fixed, documented order.
//
// The primitives — RoughLine, RoughRectangle, RoughEllipse,
// RoughPolygon, RoughPath — return geometry.Path values that can be
// fed directly to a render.Renderer. Fill primitives — Hatch,
// CrossHatch, Zigzag — return path data for the inside of a closed
// shape.
package rough
