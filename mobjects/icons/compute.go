package icons

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/icon"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// NewCache renders a rectangle with a large centered lightning bolt —
// "fast lookup." The bolt is the dominant visual cue so the icon is
// recognizable as Cache at small sizes and in any sloppiness level.
func NewCache(seed int64, label string) *icon.IconBase {
	const w, h = 180, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newBolt(seed+301, w, h))
	return ic
}

// NewWorker is a rectangle with a gear detail — background processor.
//
// Label is positioned BELOW: the gear sits in the top-right corner
// but is large enough that a centered label would crowd it under
// dense Cartoonist hatching.
func NewWorker(seed int64, label string) *icon.IconBase {
	const w, h = 200, 130
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newGear(seed+401, w, h))
	return ic
}

// NewAPIGateway renders as a rectangle with two prominent vertical
// bars near the center — the "portal" visual that traffic passes
// through. The bars are placed symmetrically around the icon center
// and span ~60% of the height so the gateway metaphor reads even at
// small sizes and in sketchy modes.
func NewAPIGateway(seed int64, label string) *icon.IconBase {
	const w, h = 200, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newGatewayBars(seed+501, w, h))
	return ic
}

// NewFunction renders a rectangle with a large centered "λ" lambda
// symbol — for serverless / FaaS. The λ fills most of the icon so it
// reads even at small sizes and in sketchy mode.
func NewFunction(seed int64, label string) *icon.IconBase {
	const w, h = 150, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	lambda := mobject.NewText(seed+601, "λ").MoveTo(0, 0)
	lambda.SetStyle(style.Style{FontSize: style.FontXLarge})
	ic.AddDetail(lambda)
	return ic
}

// bolt draws a small lightning-bolt detail in the upper-right of an
// icon body.
type bolt struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newBolt(seed int64, w, h float64) *bolt {
	return &bolt{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

// Bounds returns a tight rectangle around the bolt silhouette, used
// by IconBase's knock-out pass.
func (b *bolt) Bounds() geometry.Rect {
	s := minF(b.w, b.h) * 0.32
	// Bolt extents from the polygon: x in [-0.35s, +0.55s], y in [-0.95s, +0.95s].
	return geometry.Rect{
		Min: geometry.Pt(b.cx-s*0.5, b.cy-s),
		Max: geometry.Pt(b.cx+s*0.6, b.cy+s),
	}
}
func (b *bolt) Children() []mobject.Mobject  { return nil }
func (b *bolt) Seed() int64                  { return b.seed }
func (b *bolt) Style() *style.Style          { return b.Group.Style() }
func (b *bolt) SetStyle(s style.Style)       { b.Group.SetStyle(s) }
func (b *bolt) Position() (float64, float64) { return b.cx, b.cy }
func (b *bolt) SetPosition(x, y float64)     { b.cx, b.cy = x, y }
func (b *bolt) SetReveal(t float64)          { b.reveal = t }
func (b *bolt) SetVisualScale(float64)       {}

func (b *bolt) Render(rd render.Renderer, ctx style.Context) {
	if b.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*b.Group.Style())
	tok := style.TokensFor(eff)
	fillColor := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*b.reveal)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.4
	if b.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*b.reveal)
	}

	// Big centered lightning bolt — fills ~60% of the icon. Designed to
	// be the dominant visual cue, not a corner ornament.
	cx, cy := b.cx, b.cy
	s := minF(b.w, b.h) * 0.32 // bolt half-height
	// 6-point standard lightning silhouette.
	pts := []geometry.Point{
		{X: cx - s*0.35, Y: cy - s*0.20},
		{X: cx + s*0.20, Y: cy + s*0.95},
		{X: cx - s*0.05, Y: cy + s*0.10},
		{X: cx + s*0.55, Y: cy + s*0.10},
		{X: cx - s*0.05, Y: cy - s*0.95},
		{X: cx + s*0.15, Y: cy - s*0.20},
	}
	if tok.Roughness == 0 {
		p := geometry.NewPath()
		p.MoveTo(pts[0].X, pts[0].Y)
		for i := 1; i < len(pts); i++ {
			p.LineTo(pts[i].X, pts[i].Y)
		}
		p.Close()
		rd.DrawPath(p, render.PathStyle{Fill: fillColor})
	} else {
		opts := style.RoughOptions(eff, tok, b.seed)
		opts.Roughness = tok.Roughness * 0.7
		opts.StrokeWidth = stroke.StrokeWidth
		opts.PreserveVertices = true
		rd.DrawPath(rough.RoughPolygon(pts, opts), stroke)
	}
}

