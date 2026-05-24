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

// NewFirewall renders a shield-shaped outline filled with a brick
// pattern — the "traffic boundary" metaphor. Shield distinguishes it
// from LB / Gateway / ReverseProxy (all rectangles).
func NewFirewall(seed int64, label string) *icon.IconBase {
	const w, h = 180, 200
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newShield(seed, w, h))
	ic.AddDetail(newBricks(seed+3101, w, h))
	return ic
}

// NewReverseProxy renders a rectangle with two opposing horizontal
// arrows — one entering on the left, one exiting on the right —
// suggesting "traffic redirection." Distinct from LB (multi-out fan)
// and Gateway (vertical bars).
func NewReverseProxy(seed int64, label string) *icon.IconBase {
	const w, h = 220, 130
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newProxyArrows(seed+3201, w, h))
	return ic
}

// NewDNS renders a small tree shape (root with branches) suggesting
// hierarchical name resolution. Distinct from CDN (which is a circle
// with perimeter dots).
func NewDNS(seed int64, label string) *icon.IconBase {
	const w, h = 200, 160
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newDNSTree(seed+3301, w, h))
	return ic
}

// NewEventStream renders a long horizontal flow shape with several
// chevron markers indicating "flow direction" and fading edges
// (suggesting unbounded). Distinct from Queue (which has discrete
// slots and a "next-out" cue) and Broker (which has pub/sub dots).
func NewEventStream(seed int64, label string) *icon.IconBase {
	const w, h = 280, 60
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newStreamBody(seed, w, h))
	ic.AddDetail(newStreamChevrons(seed+3401, w, h))
	return ic
}

// NewPubSubTopic renders a small disc on the LEFT with multiple arrows
// fanning OUT to the right — the "one-to-many event distribution"
// cue. Distinct from LB (which is a rectangle with fan-out) by being
// a circle source rather than a rectangle.
func NewPubSubTopic(seed int64, label string) *icon.IconBase {
	const w, h = 240, 160
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newTopicDisc(seed, w, h))
	ic.AddDetail(newTopicFanout(seed+3501, w, h))
	return ic
}

// --- shield -----------------------------------------------------------------

// shield draws a heraldic shield outline (rectangle with pointed
// bottom). Renders its own hatch fill internally so the icon's body
// fill style is honored.
type shield struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	scale  float64
	reveal float64
}

func newShield(seed int64, w, h float64) *shield {
	return &shield{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, scale: 1, reveal: 1}
}

func (s *shield) shapePoints() []geometry.Point {
	hw, hh := s.w/2, s.h/2
	pts := []geometry.Point{
		{X: s.cx - hw, Y: s.cy + hh},      // top-left
		{X: s.cx + hw, Y: s.cy + hh},      // top-right
		{X: s.cx + hw, Y: s.cy - hh*0.35}, // right shoulder
		{X: s.cx, Y: s.cy - hh},           // bottom point
		{X: s.cx - hw, Y: s.cy - hh*0.35}, // left shoulder
	}
	return pts
}

func (s *shield) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(s.cx, s.cy), s.w, s.h)
}
func (s *shield) Children() []mobject.Mobject  { return nil }
func (s *shield) Seed() int64                  { return s.seed }
func (s *shield) Style() *style.Style          { return s.Group.Style() }
func (s *shield) SetStyle(st style.Style)      { s.Group.SetStyle(st) }
func (s *shield) Position() (float64, float64) { return s.cx, s.cy }
func (s *shield) SetPosition(x, y float64)     { s.cx, s.cy = x, y }
func (s *shield) SetReveal(t float64)          { s.reveal = t }
func (s *shield) SetVisualScale(v float64)     { s.scale = v }

func (s *shield) Render(rd render.Renderer, ctx style.Context) {
	if s.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*s.Group.Style())
	tok := style.TokensFor(eff)
	pts := s.shapePoints()

	// Fill (solid + hatch as appropriate).
	if eff.FillStyle != style.FillNone && eff.FillStyle != style.FillStyleUnset && eff.FillColor != nil {
		// Use the cylinder body's render approach: solid fill polygon
		// then optional hatch on top.
		polyPath := geometry.NewPath()
		polyPath.MoveTo(pts[0].X, pts[0].Y)
		for i := 1; i < len(pts); i++ {
			polyPath.LineTo(pts[i].X, pts[i].Y)
		}
		polyPath.Close()
		rd.DrawPath(polyPath, render.PathStyle{Fill: style.ApplyOpacity(eff.FillColor, tok.OpacityScale*s.reveal)})
	}

	stroke := style.PathStyleStroke(eff, tok)
	if s.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*s.reveal)
	}
	// Outline as connected line segments.
	for i := 0; i < len(pts); i++ {
		next := (i + 1) % len(pts)
		rd.DrawPath(makeLine(pts[i], pts[next], tok, eff, s.seed+int64(i)), stroke)
	}
}

