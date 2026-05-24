package mobject

import (
	"image/color"
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Rectangle is a style-driven rectangle (rough or clean depending on
// resolved Sloppiness).
type Rectangle struct {
	seed   int64
	cx, cy float64
	w, h   float64
	style  style.Style
	reveal float64 // 0 = hidden; 1 = fully drawn; default 1
	scale  float64 // visual scale; default 1

	// Phase-2 temporal-stability cache. The rough path is generated
	// once for a given (style, w, h, seed) and reused across frames.
	// Translation/rotation/uniform-scale don't invalidate this cache.
	cache shapeCache
}

// NewRectangle constructs a Rectangle of width w and height h centered
// at the origin. Use MoveTo to position; use With* setters to style.
func NewRectangle(seed int64, w, h float64) *Rectangle {
	return &Rectangle{seed: seed, w: w, h: h, reveal: 1, scale: 1}
}

func (r *Rectangle) MoveTo(x, y float64) *Rectangle { r.cx, r.cy = x, y; return r }
func (r *Rectangle) SetPosition(x, y float64)       { r.cx, r.cy = x, y }
func (r *Rectangle) Position() (float64, float64)   { return r.cx, r.cy }
func (r *Rectangle) Size() (float64, float64)       { return r.w, r.h }
func (r *Rectangle) SetSize(w, h float64)           { r.w, r.h = w, h; r.cache.invalidate() }
func (r *Rectangle) SetReveal(t float64)            { r.reveal = clampMobject(t) }
func (r *Rectangle) Reveal() float64                { return r.reveal }
func (r *Rectangle) SetVisualScale(s float64)       { r.scale = s }
func (r *Rectangle) VisualScale() float64           { return r.scale }
func (r *Rectangle) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(r.cx, r.cy), r.w, r.h)
}

// AttachmentPoint returns the midpoint of the named edge of the
// rectangle's bounding box, in user-space coordinates.
func (r *Rectangle) AttachmentPoint(side Side) geometry.Point {
	return boundsEdgeMidpoint(r.Bounds(), side)
}
func (r *Rectangle) Children() []Mobject { return nil }
func (r *Rectangle) Seed() int64         { return r.seed }
func (r *Rectangle) Style() *style.Style { return &r.style }
func (r *Rectangle) SetStyle(s style.Style) {
	r.style = s
	r.cache.invalidate()
}

// WithStyle replaces the per-mobject style and returns the receiver.
func (r *Rectangle) WithStyle(s style.Style) *Rectangle { r.SetStyle(s); return r }

// Render dispatches on resolved Sloppiness: Architect → clean geometric
// path; Artist/Cartoonist → seeded rough geometry, cached for stability.
//
// Reveal and VisualScale modulate appearance: reveal < 1 truncates the
// outline to that fraction of total length and fades the fill; scale
// != 1 applies a uniform scale around the center.
func (r *Rectangle) Render(rd render.Renderer, ctx style.Context) {
	if r.reveal <= 0 || r.scale <= 0 {
		return
	}
	eff := ctx.Resolve(r.style)
	tok := style.TokensFor(eff)

	w := r.w * r.scale
	h := r.h * r.scale
	x0, y0 := r.cx-w/2, r.cy-h/2

	var outline *geometry.Path
	if tok.Roughness == 0 {
		outline = geometry.RectanglePath(x0, y0, w, h, tok.CornerRadius*r.scale)
	} else {
		// Rough path is cached at the mobject's NATURAL dimensions for
		// temporal stability. We apply scale by transforming the
		// cached path before translating to position.
		rp := r.cache.roughRect(r.seed, r.w, r.h, eff, tok)
		t := geometry.Translate(r.cx, r.cy).Compose(geometry.Scale(r.scale))
		outline = rp.Transform(t)
	}

	polygon := rough.RectToPolygon(x0, y0, w, h)
	drawFilledStrokedReveal(rd, outline, eff, tok, polygon, r.seed, r.reveal)
}

// Ellipse is a style-driven ellipse.
type Ellipse struct {
	seed   int64
	cx, cy float64
	rx, ry float64
	style  style.Style
	reveal float64
	scale  float64
	cache  shapeCache
}

// NewEllipse constructs an Ellipse of x-radius rx and y-radius ry.
func NewEllipse(seed int64, rx, ry float64) *Ellipse {
	return &Ellipse{seed: seed, rx: rx, ry: ry, reveal: 1, scale: 1}
}

