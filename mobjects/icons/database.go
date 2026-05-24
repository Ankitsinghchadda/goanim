package icons

import (
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/icon"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// NewDatabase renders the canonical cylinder: top ellipse, body, bottom
// ellipse. Distinguished from Storage (a flatter, single-ellipse can).
//
// The cylinder is a frame part — it draws its own fill / hatch
// internally. Label sits inside the body, which IconBase's label
// knock-out clears against the hatch.
func NewDatabase(seed int64, label string) *icon.IconBase {
	const w, h = 220, 200
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelInside)
	ic.Add(newCylinder(seed, w, h, 28))
	return ic
}

// NewStorage renders a stack of three flat horizontal disks — like a
// platter stack viewed from the side. This is intentionally distinct
// from Database (a single tall cylinder) so the two icons read as
// different concepts at a glance.
//
// The disks are strokes-only (no fill) so there's no hatch to worry
// about — disk stack is a frame part. Label below.
func NewStorage(seed int64, label string) *icon.IconBase {
	const w, h = 200, 150
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newDiskStack(seed, w, h))
	return ic
}

// diskStack draws three flat disks stacked vertically — the "stack of
// platters" metaphor for object storage.
type diskStack struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newDiskStack(seed int64, w, h float64) *diskStack {
	return &diskStack{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (d *diskStack) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(d.cx, d.cy), d.w, d.h)
}
func (d *diskStack) Children() []mobject.Mobject  { return nil }
func (d *diskStack) Seed() int64                  { return d.seed }
func (d *diskStack) Style() *style.Style          { return d.Group.Style() }
func (d *diskStack) SetStyle(s style.Style)       { d.Group.SetStyle(s) }
func (d *diskStack) Position() (float64, float64) { return d.cx, d.cy }
func (d *diskStack) SetPosition(x, y float64)     { d.cx, d.cy = x, y }
func (d *diskStack) SetReveal(t float64)          { d.reveal = t }
func (d *diskStack) SetVisualScale(float64)       {}

func (d *diskStack) Render(rd render.Renderer, ctx style.Context) {
	if d.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*d.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if d.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*d.reveal)
	}
	// 3 disks at three heights. Each disk = a flat ellipse + two short
	// vertical lines for the edges of the disk's "rim."
	rx := d.w / 2
	ry := 12.0
	gap := (d.h - 2*ry*3) / 2 // vertical gap between disk centers
	cys := []float64{
		d.cy + ry + gap, // top disk
		d.cy,            // middle disk
		d.cy - ry - gap, // bottom disk
	}
	for i, cy := range cys {
		// Disk ellipse outline.
		if tok.Roughness == 0 {
			rd.DrawPath(geometry.EllipsePath(d.cx, cy, rx, ry), stroke)
		} else {
			opts := style.RoughOptions(eff, tok, d.seed+int64(i*7))
			rd.DrawPath(rough.RoughEllipse(d.cx, cy, rx, ry, opts), stroke)
		}
	}
}

// cylinder is a private composite that draws a 3D-looking cylinder.
type cylinder struct {
	*mobject.Group
	seed   int64
	w, h   float64
	capH   float64 // cap depth
	cx, cy float64
	scale  float64
	reveal float64
}

func newCylinder(seed int64, w, h, capH float64) *cylinder {
	return &cylinder{
		Group:  mobject.NewGroup(seed),
		seed:   seed,
		w:      w,
		h:      h,
		capH:   capH,
		scale:  1,
		reveal: 1,
	}
}

func (c *cylinder) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(c.cx, c.cy), c.w*c.scale, c.h*c.scale)
}
func (c *cylinder) Children() []mobject.Mobject  { return nil }
func (c *cylinder) Seed() int64                  { return c.seed }
func (c *cylinder) Style() *style.Style          { return c.Group.Style() }
func (c *cylinder) SetStyle(s style.Style)       { c.Group.SetStyle(s) }
func (c *cylinder) Position() (float64, float64) { return c.cx, c.cy }
func (c *cylinder) SetPosition(x, y float64)     { c.cx, c.cy = x, y }
func (c *cylinder) SetReveal(t float64)          { c.reveal = t }
func (c *cylinder) SetVisualScale(s float64)     { c.scale = s }

