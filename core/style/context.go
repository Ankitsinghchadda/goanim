package style

import (
	"image/color"

	"github.com/ankitsinghchadda/goanim/core/render"
)

// Context carries the resolution chain used at render time. A mobject
// resolves its effective style by merging its own override onto the
// scene default onto the library default.
type Context struct {
	// SceneDefault is the per-scene style baseline. Unset fields fall
	// through to the library default.
	SceneDefault Style

	// Fonts maps logical FontFamily values to loaded FontFace handles.
	// At least FontHandDrawn must be populated; FontSans is recommended.
	Fonts map[FontFamily]render.FontFace

	// BgColor is the scene background. Used by icons that need to draw
	// a knock-out region against the background (when their fill is
	// nil — e.g. blueprint preset). When non-nil fill is present,
	// icons knock out against the fill color instead.
	BgColor color.Color

	// OpacityMultiplier is a per-frame multiplier applied to the
	// resolved opacity of each mobject. The Phase-7 direction layer
	// uses it to dim non-focused mobjects during Camera.Focus and
	// (future) Spotlight. Default 0 means "no multiplier" (treated as
	// 1.0); values in [0, 1] dim the mobject. Multipliers > 1 are
	// clamped to 1 (we can't over-brighten).
	//
	// The scene player sets this per-mobject before calling Render
	// based on the camera's focus state.
	OpacityMultiplier float64
}

// NewContext returns a fresh Context with no scene override.
func NewContext() Context {
	return Context{Fonts: map[FontFamily]render.FontFace{}}
}

// Resolve walks override → SceneDefault → library default and returns
// a Style with every field set. Unset fields after resolution have
// taken on their library default value.
//
// After the merge, the context's OpacityMultiplier (when non-zero) is
// folded into the resolved Opacity. That's the hook the direction
// layer uses to dim non-focused mobjects: the scene sets a multiplier
// per-mobject before Render, the resolved style's Opacity carries the
// dim through to every PathStyle that consumes it.
func (c Context) Resolve(override Style) Style {
	merged := override.Merge(c.SceneDefault).Merge(libraryDefault())
	if c.OpacityMultiplier > 0 && c.OpacityMultiplier < 1 {
		base := 1.0
		if merged.Opacity != nil {
			base = *merged.Opacity
		}
		dimmed := base * c.OpacityMultiplier
		merged.Opacity = &dimmed
	}
	return merged
}

// FontFace returns the loaded face for fam, falling back to HandDrawn
// then nil if neither is loaded. The renderer treats a nil face as
// "use the renderer's DefaultFont."
func (c Context) FontFace(fam FontFamily) render.FontFace {
	if face, ok := c.Fonts[fam]; ok && face != nil {
		return face
	}
	if face, ok := c.Fonts[FontHandDrawn]; ok && face != nil {
		return face
	}
	return nil
}

// libraryDefault is the bottom of the inheritance chain. Every field
// is set; no fall-through is possible past this point.
func libraryDefault() Style {
	o := 1.0
	return Style{
		Sloppiness:  SloppinessArtist,
		Edges:       EdgesRound,
		StrokeWidth: StrokeNormal,
		StrokeStyle: StrokeSolid,
		StrokeColor: color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 0xFF},
		FillStyle:   FillNone,
		FillColor:   nil,
		FontFamily:  FontHandDrawn,
		FontSize:    FontMedium,
		Opacity:     &o,
	}
}
