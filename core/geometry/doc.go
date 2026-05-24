// Package geometry provides primitive vector types and operations used
// throughout goanim: points, paths (sequences of move/line/cubic commands),
// affine transforms, and axis-aligned bounding rectangles.
//
// The geometry layer is pure data and math — no rendering, no randomness,
// no global state. Coordinates use a center-origin convention where (0,0)
// is the middle of the canvas and Y points up; the renderer flips Y at
// the boundary to match image-space conventions.
package geometry
