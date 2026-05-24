// Package style is goanim's style system: a set of independent
// attributes (Sloppiness, Edges, FillStyle, StrokeStyle, colors, font,
// opacity) that every mobject carries, resolved through a three-layer
// inheritance chain (library defaults → scene default → per-mobject
// override) at render time.
//
// The design mirrors Excalidraw's UI: users pick orthogonal options
// from a panel rather than choosing a "mode." That makes combinations
// possible that Excalidraw itself doesn't offer — crisp geometric
// shapes with hatch fills (blueprint), or hand-drawn shapes with solid
// fills (cartoon).
//
// Inheritance is encoded with sentinel "Unset" zero values on every
// enum, so `Style{}` denotes "inherit everything" and partial-literal
// construction works naturally:
//
//	style.Style{Sloppiness: style.SloppinessCartoonist, FillStyle: style.FillSolid}
//
// At resolution time the chain is walked from mobject to scene to
// library, taking the first non-Unset value at each step.
package style
