package style

import (
	"image/color"
	"math"

	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
)

// Tokens are pixel-and-numeric values derived from a resolved Style.
//
// Tokens are the renderer-facing form of a style. Mobjects don't draw
// "FontMedium"; they draw text at TokenSize() pixels. The mapping
// depends on a few attributes — rough strokes look better slightly
// thicker, so StrokeNormal at Cartoonist sloppiness is wider than at
// Architect.
type Tokens struct {
	Roughness     float64 // 0, ~1, or ~2 depending on Sloppiness
	Bowing        float64
	MaxJitter     float64 // rough.Options.MaxRandomnessOffset
	StrokeWidthPx float64
	FillWidthPx   float64 // hatch / cross-hatch line width
	HachureGap    float64
	HachureAngle  float64
	CornerRadius  float64 // 0 if EdgesSharp
	FontSizePx    float64
	OpacityScale  float64 // [0, 1]
	DashArray     []float64
}

// TokensFor derives drawing constants from a resolved style.
func TokensFor(s Style) Tokens {
	// Sloppiness → roughness. Roughness override wins if set.
	rough := 1.0
	switch s.Sloppiness {
	case SloppinessArchitect:
		rough = 0
	case SloppinessArtist:
		rough = 1.0
	case SloppinessCartoonist:
		rough = 2.0
	}
	if s.Roughness != nil {
		rough = *s.Roughness
	}

	maxJitter := 2.0
	bowing := 1.0
	if rough >= 2 {
		maxJitter = 3.5
		bowing = 1.5
	} else if rough >= 1 {
		maxJitter = 2.5
		bowing = 1.2
	}

	// Stroke widths are slightly wider at higher sloppiness so the
	// double-stroke still reads at distance.
	sw := strokeWidthPx(s.StrokeWidth)
	if rough >= 1 {
		sw += 0.5
	}
	if rough >= 2 {
		sw += 0.5
	}

	// Hachure gap scales with stroke weight so dense strokes don't
	// crowd thin lines. Cartoonist sloppiness gets +50% so the fill
	// reads as "scribbled" rather than "solidly inked" — real
	// hand-drawn sketches have GAPS in their hatching.
	gap := math.Max(sw*4, 8)
	if rough >= 2 {
		gap *= 1.5
	}

	// DetailDensity is a per-icon override applied AFTER sloppiness
	// scaling. Icons with internal details (bolt, gear, fan-out)
	// default to Light so the details can read through the fill.
	switch s.DetailDensity {
	case DetailDensityLight:
		gap *= 1.6
	case DetailDensityDense:
		gap *= 0.7
	}

	cornerR := 0.0
	if s.Edges == EdgesRound {
		cornerR = 12.0
	}

	op := 1.0
	if s.Opacity != nil {
		op = *s.Opacity
	}

	var dashes []float64
	switch s.StrokeStyle {
	case StrokeDashed:
		dashes = []float64{sw * 4, sw * 3}
	case StrokeDotted:
		dashes = []float64{sw * 0.6, sw * 2.4}
	}

	// Font size: Role wins over the FontSize enum when set. Phase-9
	// hierarchy work — Role gives style-aware (sketchy vs crisp)
	// absolute sizes that the FontSize enum can't express without
	// adding 7 new enum variants. Legacy callers leave Role unset
	// and continue to drive size via FontSize.
	fontPx := SizePxForRole(s.Role, s.Sloppiness)
	if fontPx == 0 {
		fontPx = fontSizePx(s.FontSize)
	}

	return Tokens{
		Roughness:     rough,
		Bowing:        bowing,
		MaxJitter:     maxJitter,
		StrokeWidthPx: sw,
		FillWidthPx:   math.Max(sw*0.7, 1.0),
		HachureGap:    gap,
		HachureAngle:  -41,
		CornerRadius:  cornerR,
		FontSizePx:    fontPx,
		OpacityScale:  op,
		DashArray:     dashes,
	}
}

func strokeWidthPx(w StrokeWidth) float64 {
	switch w {
	case StrokeThin:
		return 1.5
	case StrokeThick:
		return 4.0
	default: // Normal
		return 2.5
	}
}

func fontSizePx(s FontSize) float64 {
	switch s {
	case FontSmall:
		return 32
	case FontLarge:
		return 64
	case FontXLarge:
		return 96
	case FontHuge:
		return 192
	case FontDisplay:
		return 280
	default: // Medium
		return 48
	}
}

// ApplyOpacity returns c with alpha multiplied by op.
func ApplyOpacity(c color.Color, op float64) color.Color {
	if c == nil || op >= 0.999 {
		return c
	}
	if op <= 0 {
		return color.RGBA{}
	}
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(float64(a>>8) * op),
	}
}

// PathStyleStroke returns a render.PathStyle that strokes-only with
// the resolved style's stroke color and tokens.
func PathStyleStroke(s Style, t Tokens) render.PathStyle {
	ps := render.PathStyle{
		Stroke:      ApplyOpacity(s.StrokeColor, t.OpacityScale),
		StrokeWidth: t.StrokeWidthPx,
		StrokeCap:   render.CapRound,
		StrokeJoin:  render.JoinRound,
	}
	if s.Edges == EdgesSharp {
		ps.StrokeJoin = render.JoinMiter
	}
	return ps
}

// PathStyleFill returns a render.PathStyle that fills-only with the
// resolved style's fill color.
func PathStyleFill(s Style, t Tokens) render.PathStyle {
	return render.PathStyle{Fill: ApplyOpacity(s.FillColor, t.OpacityScale)}
}

// RoughOptions builds rough.Options from a resolved style + tokens +
// seed, for shapes that should be drawn via the rough engine.
func RoughOptions(s Style, t Tokens, seed int64) rough.Options {
	o := rough.DefaultOptions()
	o.Seed = seed
	o.Roughness = t.Roughness
	o.Bowing = t.Bowing
	o.MaxRandomnessOffset = t.MaxJitter
	o.StrokeWidth = t.StrokeWidthPx
	o.HachureGap = t.HachureGap
	o.HachureAngle = t.HachureAngle
	o.Stroke = s.StrokeColor
	o.Fill = s.FillColor
	o.PreserveVertices = (s.Edges == EdgesSharp)
	return o
}
