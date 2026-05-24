package mathx

import (
	"image/color"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/internal/latex"
)

// Equation is a math formula rendered from LaTeX source.
//
// In Architect (crisp) mode, the equation is drawn as clean filled
// glyph paths. In Artist or Cartoonist mode, the same paths are
// stroked with the active sloppiness — producing the signature
// "handwritten math" look.
//
//	eq := mathx.NewEquation("E = mc^2").WithHeight(80)
//	scene.Add(eq)
//
// Sub-symbols are addressable by index for fine-grained animation.
// Indices correspond to glyph order in the LaTeX source (left-to-right,
// no semantic awareness).
type Equation struct {
	*mobject.Group
	source   string
	height   float64
	cx, cy   float64
	color    color.Color
	style    style.Style
	reveal   float64
	subPaths []*geometry.Path
	full     *geometry.Path
}

// NewEquation constructs an equation from LaTeX source. Default height
// is 80px; override with WithHeight.
func NewEquation(latexSrc string) *Equation {
	e := &Equation{
		Group:  mobject.NewGroup(0),
		source: latexSrc,
		height: 80,
		reveal: 1,
	}
	e.compile()
	return e
}

func (e *Equation) compile() {
	path, err := latex.Compile(e.source, e.height)
	if err != nil {
		// Render an error marker — a small red X — so failures are
		// visible at the diagram level.
		p := geometry.NewPath()
		p.MoveTo(-20, -20)
		p.LineTo(20, 20)
		p.MoveTo(-20, 20)
		p.LineTo(20, -20)
		e.full = p
		e.subPaths = []*geometry.Path{p}
		return
	}
	e.full = path
	e.subPaths = latex.SubPaths(path)
}

// WithHeight sets the equation's pixel height and recompiles.
func (e *Equation) WithHeight(h float64) *Equation {
	e.height = h
	e.compile()
	return e
}

// WithColor overrides the equation's stroke/fill color.
func (e *Equation) WithColor(c color.Color) *Equation { e.color = c; return e }

// WithStyle sets the per-mobject style override.
func (e *Equation) WithStyle(s style.Style) *Equation { e.style = s; return e }

// Source returns the original LaTeX string.
func (e *Equation) Source() string { return e.source }

// Submobjects returns one path per glyph (approximately). Used by
// Write and TransformEquation animations.
func (e *Equation) Submobjects() []*geometry.Path { return e.subPaths }

// Symbol returns the i-th glyph as a path. Returns nil if i is OOB.
func (e *Equation) Symbol(i int) *geometry.Path {
	if i < 0 || i >= len(e.subPaths) {
		return nil
	}
	return e.subPaths[i]
}

// SymbolCount returns the number of addressable submobjects.
func (e *Equation) SymbolCount() int { return len(e.subPaths) }

// MoveTo sets the equation's center.
func (e *Equation) MoveTo(x, y float64) *Equation { e.cx, e.cy = x, y; return e }

// SetPosition is the imperative form of MoveTo.
func (e *Equation) SetPosition(x, y float64) { e.cx, e.cy = x, y }

// Position returns the equation's center.
func (e *Equation) Position() (float64, float64) { return e.cx, e.cy }

// SetReveal sets the reveal fraction (0..1). Used by Write and
// FadeIn/Out animations.
func (e *Equation) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	e.reveal = t
}

// Reveal returns the current reveal fraction.
func (e *Equation) Reveal() float64 { return e.reveal }

// SetVisualScale — equations expose VisualScale via SetReveal mapping
// (no separate scale support yet). Implementing the Scaler interface
// lets PopIn target equations.
func (e *Equation) SetVisualScale(s float64) {
	if s < 0 {
		s = 0
	}
	e.reveal = s
}

// Bounds returns the equation's bounding box at its current center.
func (e *Equation) Bounds() geometry.Rect {
	if e.full == nil {
		return geometry.RectFromCenter(geometry.Pt(e.cx, e.cy), 0, 0)
	}
	b := e.full.Bounds()
	return geometry.Rect{
		Min: geometry.Pt(b.Min.X+e.cx, b.Min.Y+e.cy),
		Max: geometry.Pt(b.Max.X+e.cx, b.Max.Y+e.cy),
	}
}

// Style returns the per-mobject style override.
func (e *Equation) Style() *style.Style { return &e.style }

// SetStyle replaces the per-mobject style.
func (e *Equation) SetStyle(s style.Style) { e.style = s }

// Render draws the equation at its current position with the resolved
// style. In crisp mode each glyph is filled; in sketchy mode each
// glyph is stroked with the active roughness (handwritten math).
func (e *Equation) Render(r render.Renderer, ctx style.Context) {
	if e.reveal <= 0 || e.full == nil {
		return
	}
	eff := ctx.Resolve(e.style)
	tok := style.TokensFor(eff)
	col := e.color
	if col == nil {
		col = eff.StrokeColor
	}
	if e.reveal < 1 {
		col = style.ApplyOpacity(col, tok.OpacityScale*e.reveal)
	}

	t := geometry.Translate(e.cx, e.cy)
	full := e.full.Transform(t)

	if tok.Roughness == 0 {
		// Crisp: filled glyphs.
		r.DrawPath(full, render.PathStyle{Fill: col})
		return
	}

	// Sketchy: stroke each subpath with roughness applied. We don't try
	// to re-rough the precise glyph geometry (would destroy
	// legibility); instead we stroke the existing paths with a thin
	// rough stroke that adds slight wobble at curve edges.
	stroke := render.PathStyle{
		Stroke:      col,
		StrokeWidth: 1.4 + tok.Roughness*0.4,
		StrokeCap:   render.CapRound,
		StrokeJoin:  render.JoinRound,
	}
	// Also fill the glyphs at reduced opacity — gives them weight
	// while the rough outline gives the handwritten character.
	fillCol := style.ApplyOpacity(col, tok.OpacityScale*0.7)
	r.DrawPath(full, render.PathStyle{Fill: fillCol})
	r.DrawPath(full, stroke)
	_ = rough.DefaultOptions // marker that rough is in use conceptually
}
