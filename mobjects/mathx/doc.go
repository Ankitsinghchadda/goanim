// Package mathx provides math-rendering mobjects: LaTeX equations,
// number lines, 2D coordinate axes, and plotted functions.
//
// Equations are rendered via a pure-Go LaTeX parser
// (tdewolff/canvas.ParseLaTeX, backed by star-tex.org/x/tex). No
// external LaTeX installation is required. The parser supports a
// useful subset of math expressions; complex documents and custom
// packages are out of scope.
//
// When rendered in sketchy mode (Sloppiness > Architect), equations
// pick up the active style's roughness — producing a "handwritten
// math" effect that's unique to this library. The math stays correct
// (path topology preserved); only the stroke style changes.
package mathx
