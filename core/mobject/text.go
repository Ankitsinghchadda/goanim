package mobject

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Text is a single line of text positioned at (x, y).
//
// The font face and size come from the resolved style (FontFamily +
// FontSize). The text color comes from the stroke color, since "stroke"
// is the natural mapping for type. Override by setting a per-mobject
// StrokeColor.
type Text struct {
	seed     int64
	cx, cy   float64
	Content  string
	Align    render.TextAlign
	Baseline render.TextBaseline
	style    style.Style
	reveal   float64
}

// NewText constructs a centered, middle-baseline text run.
func NewText(seed int64, content string) *Text {
	return &Text{
		seed:     seed,
		Content:  content,
		Align:    render.AlignCenter,
		Baseline: render.BaselineMiddle,
		reveal:   1,
	}
}

// MoveTo sets the anchor point.
func (t *Text) MoveTo(x, y float64) *Text { t.cx, t.cy = x, y; return t }

// SetPosition is the imperative form of MoveTo for animation use.
func (t *Text) SetPosition(x, y float64) { t.cx, t.cy = x, y }

// Position returns the anchor point.
func (t *Text) Position() (float64, float64) { return t.cx, t.cy }

// SetReveal sets the reveal fraction. For Text, this fades the
// overall opacity rather than revealing glyph-by-glyph.
func (t *Text) SetReveal(v float64) { t.reveal = clampMobject(v) }

// Reveal returns the current reveal fraction.
func (t *Text) Reveal() float64 { return t.reveal }

// Bounds returns a crude bounding box estimated from glyph count and
// font size. Suitable for label placement; not precise hit testing.
func (t *Text) Bounds() geometry.Rect {
	tok := style.TokensFor(t.style.Merge(style.PresetExcalidraw))
	w := float64(len(t.Content)) * tok.FontSizePx * 0.55
	h := tok.FontSizePx * 1.2
	return geometry.RectFromCenter(geometry.Pt(t.cx, t.cy), w, h)
}

func (t *Text) Children() []Mobject           { return nil }
func (t *Text) Seed() int64                   { return t.seed }
func (t *Text) Style() *style.Style           { return &t.style }
func (t *Text) SetStyle(s style.Style)        { t.style = s }
func (t *Text) WithStyle(s style.Style) *Text { t.SetStyle(s); return t }

// WithRole sets the Phase-9 text role (Title, Heading, Body, Label, …)
// which picks a style-aware absolute pixel size from
// style.SizePxForRole. Use this instead of WithStyle when the only
// thing you need from the style system is "this is a title" /
// "this is a label" / etc.; the library handles the actual size,
// adapting to sketchy vs crisp metrics.
//
//	mobject.NewText(0, "BGMI Architecture").WithRole(style.RoleTitle)
func (t *Text) WithRole(r style.Role) *Text {
	t.style.Role = r
	return t
}

func (t *Text) Render(r render.Renderer, ctx style.Context) {
	if t.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(t.style)
	tok := style.TokensFor(eff)
	// Label fades in with reveal — cube it so it stays mostly invisible
	// until the parent shape's outline is well underway.
	op := tok.OpacityScale * t.reveal * t.reveal * t.reveal
	r.DrawText(t.Content, t.cx, t.cy, render.TextStyle{
		Face:     ctx.FontFace(eff.FontFamily),
		Size:     tok.FontSizePx,
		Color:    style.ApplyOpacity(eff.StrokeColor, op),
		Align:    t.Align,
		Baseline: t.Baseline,
	})
}
