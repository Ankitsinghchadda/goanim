package style

import "image/color"

// Hex colors used in the preset palette. Kept here so presets are
// self-contained — internal/color is still the canonical palette but
// the style layer needs its own minimal subset to avoid an import
// cycle into internal/* from a core/* package.
var (
	colorInk         = color.RGBA{0x1E, 0x1E, 0x1E, 0xFF}
	colorBlueFill    = color.RGBA{0xA5, 0xD8, 0xFF, 0xFF}
	colorBlueprintFG = color.RGBA{0xE3, 0xF2, 0xFD, 0xFF} // pale-blue lines on the dark blue paper
	colorSlate900    = color.RGBA{0x0F, 0x17, 0x2A, 0xFF}
	colorSlate200    = color.RGBA{0xE2, 0xE8, 0xF0, 0xFF}
	colorSlate100    = color.RGBA{0xF1, 0xF5, 0xF9, 0xFF}
	colorCreamFill   = color.RGBA{0xFF, 0xF8, 0xE1, 0xFF}
)

// PresetExcalidraw — the out-of-the-box default. Hand-drawn, hatched,
// pale fills, slightly wobbly. Closest to what Excalidraw itself
// produces when you drop in a shape.
var PresetExcalidraw = Style{
	Sloppiness:  SloppinessArtist,
	Edges:       EdgesRound,
	StrokeWidth: StrokeNormal,
	StrokeStyle: StrokeSolid,
	StrokeColor: colorInk,
	FillStyle:   FillHatch,
	FillColor:   colorBlueFill,
	FontFamily:  FontHandDrawn,
	FontSize:    FontMedium,
	Opacity:     PFloat(1),
}

// PresetSketchy — maximum hand-drawn feel. Heavy wobble, cross-hatch
// fills, slightly thicker strokes.
var PresetSketchy = Style{
	Sloppiness:  SloppinessCartoonist,
	Edges:       EdgesRound,
	StrokeWidth: StrokeNormal,
	StrokeStyle: StrokeSolid,
	StrokeColor: colorInk,
	FillStyle:   FillCrossHatch,
	FillColor:   colorCreamFill,
	FontFamily:  FontHandDrawn,
	FontSize:    FontMedium,
	Opacity:     PFloat(1),
}

// PresetCrisp — clean geometric shapes, no wobble, solid fills, sans
// font. The opposite end of the spectrum from Sketchy.
var PresetCrisp = Style{
	Sloppiness:  SloppinessArchitect,
	Edges:       EdgesSharp,
	StrokeWidth: StrokeNormal,
	StrokeStyle: StrokeSolid,
	StrokeColor: colorSlate900,
	FillStyle:   FillSolid,
	FillColor:   colorSlate100,
	FontFamily:  FontSans,
	FontSize:    FontMedium,
	Opacity:     PFloat(1),
}

// PresetBlueprint — crisp geometric shapes with hatch fill on a deep
// blue paper. Demonstrates style attributes composing independently
// (Architect sloppiness × Hatch fill on a dark background — a
// combination Excalidraw itself doesn't expose).
//
// Pair with a background of #1A4E8C; shapes and text render in a pale
// blue so they read like blueprint linework.
var PresetBlueprint = Style{
	Sloppiness:  SloppinessArchitect,
	Edges:       EdgesSharp,
	StrokeWidth: StrokeNormal,
	StrokeStyle: StrokeSolid,
	StrokeColor: colorBlueprintFG,
	FillStyle:   FillHatch,
	FillColor:   nil, // transparent fill — hatch only against the paper
	FontFamily:  FontSans,
	FontSize:    FontMedium,
	Opacity:     PFloat(1),
}

// PresetNotebook — gentle wobble with a muted, paper-like palette.
// Like jotting a quick diagram in a notebook.
var PresetNotebook = Style{
	Sloppiness:  SloppinessArtist,
	Edges:       EdgesRound,
	StrokeWidth: StrokeThin,
	StrokeStyle: StrokeSolid,
	StrokeColor: colorSlate900,
	FillStyle:   FillDots,
	FillColor:   colorSlate200,
	FontFamily:  FontHandDrawn,
	FontSize:    FontMedium,
	Opacity:     PFloat(1),
}