// --- bricks ----------------------------------------------------------------

// bricks draws a staggered brick pattern centered on the icon.
type bricks struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newBricks(seed int64, w, h float64) *bricks {
	return &bricks{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (b *bricks) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(b.cx, b.cy), b.w*0.6, b.h*0.45)
}
func (b *bricks) Children() []mobject.Mobject  { return nil }
func (b *bricks) Seed() int64                  { return b.seed }
func (b *bricks) Style() *style.Style          { return b.Group.Style() }
func (b *bricks) SetStyle(s style.Style)       { b.Group.SetStyle(s) }
func (b *bricks) Position() (float64, float64) { return b.cx, b.cy }
func (b *bricks) SetPosition(x, y float64)     { b.cx, b.cy = x, y }
func (b *bricks) SetReveal(t float64)          { b.reveal = t }
func (b *bricks) SetVisualScale(float64)       {}

func (b *bricks) Render(rd render.Renderer, ctx style.Context) {
	if b.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*b.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 0.85
	if b.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*b.reveal)
	}
	// 3 horizontal courses, alternating stagger.
	bw := b.w * 0.55
	bh := 14.0
	left, right := b.cx-bw/2, b.cx+bw/2
	for row := 0; row < 3; row++ {
		y := b.cy + 18 - float64(row)*16
		// Horizontal line.
		rd.DrawPath(makeLine(geometry.Pt(left, y), geometry.Pt(right, y), tok, eff, b.seed+int64(row*7)), stroke)
		// Vertical mortar marks (staggered).
		nMarks := 2
		startOff := bw / float64(2*nMarks)
		if row%2 == 1 {
			startOff = 0
		}
		for i := 0; i < nMarks; i++ {
			x := left + startOff + float64(i)*bw/float64(nMarks)
			if x <= left || x >= right {
				continue
			}
			rd.DrawPath(makeLine(geometry.Pt(x, y), geometry.Pt(x, y-bh), tok, eff, b.seed+int64(row*100+i*3)), stroke)
		}
	}
}

// --- proxyArrows -----------------------------------------------------------

// proxyArrows draws two opposing horizontal arrows centered vertically
// in the icon: one pointing right (entering), one pointing left
// (returning).
type proxyArrows struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newProxyArrows(seed int64, w, h float64) *proxyArrows {
	return &proxyArrows{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (p *proxyArrows) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(p.cx, p.cy), p.w*0.75, p.h*0.55)
}
func (p *proxyArrows) Children() []mobject.Mobject  { return nil }
func (p *proxyArrows) Seed() int64                  { return p.seed }
func (p *proxyArrows) Style() *style.Style          { return p.Group.Style() }
func (p *proxyArrows) SetStyle(s style.Style)       { p.Group.SetStyle(s) }
func (p *proxyArrows) Position() (float64, float64) { return p.cx, p.cy }
func (p *proxyArrows) SetPosition(x, y float64)     { p.cx, p.cy = x, y }
func (p *proxyArrows) SetReveal(t float64)          { p.reveal = t }
func (p *proxyArrows) SetVisualScale(float64)       {}

func (p *proxyArrows) Render(rd render.Renderer, ctx style.Context) {
	if p.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*p.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.1
	if p.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*p.reveal)
	}
	lineW := p.w * 0.65
	left := p.cx - lineW/2
	right := p.cx + lineW/2
	// Top arrow: left → right
	topY := p.cy + 16
	rd.DrawPath(makeLine(geometry.Pt(left, topY), geometry.Pt(right, topY), tok, eff, p.seed), stroke)
	drawArrowhead(rd, geometry.Pt(right, topY), geometry.Pt(1, 0), 8, stroke)
	// Bottom arrow: right → left
	botY := p.cy - 16
	rd.DrawPath(makeLine(geometry.Pt(right, botY), geometry.Pt(left, botY), tok, eff, p.seed+1), stroke)
	drawArrowhead(rd, geometry.Pt(left, botY), geometry.Pt(-1, 0), 8, stroke)
}

