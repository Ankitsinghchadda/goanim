package systemdesign

import (
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Database is the cylinder shape — a top ellipse + a bottom ellipse +
// two vertical side lines + a label centered in the body.
//
// Like Client and Server, Database renders according to the resolved
// style: rough double-stroke at Sloppiness > 0; clean cubic-Bezier
// ellipses at Sloppiness == 0.
type Database struct {
	*mobject.Group
	cx, cy float64
	w, h   float64
	label  *mobject.Text
	style  style.Style
	cacheS shapeCacheSegment
	reveal float64
	scale  float64
}

// NewDatabase constructs a Database with the given label.
func NewDatabase(seed int64, labelText string) *Database {
	d := &Database{
		Group:  mobject.NewGroup(seed),
		w:      280,
		h:      240,
		label:  mobject.NewText(seed+777, labelText),
		reveal: 1,
		scale:  1,
	}
	d.Group.Add(d.label)
	return d
}

// Position returns the center.
func (d *Database) Position() (float64, float64) { return d.cx, d.cy }

// SetPosition sets the center (animation hook).
func (d *Database) SetPosition(x, y float64) {
	d.cx, d.cy = x, y
	d.label.SetPosition(x, y)
}

// SetReveal cascades the reveal fraction.
func (d *Database) SetReveal(t float64) {
	d.reveal = clampMobjectLocal(t)
	d.label.SetReveal(t)
}

// SetVisualScale applies a uniform scale.
func (d *Database) SetVisualScale(s float64) { d.scale = s }

// MoveTo sets the center.
func (d *Database) MoveTo(x, y float64) *Database {
	d.SetPosition(x, y)
	return d
}

// WithStyle sets the per-mobject style override.
func (d *Database) WithStyle(s style.Style) *Database {
	d.style = s
	d.cacheS.invalidate()
	return d
}

// Style returns the per-mobject style override (in-place editable).
func (d *Database) Style() *style.Style { return &d.style }

// SetStyle replaces the style override.
func (d *Database) SetStyle(s style.Style) { d.style = s; d.cacheS.invalidate() }

// Bounds includes the cylinder body (sans label, which sits inside it).
func (d *Database) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(d.cx, d.cy), d.w, d.h)
}

// CylinderBounds is what arrows attach to.
func (d *Database) CylinderBounds() geometry.Rect { return d.Bounds() }

// AttachmentPoint returns the attachment point on the cylinder body.
// For top/bottom sides we use the midpoint of the bounding-box edge
// (the cap of the cylinder). For left/right sides we use the ellipse
// boundary intersection so the arrow lands tangentially on the curve.
func (d *Database) AttachmentPoint(side mobject.Side) geometry.Point {
	bb := d.CylinderBounds()
	switch side {
	case mobject.SideTop, mobject.SideBottom:
		return mobject.AttachToBoundsEdge(bb, side)
	default:
		// Use ellipse-aware attachment for the curved sides.
		return mobject.AttachToEllipse(bb.Center().X, bb.Center().Y, bb.Width()/2, bb.Height()/2, side)
	}
}

// Render draws the cylinder body, fill, side lines, and ellipses.
func (d *Database) Render(r render.Renderer, ctx style.Context) {
	if d.reveal <= 0 || d.scale <= 0 {
		return
	}
	eff := ctx.Resolve(d.style)
	tok := style.TokensFor(eff)

	w := d.w * d.scale
	h := d.h * d.scale

	rx := w / 2
	ry := 36.0 * d.scale // cap depth scales with the body

	topCY := d.cy + h/2 - ry
	botCY := d.cy - h/2 + ry

	// 1. Body fill. Solid first (so back-half of bottom ellipse is
	// masked) when a fill color is provided. Then optional sketchy
	// pattern. nil FillColor → no solid underlay, just pattern.
	body := cylinderBodyPolygon(d.cx, topCY, botCY, rx, ry)
	if eff.FillStyle != style.FillNone && eff.FillStyle != style.FillStyleUnset {
		if eff.FillColor != nil {
			r.DrawPath(rough.SolidFill(body), render.PathStyle{
				Fill: style.ApplyOpacity(eff.FillColor, tok.OpacityScale),
			})
		}
		if eff.FillStyle != style.FillSolid {
			hatchColor := style.ApplyOpacity(hatchColorOrStroke(eff), tok.OpacityScale)
			opts := style.RoughOptions(eff, tok, d.Seed()+12001)
			ps := render.PathStyle{
				Stroke:      hatchColor,
				StrokeWidth: tok.FillWidthPx,
				StrokeCap:   render.CapRound,
			}
			switch eff.FillStyle {
			case style.FillHatch:
				r.DrawPath(rough.Hatch(body, opts), ps)
			case style.FillCrossHatch:
				r.DrawPath(rough.CrossHatch(body, opts), ps)
			case style.FillZigzag:
				r.DrawPath(rough.Zigzag(body, opts), ps)
			case style.FillDots:
				r.DrawPath(rough.Dots(body, opts), render.PathStyle{Fill: hatchColor})
			}
		}
	}

	// 3. Outlines. Bottom first, then sides, then top.
	stroke := style.PathStyleStroke(eff, tok)
	stroke.DashArray = tok.DashArray

	bottomPath := d.cacheS.ellipsePath(d.Seed()+1, d.w/2, 36, eff, tok).
		Transform(geometry.Scale(d.scale)).
		Transform(geometry.Translate(d.cx, botCY))
	r.DrawPath(bottomPath, applyOpacityToStroke(stroke, eff, tok, d.reveal))

	// Sides.
	leftSrc := geometry.Pt(d.cx-rx, topCY)
	leftDst := geometry.Pt(d.cx-rx, botCY)
	rightSrc := geometry.Pt(d.cx+rx, topCY)
	rightDst := geometry.Pt(d.cx+rx, botCY)
	var leftPath, rightPath *geometry.Path
	if tok.Roughness == 0 {
		leftPath = geometry.LinePath(leftSrc, leftDst)
		rightPath = geometry.LinePath(rightSrc, rightDst)
	} else {
		lopts := style.RoughOptions(eff, tok, d.Seed()+31)
		ropts := style.RoughOptions(eff, tok, d.Seed()+53)
		leftPath = rough.RoughLine(leftSrc, leftDst, lopts)
		rightPath = rough.RoughLine(rightSrc, rightDst, ropts)
	}
	r.DrawPath(leftPath, applyOpacityToStroke(stroke, eff, tok, d.reveal))
	r.DrawPath(rightPath, applyOpacityToStroke(stroke, eff, tok, d.reveal))

	topPath := d.cacheS.ellipsePathTop(d.Seed()+71, d.w/2, 36, eff, tok).
		Transform(geometry.Scale(d.scale)).
		Transform(geometry.Translate(d.cx, topCY))
	r.DrawPath(topPath, applyOpacityToStroke(stroke, eff, tok, d.reveal))

	// 4. Label.
	d.label.Render(r, ctx)
}