func (c *cylinder) Render(rd render.Renderer, ctx style.Context) {
	if c.reveal <= 0 || c.scale <= 0 {
		return
	}
	eff := ctx.Resolve(*c.Group.Style())
	tok := style.TokensFor(eff)

	w := c.w * c.scale
	h := c.h * c.scale
	rx := w / 2
	ry := c.capH * c.scale
	topCY := c.cy + h/2 - ry
	botCY := c.cy - h/2 + ry

	// 1) Body fill, with optional sketchy hatching.
	body := cylinderBodyPolygon(c.cx, topCY, botCY, rx, ry)
	if eff.FillStyle != style.FillNone && eff.FillStyle != style.FillStyleUnset {
		if eff.FillColor != nil {
			rd.DrawPath(rough.SolidFill(body), render.PathStyle{
				Fill: style.ApplyOpacity(eff.FillColor, tok.OpacityScale),
			})
		}
		if eff.FillStyle != style.FillSolid {
			hatchC := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale)
			if eff.StrokeColor == nil {
				hatchC = style.ApplyOpacity(eff.FillColor, tok.OpacityScale)
			}
			opts := style.RoughOptions(eff, tok, c.seed+12001)
			ps := render.PathStyle{Stroke: hatchC, StrokeWidth: tok.FillWidthPx, StrokeCap: render.CapRound}
			switch eff.FillStyle {
			case style.FillHatch:
				rd.DrawPath(rough.Hatch(body, opts), ps)
			case style.FillCrossHatch:
				rd.DrawPath(rough.CrossHatch(body, opts), ps)
			case style.FillZigzag:
				rd.DrawPath(rough.Zigzag(body, opts), ps)
			case style.FillDots:
				rd.DrawPath(rough.Dots(body, opts), render.PathStyle{Fill: hatchC})
			}
		}
	}

	// 2) Outlines.
	stroke := style.PathStyleStroke(eff, tok)
	stroke.DashArray = tok.DashArray
	if c.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*c.reveal)
	}

	rd.DrawPath(c.ellipsePath(rx, ry, eff, tok, c.seed+1).
		Transform(geometry.Translate(c.cx, botCY)), stroke)

	leftSrc, leftDst := geometry.Pt(c.cx-rx, topCY), geometry.Pt(c.cx-rx, botCY)
	rightSrc, rightDst := geometry.Pt(c.cx+rx, topCY), geometry.Pt(c.cx+rx, botCY)
	if tok.Roughness == 0 {
		rd.DrawPath(geometry.LinePath(leftSrc, leftDst), stroke)
		rd.DrawPath(geometry.LinePath(rightSrc, rightDst), stroke)
	} else {
		l := style.RoughOptions(eff, tok, c.seed+31)
		r2 := style.RoughOptions(eff, tok, c.seed+53)
		rd.DrawPath(rough.RoughLine(leftSrc, leftDst, l), stroke)
		rd.DrawPath(rough.RoughLine(rightSrc, rightDst, r2), stroke)
	}

	rd.DrawPath(c.ellipsePath(rx, ry, eff, tok, c.seed+71).
		Transform(geometry.Translate(c.cx, topCY)), stroke)
}

func (c *cylinder) ellipsePath(rx, ry float64, eff style.Style, tok style.Tokens, seed int64) *geometry.Path {
	if tok.Roughness == 0 {
		return geometry.EllipsePath(0, 0, rx, ry)
	}
	opts := style.RoughOptions(eff, tok, seed)
	return rough.RoughEllipse(0, 0, rx, ry, opts)
}

// cylinderBodyPolygon: rectangle between cap centers, bulging down into
// the bottom cap's lower arc.
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