// drawArrowhead draws an arrowhead at `tip` pointing in `dir`.
func drawArrowhead(rd render.Renderer, tip geometry.Point, dir geometry.Point, headLen float64, ps render.PathStyle) {
	a := 0.55
	cosA, sinA := math.Cos(a), math.Sin(a)
	back := geometry.Pt(-dir.X, -dir.Y)
	left := geometry.Pt(back.X*cosA-back.Y*sinA, back.X*sinA+back.Y*cosA).Scale(headLen)
	right := geometry.Pt(back.X*cosA+back.Y*sinA, -back.X*sinA+back.Y*cosA).Scale(headLen)
	rd.DrawPath(geometry.LinePath(tip, tip.Add(left)), ps)
	rd.DrawPath(geometry.LinePath(tip, tip.Add(right)), ps)
}

// --- dnsTree ---------------------------------------------------------------

// dnsTree draws a simple 3-level hierarchical tree (root → 3 nodes →
// leaves) — the "name resolution hierarchy" cue.
type dnsTree struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newDNSTree(seed int64, w, h float64) *dnsTree {
	return &dnsTree{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (d *dnsTree) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(d.cx, d.cy), d.w*0.7, d.h*0.7)
}
func (d *dnsTree) Children() []mobject.Mobject  { return nil }
func (d *dnsTree) Seed() int64                  { return d.seed }
func (d *dnsTree) Style() *style.Style          { return d.Group.Style() }
func (d *dnsTree) SetStyle(s style.Style)       { d.Group.SetStyle(s) }
func (d *dnsTree) Position() (float64, float64) { return d.cx, d.cy }
func (d *dnsTree) SetPosition(x, y float64)     { d.cx, d.cy = x, y }
func (d *dnsTree) SetReveal(t float64)          { d.reveal = t }
func (d *dnsTree) SetVisualScale(float64)       {}

func (d *dnsTree) Render(rd render.Renderer, ctx style.Context) {
	if d.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*d.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if d.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*d.reveal)
	}
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*d.reveal)
	const dotR = 5.0
	root := geometry.Pt(d.cx, d.cy+30)
	mid := []geometry.Point{
		{X: d.cx - 40, Y: d.cy},
		{X: d.cx, Y: d.cy},
		{X: d.cx + 40, Y: d.cy},
	}
	leaves := []geometry.Point{
		{X: d.cx - 55, Y: d.cy - 28},
		{X: d.cx - 25, Y: d.cy - 28},
		{X: d.cx + 25, Y: d.cy - 28},
		{X: d.cx + 55, Y: d.cy - 28},
	}
	// Edges: root → 3 mids.
	for i, m := range mid {
		rd.DrawPath(makeLine(root, m, tok, eff, d.seed+int64(i)), stroke)
	}
	// Mid[0] → leaf[0], leaf[1] ; mid[2] → leaf[2], leaf[3]. Skip mid[1].
	rd.DrawPath(makeLine(mid[0], leaves[0], tok, eff, d.seed+10), stroke)
	rd.DrawPath(makeLine(mid[0], leaves[1], tok, eff, d.seed+11), stroke)
	rd.DrawPath(makeLine(mid[2], leaves[2], tok, eff, d.seed+12), stroke)
	rd.DrawPath(makeLine(mid[2], leaves[3], tok, eff, d.seed+13), stroke)
	// Nodes.
	allNodes := append([]geometry.Point{root}, mid...)
	allNodes = append(allNodes, leaves...)
	for _, n := range allNodes {
		rd.DrawPath(geometry.EllipsePath(n.X, n.Y, dotR, dotR), render.PathStyle{Fill: col})
	}
}

// --- streamBody + streamChevrons -------------------------------------------

// streamBody is a horizontal rectangle WITHOUT a solid outline on the
// short edges — instead, the short edges fade. We achieve this by
// drawing only the two long horizontal edges; combined with the
// internal hatch (the body is added as a frame part but with no full
// outline), it visually reads as an "unbounded flow" rather than a
// bounded slot.
type streamBody struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newStreamBody(seed int64, w, h float64) *streamBody {
	return &streamBody{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (s *streamBody) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(s.cx, s.cy), s.w, s.h)
}
func (s *streamBody) Children() []mobject.Mobject  { return nil }
func (s *streamBody) Seed() int64                  { return s.seed }
func (s *streamBody) Style() *style.Style          { return s.Group.Style() }
func (s *streamBody) SetStyle(st style.Style)      { s.Group.SetStyle(st) }
func (s *streamBody) Position() (float64, float64) { return s.cx, s.cy }
func (s *streamBody) SetPosition(x, y float64)     { s.cx, s.cy = x, y }
func (s *streamBody) SetReveal(t float64)          { s.reveal = t }
func (s *streamBody) SetVisualScale(float64)       {}