// applyOpacityToStroke returns a copy of ps with the stroke color
// scaled by reveal. We opacity-fade DrawOn-style reveals on closed
// shapes; path truncation interacts poorly with canvas's closed-path
// fill detection.
func applyOpacityToStroke(ps render.PathStyle, eff style.Style, tok style.Tokens, reveal float64) render.PathStyle {
	if reveal >= 1 {
		return ps
	}
	out := ps
	out.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*reveal)
	return out
}

// hatchColorOrStroke is the color used for decorative fill patterns;
// defaults to the stroke color so hatching reads as shading.
func hatchColorOrStroke(eff style.Style) interface {
	RGBA() (uint32, uint32, uint32, uint32)
} {
	if eff.StrokeColor != nil {
		return eff.StrokeColor
	}
	return eff.FillColor
}

// cylinderBodyPolygon traces the visible (front) body region of a
// cylinder: the rectangle between the two cap centers, bulging down
// into the bottom cap's lower arc.
func cylinderBodyPolygon(cx, topCY, botCY, rx, ry float64) []geometry.Point {
	const samples = 24
	pts := make([]geometry.Point, 0, samples+4)
	pts = append(pts, geometry.Pt(cx-rx, topCY))
	pts = append(pts, geometry.Pt(cx-rx, botCY))
	for i := 1; i < samples; i++ {
		theta := math.Pi + float64(i)*math.Pi/float64(samples)
		pts = append(pts, geometry.Pt(
			cx+rx*math.Cos(theta),
			botCY+ry*math.Sin(theta),
		))
	}
	pts = append(pts, geometry.Pt(cx+rx, botCY))
	pts = append(pts, geometry.Pt(cx+rx, topCY))
	return pts
}

// shapeCacheSegment caches the two ellipse paths used by Database
// (top and bottom). Same logic as core/mobject/cache.go but private
// to this package so we don't have to expose internals.
type shapeCacheSegment struct {
	bottomKey cylKey
	bottom    *geometry.Path
	topKey    cylKey
	topPath   *geometry.Path
}

type cylKey struct {
	seed      int64
	rx, ry    float64
	roughness float64
	bowing    float64
	jit       float64
	strokeW   float64
}

func (c *shapeCacheSegment) invalidate() {
	c.bottom = nil
	c.topPath = nil
}

func (c *shapeCacheSegment) keyFor(seed int64, rx, ry float64, tok style.Tokens) cylKey {
	return cylKey{seed: seed, rx: rx, ry: ry,
		roughness: tok.Roughness, bowing: tok.Bowing,
		jit: tok.MaxJitter, strokeW: tok.StrokeWidthPx}
}

func (c *shapeCacheSegment) ellipsePath(seed int64, rx, ry float64, eff style.Style, tok style.Tokens) *geometry.Path {
	k := c.keyFor(seed, rx, ry, tok)
	if c.bottom != nil && c.bottomKey == k {
		return c.bottom
	}
	c.bottom = buildEllipsePath(seed, rx, ry, eff, tok)
	c.bottomKey = k
	return c.bottom
}

func (c *shapeCacheSegment) ellipsePathTop(seed int64, rx, ry float64, eff style.Style, tok style.Tokens) *geometry.Path {
	k := c.keyFor(seed, rx, ry, tok)
	if c.topPath != nil && c.topKey == k {
		return c.topPath
	}
	c.topPath = buildEllipsePath(seed, rx, ry, eff, tok)
	c.topKey = k
	return c.topPath
}

func buildEllipsePath(seed int64, rx, ry float64, eff style.Style, tok style.Tokens) *geometry.Path {
	if tok.Roughness == 0 {
		return geometry.EllipsePath(0, 0, rx, ry)
	}
	opts := style.RoughOptions(eff, tok, seed)
	return rough.RoughEllipse(0, 0, rx, ry, opts)
}