func (e *Ellipse) MoveTo(x, y float64) *Ellipse { e.cx, e.cy = x, y; return e }
func (e *Ellipse) SetPosition(x, y float64)     { e.cx, e.cy = x, y }
func (e *Ellipse) Position() (float64, float64) { return e.cx, e.cy }
func (e *Ellipse) Radii() (float64, float64)    { return e.rx, e.ry }
func (e *Ellipse) SetRadii(rx, ry float64)      { e.rx, e.ry = rx, ry; e.cache.invalidate() }
func (e *Ellipse) SetReveal(t float64)          { e.reveal = clampMobject(t) }
func (e *Ellipse) Reveal() float64              { return e.reveal }
func (e *Ellipse) SetVisualScale(s float64)     { e.scale = s }
func (e *Ellipse) VisualScale() float64         { return e.scale }
func (e *Ellipse) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(e.cx, e.cy), e.rx*2, e.ry*2)
}

// AttachmentPoint returns the point on the ellipse boundary along the
// line from the center toward the named bounding-edge midpoint.
func (e *Ellipse) AttachmentPoint(side Side) geometry.Point {
	return ellipseEdgeFromSide(e.cx, e.cy, e.rx, e.ry, side)
}
func (e *Ellipse) Children() []Mobject              { return nil }
func (e *Ellipse) Seed() int64                      { return e.seed }
func (e *Ellipse) Style() *style.Style              { return &e.style }
func (e *Ellipse) SetStyle(s style.Style)           { e.style = s; e.cache.invalidate() }
func (e *Ellipse) WithStyle(s style.Style) *Ellipse { e.SetStyle(s); return e }

func (e *Ellipse) Render(rd render.Renderer, ctx style.Context) {
	if e.reveal <= 0 || e.scale <= 0 {
		return
	}
	eff := ctx.Resolve(e.style)
	tok := style.TokensFor(eff)
	rx := e.rx * e.scale
	ry := e.ry * e.scale
	var outline *geometry.Path
	if tok.Roughness == 0 {
		outline = geometry.EllipsePath(e.cx, e.cy, rx, ry)
	} else {
		rp := e.cache.roughEllipse(e.seed, e.rx, e.ry, eff, tok)
		t := geometry.Translate(e.cx, e.cy).Compose(geometry.Scale(e.scale))
		outline = rp.Transform(t)
	}
	poly := rough.EllipseToPolygon(e.cx, e.cy, rx, ry, 64)
	drawFilledStrokedReveal(rd, outline, eff, tok, poly, e.seed, e.reveal)
}

// Line is a single line between two points.
type Line struct {
	seed   int64
	p1, p2 geometry.Point
	style  style.Style
	reveal float64
}

func NewLine(seed int64, p1, p2 geometry.Point) *Line {
	return &Line{seed: seed, p1: p1, p2: p2, reveal: 1}
}
func (l *Line) From() geometry.Point          { return l.p1 }
func (l *Line) To() geometry.Point            { return l.p2 }
func (l *Line) SetFrom(p geometry.Point)      { l.p1 = p }
func (l *Line) SetTo(p geometry.Point)        { l.p2 = p }
func (l *Line) Bounds() geometry.Rect         { return geometry.RectFromPoints(l.p1, l.p2) }
func (l *Line) Children() []Mobject           { return nil }
func (l *Line) Seed() int64                   { return l.seed }
func (l *Line) Style() *style.Style           { return &l.style }
func (l *Line) SetStyle(s style.Style)        { l.style = s }
func (l *Line) SetReveal(t float64)           { l.reveal = clampMobject(t) }
func (l *Line) Reveal() float64               { return l.reveal }
func (l *Line) WithStyle(s style.Style) *Line { l.SetStyle(s); return l }

func (l *Line) Render(rd render.Renderer, ctx style.Context) {
	if l.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(l.style)
	tok := style.TokensFor(eff)
	var outline *geometry.Path
	if tok.Roughness == 0 {
		outline = geometry.LinePath(l.p1, l.p2)
	} else {
		opts := style.RoughOptions(eff, tok, l.seed)
		outline = rough.RoughLine(l.p1, l.p2, opts)
	}
	if l.reveal < 1 {
		outline = geometry.PathPrefix(outline, geometry.PathLength(outline)*l.reveal)
	}
	rd.DrawPath(outline, style.PathStyleStroke(eff, tok))
}