func (s *streamBody) Render(rd render.Renderer, ctx style.Context) {
	if s.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*s.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if s.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*s.reveal)
	}
	// Two horizontal rails only — no short-edge endings, so it reads
	// as an open-ended flow.
	top := s.cy + s.h/2
	bot := s.cy - s.h/2
	left := s.cx - s.w/2
	right := s.cx + s.w/2
	rd.DrawPath(makeLine(geometry.Pt(left, top), geometry.Pt(right, top), tok, eff, s.seed), stroke)
	rd.DrawPath(makeLine(geometry.Pt(left, bot), geometry.Pt(right, bot), tok, eff, s.seed+1), stroke)
}

// streamChevrons draws several rightward-pointing chevrons across the
// flow body — direction-of-flow cue.
type streamChevrons struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newStreamChevrons(seed int64, w, h float64) *streamChevrons {
	return &streamChevrons{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (c *streamChevrons) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(c.cx, c.cy), c.w*0.8, c.h*0.7)
}
func (c *streamChevrons) Children() []mobject.Mobject  { return nil }
func (c *streamChevrons) Seed() int64                  { return c.seed }
func (c *streamChevrons) Style() *style.Style          { return c.Group.Style() }
func (c *streamChevrons) SetStyle(s style.Style)       { c.Group.SetStyle(s) }
func (c *streamChevrons) Position() (float64, float64) { return c.cx, c.cy }
func (c *streamChevrons) SetPosition(x, y float64)     { c.cx, c.cy = x, y }
func (c *streamChevrons) SetReveal(t float64)          { c.reveal = t }
func (c *streamChevrons) SetVisualScale(float64)       {}

func (c *streamChevrons) Render(rd render.Renderer, ctx style.Context) {
	if c.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*c.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.2
	if c.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*c.reveal)
	}
	// Per-chevron knock-out color (chevrons sit inside the stream body
	// where hatch may fight them).
	var koColor color.Color
	switch eff.FillStyle {
	case style.FillHatch, style.FillCrossHatch, style.FillZigzag, style.FillDots:
		if eff.FillColor != nil {
			koColor = eff.FillColor
		} else if ctx.BgColor != nil {
			koColor = ctx.BgColor
		}
	}
	// Four chevrons spread across the width.
	count := 4
	step := c.w * 0.7 / float64(count-1)
	startX := c.cx - c.w*0.35
	for i := 0; i < count; i++ {
		x := startX + float64(i)*step
		drawChevron(rd, x, c.cy, 8, tok, eff, stroke, koColor, c.reveal, c.seed+int64(i))
	}
}

func drawChevron(rd render.Renderer, x, y, size float64, tok style.Tokens, eff style.Style, stroke render.PathStyle, koColor color.Color, reveal float64, seed int64) {
	// Per-chevron halo so it reads through hatch fill.
	if koColor != nil {
		haloR := size + 4
		rd.DrawPath(geometry.RectanglePath(x-haloR, y-haloR, 2*haloR, 2*haloR, 0),
			render.PathStyle{Fill: style.ApplyOpacity(koColor, tok.OpacityScale*reveal)})
	}
	// Chevron: a "<" pointing right (i.e. an inverted "<", > shape).
	p1 := geometry.Pt(x-size/2, y+size)
	p2 := geometry.Pt(x+size/2, y)
	p3 := geometry.Pt(x-size/2, y-size)
	rd.DrawPath(makeLine(p1, p2, tok, eff, seed), stroke)
	rd.DrawPath(makeLine(p2, p3, tok, eff, seed+1), stroke)
}

// --- topicDisc + topicFanout -----------------------------------------------

