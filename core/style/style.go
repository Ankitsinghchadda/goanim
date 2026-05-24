package style

import "image/color"

// Sloppiness picks how hand-drawn the strokes look.
type Sloppiness uint8

const (
	SloppinessUnset      Sloppiness = iota
	SloppinessArchitect             // 0 roughness — clean geometric lines
	SloppinessArtist                // ~1 roughness — Excalidraw default
	SloppinessCartoonist            // ~2 roughness — very hand-drawn
)

// Edges picks corner geometry.
type Edges uint8

const (
	EdgesUnset Edges = iota
	EdgesSharp
	EdgesRound
)

// StrokeWidth picks the outline weight (mapped to pixels via Tokens).
type StrokeWidth uint8

const (
	StrokeWidthUnset StrokeWidth = iota
	StrokeThin
	StrokeNormal
	StrokeThick
)

// StrokeStyle picks the outline pattern.
type StrokeStyle uint8

const (
	StrokeStyleUnset StrokeStyle = iota
	StrokeSolid
	StrokeDashed
	StrokeDotted
)

// FillStyle picks the interior pattern.
type FillStyle uint8

const (
	FillStyleUnset FillStyle = iota
	FillNone
	FillSolid
	FillHatch
	FillCrossHatch
	FillZigzag
	FillDots
)

// FontFamily picks the typeface.
type FontFamily uint8

const (
	FontFamilyUnset FontFamily = iota
	FontHandDrawn              // Excalifont — pairs with Artist/Cartoonist
	FontSans                   // Inter — pairs with Architect
	FontMono                   // reserved
)

// FontSize picks the type size (mapped to points via Tokens).
type FontSize uint8

const (
	FontSizeUnset FontSize = iota
	FontSmall
	FontMedium
	FontLarge
	FontXLarge
	FontHuge    // 2× XLarge — large display copy
	FontDisplay // ~3× XLarge — title-card / wordmark sizes
)

// DetailDensity tunes hachure / cross-hatch density for icons that
// carry internal detail elements. Sparser fills let small details
// (lightning bolts, gears, fan-out arrows) read through the pattern.
type DetailDensity uint8

const (
	DetailDensityUnset  DetailDensity = iota
	DetailDensityNone                 // no fill, regardless of FillStyle
	DetailDensityLight                // sparse hachure (gap × ~1.6)
	DetailDensityNormal               // hachure at default gap
	DetailDensityDense                // cross-hatch / dense hachure (gap × ~0.7)
)

// Style is the bundle of attributes a mobject carries.
//
// Every enum field uses an Unset zero value, so `Style{}` denotes
// "inherit all." Numeric fields that can legitimately be zero
// (Opacity, Roughness) are stored as pointers — nil means "inherit."
//
// Colors are nil to inherit; any non-nil image/color.Color overrides.
type Style struct {
	Sloppiness    Sloppiness
	Edges         Edges
	StrokeWidth   StrokeWidth
	StrokeStyle   StrokeStyle
	StrokeColor   color.Color
	FillStyle     FillStyle
	FillColor     color.Color
	FontFamily    FontFamily
	FontSize      FontSize
	DetailDensity DetailDensity
	// Role is the Phase-9 text hierarchy hint. When set, it picks an
	// absolute pixel size from style.SizePxForRole that overrides
	// the FontSize enum. RoleUnset (zero value) falls through to the
	// FontSize enum so legacy text mobjects render as before.
	Role Role

	// Opacity in [0, 1]. nil = inherit.
	Opacity *float64
	// Roughness overrides Sloppiness when non-nil. Useful for fine
	// tuning; most callers should set Sloppiness instead.
	Roughness *float64
}

// Merge returns a copy of s with any Unset fields filled from parent.
// Explicit ("set") fields in s win.
func (s Style) Merge(parent Style) Style {
	out := s
	if out.Sloppiness == SloppinessUnset {
		out.Sloppiness = parent.Sloppiness
	}
	if out.Edges == EdgesUnset {
		out.Edges = parent.Edges
	}
	if out.StrokeWidth == StrokeWidthUnset {
		out.StrokeWidth = parent.StrokeWidth
	}
	if out.StrokeStyle == StrokeStyleUnset {
		out.StrokeStyle = parent.StrokeStyle
	}
	if out.StrokeColor == nil {
		out.StrokeColor = parent.StrokeColor
	}
	if out.FillStyle == FillStyleUnset {
		out.FillStyle = parent.FillStyle
	}
	if out.FillColor == nil {
		out.FillColor = parent.FillColor
	}
	if out.FontFamily == FontFamilyUnset {
		out.FontFamily = parent.FontFamily
	}
	if out.FontSize == FontSizeUnset {
		out.FontSize = parent.FontSize
	}
	if out.DetailDensity == DetailDensityUnset {
		out.DetailDensity = parent.DetailDensity
	}
	if out.Role == RoleUnset {
		out.Role = parent.Role
	}
	if out.Opacity == nil {
		out.Opacity = parent.Opacity
	}
	if out.Roughness == nil {
		out.Roughness = parent.Roughness
	}
	return out
}

// PFloat is a convenience helper for pointer-float fields.
func PFloat(v float64) *float64 { return &v }
