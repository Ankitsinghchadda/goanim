package icons

import (
	"image/color"
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/icon"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// NewLoadBalancer is a rectangle with three small fan-out arrows on
// the right side, indicating the LB distributes traffic.
//
// Label is positioned BELOW: with the fan-out arrows occupying the
// right side and the icon body relatively short, a centered label
// gets crowded against the arrows under dense hatching.
func NewLoadBalancer(seed int64, label string) *icon.IconBase {
	const w, h = 220, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newFanout(seed+701, w, h))
	return ic
}

// NewCDN renders a clean circle with four dots positioned around its
// perimeter — the "geographic distribution" metaphor. Simpler than the
// previous grid pattern so it reads clearly in sketchy modes.
//
// The four perimeter dots emit their own per-dot knock-outs in
// Render (we can't use IconBase's knock-out: its Bounds covers the
// whole ellipse, so the halo would erase the entire hatch fill).
// Keep them as a frame part with self-haloing.
func NewCDN(seed int64, label string) *icon.IconBase {
	const w, h = 160, 160
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewEllipse(seed, w/2-10, h/2-10))
	ic.Add(newCDNDetail(seed+801, w, h))
	return ic
}

// NewUser is a simple person silhouette: a head circle and a shoulders
// arc/trapezoid below.
//
// The icon's body dimensions are chosen to tightly enclose the visual
// (head + shoulders), so arrows attach to the actual person shape
// rather than to empty space inside a looser bounding box.
//
// No frame rectangle, so no fill / hatch to knock through — the
// person is a frame part.
func NewUser(seed int64, label string) *icon.IconBase {
	const w, h = 110, 150
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newPerson(seed+901, w, h))
	return ic
}

// fanout draws three small arrows fanning out from the right side of
// an icon body.
type fanout struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newFanout(seed int64, w, h float64) *fanout {
	return &fanout{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

// Fanout geometry constants. Kept in one place so Bounds and Render
// agree on extent — earlier they drifted, leaving the right-edge
// arrowheads outside the knock-out halo and invisible against the
// hatch.
const (
	fanoutBaseInset = 36 // base of leftmost arrow, measured from right edge of body
	fanoutShaftLen  = 30 // shaft length
	fanoutTipFanOut = 22 // vertical span of outer arrow tips (px from center)
	fanoutHeadLen   = 9  // arrowhead length
	fanoutPadX      = 6  // horizontal padding for Bounds
	fanoutPadY      = 8  // vertical padding for Bounds
)

// Bounds returns a tight rectangle around the three fan-out arrows
// on the right side of the icon body, used by IconBase's knock-out
// pass. Includes arrowhead extent + padding so the halo always
// covers the visible glyph.
func (f *fanout) Bounds() geometry.Rect {
	baseX := f.cx + f.w/2 - fanoutBaseInset
	tipX := baseX + fanoutShaftLen
	return geometry.Rect{
		Min: geometry.Pt(baseX-fanoutPadX, f.cy-fanoutTipFanOut-fanoutPadY),
		Max: geometry.Pt(tipX+fanoutHeadLen+fanoutPadX, f.cy+fanoutTipFanOut+fanoutPadY),
	}
}
func (f *fanout) Children() []mobject.Mobject  { return nil }
func (f *fanout) Seed() int64                  { return f.seed }
func (f *fanout) Style() *style.Style          { return f.Group.Style() }
func (f *fanout) SetStyle(s style.Style)       { f.Group.SetStyle(s) }
func (f *fanout) Position() (float64, float64) { return f.cx, f.cy }
func (f *fanout) SetPosition(x, y float64)     { f.cx, f.cy = x, y }
func (f *fanout) SetReveal(t float64)          { f.reveal = t }
func (f *fanout) SetVisualScale(float64)       {}

func (f *fanout) Render(rd render.Renderer, ctx style.Context) {
	if f.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*f.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	// Keep arrows at full stroke weight — they're a primary visual cue
	// for the LB, not a corner ornament. Earlier 0.7× scaling made them
	// vanish in Cartoonist cross-hatch even with knock-out.
	if f.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*f.reveal)
	}

	// Three short divergent arrows from the right side. Geometry uses
	// the package constants so Bounds() stays in sync.
	baseX := f.cx + f.w/2 - fanoutBaseInset
	tipX := baseX + fanoutShaftLen
	for i, dy := range []float64{-fanoutTipFanOut, 0, fanoutTipFanOut} {
		from := geometry.Pt(baseX, f.cy+dy*0.4)
		to := geometry.Pt(tipX, f.cy+dy)
		rd.DrawPath(makeLine(from, to, tok, eff, f.seed+int64(i)), stroke)
		// Arrowhead at the tip.
		dir := to.Sub(from).Normalize()
		const a = 0.55
		cosA, sinA := math.Cos(a), math.Sin(a)
		back := dir.Scale(-1)
		left := geometry.Pt(back.X*cosA-back.Y*sinA, back.X*sinA+back.Y*cosA).Scale(fanoutHeadLen)
		right := geometry.Pt(back.X*cosA+back.Y*sinA, -back.X*sinA+back.Y*cosA).Scale(fanoutHeadLen)
		rd.DrawPath(geometry.LinePath(to, to.Add(left)), stroke)
		rd.DrawPath(geometry.LinePath(to, to.Add(right)), stroke)
	}
}