func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// gear draws a small gear (circle with notches) detail in the corner.
type gear struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newGear(seed int64, w, h float64) *gear {
	return &gear{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

// Bounds returns a tight rectangle around the gear (outer circle +
// spikes), used by IconBase's knock-out pass.
func (g *gear) Bounds() geometry.Rect {
	gx := g.cx + g.w/2 - 28
	gy := g.cy + g.h/2 - 28
	// 11px outer radius + 6px spike tip = 17 from center.
	const r = 17.0
	return geometry.Rect{
		Min: geometry.Pt(gx-r, gy-r),
		Max: geometry.Pt(gx+r, gy+r),
	}
}
func (g *gear) Children() []mobject.Mobject  { return nil }
func (g *gear) Seed() int64                  { return g.seed }
func (g *gear) Style() *style.Style          { return g.Group.Style() }
func (g *gear) SetStyle(s style.Style)       { g.Group.SetStyle(s) }
func (g *gear) Position() (float64, float64) { return g.cx, g.cy }
func (g *gear) SetPosition(x, y float64)     { g.cx, g.cy = x, y }
func (g *gear) SetReveal(t float64)          { g.reveal = t }
func (g *gear) SetVisualScale(float64)       {}

func (g *gear) Render(rd render.Renderer, ctx style.Context) {
	if g.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*g.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if g.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*g.reveal)
	}
	gx := g.cx + g.w/2 - 28
	gy := g.cy + g.h/2 - 28
	// Outer circle + inner small circle (gear visual cue).
	if tok.Roughness == 0 {
		rd.DrawPath(geometry.EllipsePath(gx, gy, 11, 11), stroke)
		rd.DrawPath(geometry.EllipsePath(gx, gy, 4, 4), stroke)
	} else {
		o1 := style.RoughOptions(eff, tok, g.seed)
		o2 := style.RoughOptions(eff, tok, g.seed+1)
		rd.DrawPath(rough.RoughEllipse(gx, gy, 11, 11, o1), stroke)
		rd.DrawPath(rough.RoughEllipse(gx, gy, 4, 4, o2), stroke)
	}
	// 4 short radial spikes for the gear teeth.
	const tip = 6
	for i := 0; i < 4; i++ {
		switch i {
		case 0:
			rd.DrawPath(geometry.LinePath(geometry.Pt(gx, gy-11), geometry.Pt(gx, gy-11-tip)), stroke)
		case 1:
			rd.DrawPath(geometry.LinePath(geometry.Pt(gx+11, gy), geometry.Pt(gx+11+tip, gy)), stroke)
		case 2:
			rd.DrawPath(geometry.LinePath(geometry.Pt(gx, gy+11), geometry.Pt(gx, gy+11+tip)), stroke)
		case 3:
			rd.DrawPath(geometry.LinePath(geometry.Pt(gx-11, gy), geometry.Pt(gx-11-tip, gy)), stroke)
		}
	}
}

// gatewayBars draws two parallel vertical bars inside the icon body —
// the "doorway" or gateway visual cue.
type gatewayBars struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newGatewayBars(seed int64, w, h float64) *gatewayBars {
	return &gatewayBars{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

// Bounds returns a tight rectangle around the two vertical bars,
// used by IconBase's knock-out pass.
func (g *gatewayBars) Bounds() geometry.Rect {
	margin := g.h * 0.2
	barH := g.h - 2*margin
	gap := g.w * 0.18
	return geometry.RectFromCenter(geometry.Pt(g.cx, g.cy), 2*gap+10, barH+6)
}
func (g *gatewayBars) Children() []mobject.Mobject  { return nil }
func (g *gatewayBars) Seed() int64                  { return g.seed }
func (g *gatewayBars) Style() *style.Style          { return g.Group.Style() }
func (g *gatewayBars) SetStyle(s style.Style)       { g.Group.SetStyle(s) }
func (g *gatewayBars) Position() (float64, float64) { return g.cx, g.cy }
func (g *gatewayBars) SetPosition(x, y float64)     { g.cx, g.cy = x, y }
func (g *gatewayBars) SetReveal(t float64)          { g.reveal = t }
func (g *gatewayBars) SetVisualScale(float64)       {}

func (g *gatewayBars) Render(rd render.Renderer, ctx style.Context) {
	if g.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*g.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.2 // bolder so it reads at 80px in sketchy modes
	if g.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*g.reveal)
	}
	// "Gateway" = two bold vertical bars near the center of the icon,
	// flanking the label space. The bars span ~60% of the icon's height
	// so they read as a clear "portal" visual rather than a small
	// decoration in the corner.
	margin := g.h * 0.2
	barY1 := g.cy - g.h/2 + margin
	barY2 := g.cy + g.h/2 - margin
	gap := g.w * 0.18 // half-gap from center
	leftX := g.cx - gap
	rightX := g.cx + gap
	rd.DrawPath(makeLine(geometry.Pt(leftX, barY1), geometry.Pt(leftX, barY2), tok, eff, g.seed), stroke)
	rd.DrawPath(makeLine(geometry.Pt(rightX, barY1), geometry.Pt(rightX, barY2), tok, eff, g.seed+1), stroke)
}