// topicDisc draws a small filled disc on the LEFT side of the icon —
// the pub/sub source.
type topicDisc struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newTopicDisc(seed int64, w, h float64) *topicDisc {
	return &topicDisc{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (t *topicDisc) Bounds() geometry.Rect {
	// Visual bounds = full icon area, so arrows attaching to a
	// PubSubTopic via its body get sensible attachment points.
	return geometry.RectFromCenter(geometry.Pt(t.cx, t.cy), t.w, t.h)
}
func (t *topicDisc) Children() []mobject.Mobject  { return nil }
func (t *topicDisc) Seed() int64                  { return t.seed }
func (t *topicDisc) Style() *style.Style          { return t.Group.Style() }
func (t *topicDisc) SetStyle(s style.Style)       { t.Group.SetStyle(s) }
func (t *topicDisc) Position() (float64, float64) { return t.cx, t.cy }
func (t *topicDisc) SetPosition(x, y float64)     { t.cx, t.cy = x, y }
func (t *topicDisc) SetReveal(t2 float64)         { t.reveal = t2 }
func (t *topicDisc) SetVisualScale(float64)       {}

func (t *topicDisc) Render(rd render.Renderer, ctx style.Context) {
	if t.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*t.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if t.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*t.reveal)
	}
	// Disc on the LEFT side of the icon's body.
	dx := t.cx - t.w/2 + 50
	r := 40.0
	// Phase-10 Fix 3 — render the disc as an OUTLINED circle, matching
	// the rest of the catalog (outlined bodies with detail decoration).
	// The pre-Phase-10 default filled the disc with FillColor, which
	// for crisp scenes resolved to slate-100 but rendered as a
	// "solid black bug-looking" circle at small radius. Keep an
	// explicit hatch fill so sketchy still reads as filled — but
	// FillSolid in crisp now draws stroke-only. The fanout arrows
	// remain the detail element on the right.
	if eff.FillStyle == style.FillHatch || eff.FillStyle == style.FillCrossHatch {
		// Sketchy / Excalidraw presets — keep the hatched look so the
		// disc still reads as a body.
		if eff.FillColor != nil {
			rd.DrawPath(geometry.EllipsePath(dx, t.cy, r, r),
				render.PathStyle{Fill: style.ApplyOpacity(eff.FillColor, tok.OpacityScale*t.reveal*0.5)})
		}
	}
	// Outline — always drawn so the disc is recognizable as a circle.
	if tok.Roughness == 0 {
		rd.DrawPath(geometry.EllipsePath(dx, t.cy, r, r), stroke)
	} else {
		opts := style.RoughOptions(eff, tok, t.seed)
		rd.DrawPath(rough.RoughEllipse(dx, t.cy, r, r, opts), stroke)
	}
}

// topicFanout draws three arrows fanning OUT from the topic disc to
// three "subscriber" tips on the right side.
type topicFanout struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newTopicFanout(seed int64, w, h float64) *topicFanout {
	return &topicFanout{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (t *topicFanout) Bounds() geometry.Rect {
	// Tight rect around the three fan-out arrows on the right side.
	startX := t.cx - t.w/2 + 50 + 40
	tipX := t.cx + t.w/2 - 10
	return geometry.Rect{
		Min: geometry.Pt(startX-4, t.cy-50),
		Max: geometry.Pt(tipX+10, t.cy+50),
	}
}
func (t *topicFanout) Children() []mobject.Mobject  { return nil }
func (t *topicFanout) Seed() int64                  { return t.seed }
func (t *topicFanout) Style() *style.Style          { return t.Group.Style() }
func (t *topicFanout) SetStyle(s style.Style)       { t.Group.SetStyle(s) }
func (t *topicFanout) Position() (float64, float64) { return t.cx, t.cy }
func (t *topicFanout) SetPosition(x, y float64)     { t.cx, t.cy = x, y }
func (t *topicFanout) SetReveal(t2 float64)         { t.reveal = t2 }
func (t *topicFanout) SetVisualScale(float64)       {}

func (t *topicFanout) Render(rd render.Renderer, ctx style.Context) {
	if t.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*t.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.1
	if t.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*t.reveal)
	}
	startX := t.cx - t.w/2 + 50 + 40
	tipX := t.cx + t.w/2 - 10
	for i, dy := range []float64{-45, 0, 45} {
		from := geometry.Pt(startX, t.cy)
		to := geometry.Pt(tipX, t.cy+dy)
		rd.DrawPath(makeLine(from, to, tok, eff, t.seed+int64(i)), stroke)
		dir := to.Sub(from).Normalize()
		drawArrowhead(rd, to, dir, 9, stroke)
	}
}