// cdnDetail draws the latitude/longitude lines inside a CDN globe.
type cdnDetail struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newCDNDetail(seed int64, w, h float64) *cdnDetail {
	return &cdnDetail{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (c *cdnDetail) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(c.cx, c.cy), c.w, c.h)
}
func (c *cdnDetail) Children() []mobject.Mobject  { return nil }
func (c *cdnDetail) Seed() int64                  { return c.seed }
func (c *cdnDetail) Style() *style.Style          { return c.Group.Style() }
func (c *cdnDetail) SetStyle(s style.Style)       { c.Group.SetStyle(s) }
func (c *cdnDetail) Position() (float64, float64) { return c.cx, c.cy }
func (c *cdnDetail) SetPosition(x, y float64)     { c.cx, c.cy = x, y }
func (c *cdnDetail) SetReveal(t float64)          { c.reveal = t }
func (c *cdnDetail) SetVisualScale(float64)       {}

func (c *cdnDetail) Render(rd render.Renderer, ctx style.Context) {
	if c.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*c.Group.Style())
	tok := style.TokensFor(eff)
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*c.reveal)

	// Per-dot knock-out color (matches the icon's fill behind the
	// hatch). Each dot sits on the ellipse perimeter — half of the
	// dot's area covers hatched fill and would blend away without a
	// halo. We can't use IconBase's knock-out because that operates
	// on a single bounding rect that would erase the whole ellipse.
	var koColor color.Color
	switch eff.FillStyle {
	case style.FillHatch, style.FillCrossHatch, style.FillZigzag, style.FillDots:
		if eff.FillColor != nil {
			koColor = eff.FillColor
		} else if ctx.BgColor != nil {
			koColor = ctx.BgColor
		}
	}

	// Four edge nodes around the perimeter — the "geographic
	// distribution points." Larger filled dots than the original so
	// they read clearly in sketchy modes.
	r := c.w/2 - 10
	const dotR = 9.0
	haloR := dotR + 2
	if tok.Roughness >= 2 {
		haloR = dotR + 4
	}
	thetas := []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2}
	if koColor != nil {
		haloCol := style.ApplyOpacity(koColor, tok.OpacityScale*c.reveal)
		for _, theta := range thetas {
			px := c.cx + r*math.Cos(theta)
			py := c.cy + r*math.Sin(theta)
			rd.DrawPath(geometry.EllipsePath(px, py, haloR, haloR), render.PathStyle{Fill: haloCol})
		}
	}
	for _, theta := range thetas {
		px := c.cx + r*math.Cos(theta)
		py := c.cy + r*math.Sin(theta)
		rd.DrawPath(geometry.EllipsePath(px, py, dotR, dotR), render.PathStyle{Fill: col})
	}
}

// person draws a simple head + shoulders silhouette.
type person struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newPerson(seed int64, w, h float64) *person {
	return &person{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (p *person) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(p.cx, p.cy), p.w, p.h)
}
func (p *person) Children() []mobject.Mobject  { return nil }
func (p *person) Seed() int64                  { return p.seed }
func (p *person) Style() *style.Style          { return p.Group.Style() }
func (p *person) SetStyle(s style.Style)       { p.Group.SetStyle(s) }
func (p *person) Position() (float64, float64) { return p.cx, p.cy }
func (p *person) SetPosition(x, y float64)     { p.cx, p.cy = x, y }
func (p *person) SetReveal(t float64)          { p.reveal = t }
func (p *person) SetVisualScale(float64)       {}

func (p *person) Render(rd render.Renderer, ctx style.Context) {
	if p.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*p.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if p.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*p.reveal)
	}
	// Head: small circle near the top.
	headR := 28.0
	headCY := p.cy + p.h/2 - headR - 8
	if tok.Roughness == 0 {
		rd.DrawPath(geometry.EllipsePath(p.cx, headCY, headR, headR), stroke)
	} else {
		opts := style.RoughOptions(eff, tok, p.seed)
		rd.DrawPath(rough.RoughEllipse(p.cx, headCY, headR, headR, opts), stroke)
	}
	// Shoulders: trapezoidal arc below the head.
	shouldersTop := headCY - headR - 6
	shouldersBot := p.cy - p.h/2 + 18
	left := p.cx - p.w/2 + 18
	right := p.cx + p.w/2 - 18
	leftIn := p.cx - 22
	rightIn := p.cx + 22
	pts := []geometry.Point{
		{X: leftIn, Y: shouldersTop},
		{X: left, Y: shouldersBot},
		{X: right, Y: shouldersBot},
		{X: rightIn, Y: shouldersTop},
	}
	for i := 0; i < len(pts)-1; i++ {
		rd.DrawPath(makeLine(pts[i], pts[i+1], tok, eff, p.seed+int64(10+i)), stroke)
	}
}