// clampMobject clamps a float to [0, 1]; package-private helper.
func clampMobject(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

// drawFilledStrokedReveal renders a shape with progressive disclosure
// for the DrawOn animation:
//
//	reveal == 0    →  nothing
//	0 < reveal < 1 →  outline drawn with opacity = reveal^0.7 (a
//	                  power curve that emphasizes the early stroke);
//	                  fill held back until reveal >= 0.6 then ramps to 1
//	reveal == 1    →  fully drawn at native opacity
//
// We opacity-fade rather than truncate the path because closed shapes
// (rectangles, ellipses) trip up the canvas backend's fill detection
// when partially truncated — the empty "almost-closed" path gets
// flood-filled with the stroke color. Opacity reveal is more robust and
// visually still feels like a sketch coming to life when combined with
// the fill ramp.
func drawFilledStrokedReveal(rd render.Renderer, outlinePath *geometry.Path, eff style.Style, tok style.Tokens, polygon []geometry.Point, seed int64, reveal float64) {
	if reveal <= 0 {
		return
	}

	// Fill is binary — drawn only when reveal >= 0.6 (no opacity ramp).
	// Sub-pixel transparency on filled regions runs into a color-space
	// quirk in tdewolff/canvas where partial-alpha colors get drawn at
	// near-zero brightness; keeping fills at full opacity sidesteps it.
	const fillStart = 0.6
	if reveal >= fillStart {
		drawFilled(rd, eff, tok, polygon, seed)
	}

	// Outline: opacity = reveal^0.7 (early ramp, taper to 1).
	outlineOpacity := 1.0
	if reveal < 1 {
		outlineOpacity = math.Pow(reveal, 0.7)
	}
	tokOut := tok
	tokOut.OpacityScale = tok.OpacityScale * outlineOpacity
	ps := style.PathStyleStroke(eff, tokOut)
	ps.DashArray = tok.DashArray
	ps.Fill = nil
	rd.DrawPath(outlinePath, ps)
}

// drawFilled does just the style-appropriate fill.
//
// Special case: when FillStyle is hatch-like and FillColor is nil, the
// hatch is drawn directly on the background (no solid underlay).
func drawFilled(rd render.Renderer, eff style.Style, tok style.Tokens, polygon []geometry.Point, seed int64) {
	if eff.FillStyle != style.FillNone && eff.FillStyle != style.FillStyleUnset {
		switch eff.FillStyle {
		case style.FillSolid:
			if eff.FillColor != nil {
				rd.DrawPath(rough.SolidFill(polygon), render.PathStyle{
					Fill: style.ApplyOpacity(eff.FillColor, tok.OpacityScale),
				})
			}
		case style.FillHatch, style.FillCrossHatch, style.FillZigzag, style.FillDots:
			// Solid pale background under sketchy patterns when one is
			// provided. When FillColor is nil, skip the underlay — the
			// hatch lines render directly on the background.
			if eff.FillColor != nil {
				rd.DrawPath(rough.SolidFill(polygon), render.PathStyle{
					Fill: style.ApplyOpacity(eff.FillColor, tok.OpacityScale),
				})
			}
			opts := style.RoughOptions(eff, tok, seed+9001)
			fillStrokeColor := style.ApplyOpacity(hatchStrokeColor(eff), tok.OpacityScale)
			fillStyleSpec := render.PathStyle{
				Stroke:      fillStrokeColor,
				StrokeWidth: tok.FillWidthPx,
				StrokeCap:   render.CapRound,
			}
			switch eff.FillStyle {
			case style.FillHatch:
				rd.DrawPath(rough.Hatch(polygon, opts), fillStyleSpec)
			case style.FillCrossHatch:
				rd.DrawPath(rough.CrossHatch(polygon, opts), fillStyleSpec)
			case style.FillZigzag:
				rd.DrawPath(rough.Zigzag(polygon, opts), fillStyleSpec)
			case style.FillDots:
				rd.DrawPath(rough.Dots(polygon, opts), render.PathStyle{
					Fill: fillStrokeColor,
				})
			}
		}
	}
}

// hatchStrokeColor picks the color used to draw hatch / cross-hatch /
// zigzag / dots strokes. Defaults to the resolved stroke color so the
// hatching reads as a darker shaded version of the fill region.
func hatchStrokeColor(eff style.Style) color.Color {
	if eff.StrokeColor != nil {
		return eff.StrokeColor
	}
	return eff.FillColor
}
