package icons

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/icon"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// NewContainer renders a rectangle with three short horizontal "stripes"
// across the top — the shipping-container metaphor for a containerized
// unit. Distinct from Server (which has rack lines on the right side):
// the stripes span the full top edge, signaling "this is one packaged
// unit" rather than "this is hardware."
func NewContainer(seed int64, label string) *icon.IconBase {
	const w, h = 220, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelInside)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newContainerStripes(seed+1101, w, h))
	return ic
}

// NewPod renders a Container shape with a dotted boundary around it —
// the Kubernetes-pod metaphor of "a scheduled unit holding a container."
// No vendor branding; just "container inside a dotted boundary."
//
// Body dimensions are the OUTER (dotted) boundary so arrows attach
// there. The inner container is drawn as a frame part (with stripes
// added as detail).
func NewPod(seed int64, label string) *icon.IconBase {
	const w, h = 240, 160
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	// The dotted outer boundary is drawn first as a frame part; the
	// inner container body sits inside.
	ic.Add(newPodBoundary(seed+1201, w, h))
	// Inner container — slightly smaller, no fill of its own (would
	// double-hatch the inside).
	ic.Add(newPodInner(seed+1202, w-50, h-40))
	ic.AddDetail(newContainerStripes(seed+1203, w-50, h-40))
	return ic
}

// NewCluster renders a dashed outer boundary containing three small
// node squares — the "managed group of compute resources" metaphor.
// Distinct from Pod (which has ONE inner container with a dotted
// outline) by the dashed line and multiple inner nodes.
func NewCluster(seed int64, label string) *icon.IconBase {
	const w, h = 300, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newClusterBoundary(seed+1301, w, h))
	ic.AddDetail(newClusterNodes(seed+1302, w, h, 3))
	return ic
}

// NewVM renders a rectangle inside a rectangle — the outer being the
// host machine, the inner being the virtual guest. The inner is offset
// slightly toward the top-left so the host context reads clearly.
// Distinct from Container (no shipping stripes) and Server (no rack
// lines) — just the "machine running another machine" cue.
func NewVM(seed int64, label string) *icon.IconBase {
	const w, h = 220, 150
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newVMInnerRect(seed+1401, w, h))
	return ic
}

// NewEdgeFunction renders the same λ centerpiece as Function but with
// short radiating arcs around the icon, conveying "executes at the
// edge / near the user." The arcs are the disambiguating cue against
// the plain Function icon.
//
// edgeWaves is a frame part (not a detail) because its strokes sit
// OUTSIDE the icon body — using IconBase's bounding-rect knock-out
// would erase the entire icon's hatch. Per-arc halos handled in
// edgeWaves.Render.
func NewEdgeFunction(seed int64, label string) *icon.IconBase {
	const w, h = 160, 150
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	lambda := mobject.NewText(seed+1501, "λ").MoveTo(0, 0)
	lambda.SetStyle(style.Style{FontSize: style.FontXLarge})
	ic.AddDetail(lambda)
	ic.Add(newEdgeWaves(seed+1502, w, h))
	return ic
}

// --- containerStripes -------------------------------------------------------

// containerStripes draws three short horizontal lines near the top
// edge of an icon body — the shipping-container metaphor.
type containerStripes struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newContainerStripes(seed int64, w, h float64) *containerStripes {
	return &containerStripes{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

// Bounds: a tight rect around the three stripes near the icon top.
// The icon body is h tall; stripes occupy the top ~15% with small
// padding on either side.
func (c *containerStripes) Bounds() geometry.Rect {
	stripeW := c.w * 0.6
	topY := c.cy + c.h/2 - 14
	botY := topY - 22
	return geometry.Rect{
		Min: geometry.Pt(c.cx-stripeW/2-3, botY-3),
		Max: geometry.Pt(c.cx+stripeW/2+3, topY+3),
	}
}
func (c *containerStripes) Children() []mobject.Mobject  { return nil }
func (c *containerStripes) Seed() int64                  { return c.seed }
func (c *containerStripes) Style() *style.Style          { return c.Group.Style() }
func (c *containerStripes) SetStyle(s style.Style)       { c.Group.SetStyle(s) }
func (c *containerStripes) Position() (float64, float64) { return c.cx, c.cy }
func (c *containerStripes) SetPosition(x, y float64)     { c.cx, c.cy = x, y }
func (c *containerStripes) SetReveal(t float64)          { c.reveal = t }
func (c *containerStripes) SetVisualScale(float64)       {}

func (c *containerStripes) Render(rd render.Renderer, ctx style.Context) {
	if c.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*c.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.1
	if c.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*c.reveal)
	}
	stripeW := c.w * 0.6
	topY := c.cy + c.h/2 - 14
	for i := 0; i < 3; i++ {
		y := topY - float64(i)*10
		p1 := geometry.Pt(c.cx-stripeW/2, y)
		p2 := geometry.Pt(c.cx+stripeW/2, y)
		rd.DrawPath(makeLine(p1, p2, tok, eff, c.seed+int64(i)), stroke)
	}
}

// --- podBoundary + podInner ------------------------------------------------

// podBoundary draws a DOTTED outer rectangle.
type podBoundary struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newPodBoundary(seed int64, w, h float64) *podBoundary {
	return &podBoundary{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (p *podBoundary) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(p.cx, p.cy), p.w, p.h)
}
func (p *podBoundary) Children() []mobject.Mobject  { return nil }
func (p *podBoundary) Seed() int64                  { return p.seed }
func (p *podBoundary) Style() *style.Style          { return p.Group.Style() }
func (p *podBoundary) SetStyle(s style.Style)       { p.Group.SetStyle(s) }
func (p *podBoundary) Position() (float64, float64) { return p.cx, p.cy }
func (p *podBoundary) SetPosition(x, y float64)     { p.cx, p.cy = x, y }
func (p *podBoundary) SetReveal(t float64)          { p.reveal = t }
func (p *podBoundary) SetVisualScale(float64)       {}

func (p *podBoundary) Render(rd render.Renderer, ctx style.Context) {
	if p.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*p.Group.Style())
	tok := style.TokensFor(eff)
	// Stroke only, dotted. No fill — the inner container provides the
	// filled area.
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*p.reveal)
	ps := render.PathStyle{
		Stroke:      col,
		StrokeWidth: tok.StrokeWidthPx,
		StrokeCap:   render.CapRound,
		DashArray:   []float64{tok.StrokeWidthPx * 0.6, tok.StrokeWidthPx * 2.4},
	}
	rect := geometry.RectanglePath(p.cx-p.w/2, p.cy-p.h/2, p.w, p.h, tok.CornerRadius)
	rd.DrawPath(rect, ps)
}

// podInner is the filled inner container rectangle (separate from
// mobject.NewRectangle so it doesn't double-fill the area inside the
// dotted boundary — fills the icon's resolved fill style normally).
type podInner struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newPodInner(seed int64, w, h float64) *podInner {
	return &podInner{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (p *podInner) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(p.cx, p.cy), p.w, p.h)
}
func (p *podInner) Children() []mobject.Mobject  { return nil }
func (p *podInner) Seed() int64                  { return p.seed }
func (p *podInner) Style() *style.Style          { return p.Group.Style() }
func (p *podInner) SetStyle(s style.Style)       { p.Group.SetStyle(s) }
func (p *podInner) Position() (float64, float64) { return p.cx, p.cy }
func (p *podInner) SetPosition(x, y float64)     { p.cx, p.cy = x, y }
func (p *podInner) SetReveal(t float64)          { p.reveal = t }
func (p *podInner) SetVisualScale(float64)       {}

func (p *podInner) Render(rd render.Renderer, ctx style.Context) {
	if p.reveal <= 0 {
		return
	}
	// Delegate to mobject.Rectangle's drawing behavior by constructing
	// one on the fly with our seed and dimensions, sized to our cx/cy.
	r := mobject.NewRectangle(p.seed, p.w, p.h).MoveTo(p.cx, p.cy)
	r.SetReveal(p.reveal)
	r.Render(rd, ctx)
}

// --- clusterBoundary + clusterNodes ----------------------------------------

// clusterBoundary draws a DASHED outer rectangle with no fill.
// Distinct from Pod's dotted boundary by the longer dash pattern.
type clusterBoundary struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newClusterBoundary(seed int64, w, h float64) *clusterBoundary {
	return &clusterBoundary{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (c *clusterBoundary) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(c.cx, c.cy), c.w, c.h)
}
func (c *clusterBoundary) Children() []mobject.Mobject  { return nil }
func (c *clusterBoundary) Seed() int64                  { return c.seed }
func (c *clusterBoundary) Style() *style.Style          { return c.Group.Style() }
func (c *clusterBoundary) SetStyle(s style.Style)       { c.Group.SetStyle(s) }
func (c *clusterBoundary) Position() (float64, float64) { return c.cx, c.cy }
func (c *clusterBoundary) SetPosition(x, y float64)     { c.cx, c.cy = x, y }
func (c *clusterBoundary) SetReveal(t float64)          { c.reveal = t }
func (c *clusterBoundary) SetVisualScale(float64)       {}

func (c *clusterBoundary) Render(rd render.Renderer, ctx style.Context) {
	if c.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*c.Group.Style())
	tok := style.TokensFor(eff)
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*c.reveal)
	ps := render.PathStyle{
		Stroke:      col,
		StrokeWidth: tok.StrokeWidthPx,
		StrokeCap:   render.CapRound,
		DashArray:   []float64{tok.StrokeWidthPx * 4, tok.StrokeWidthPx * 3},
	}
	rect := geometry.RectanglePath(c.cx-c.w/2, c.cy-c.h/2, c.w, c.h, tok.CornerRadius)
	rd.DrawPath(rect, ps)
}

// clusterNodes draws N small node squares evenly spaced along a
// horizontal centerline inside the cluster boundary.
type clusterNodes struct {
	*mobject.Group
	seed   int64
	w, h   float64
	nodes  int
	cx, cy float64
	reveal float64
}

func newClusterNodes(seed int64, w, h float64, nodes int) *clusterNodes {
	return &clusterNodes{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, nodes: nodes, reveal: 1}
}

func (c *clusterNodes) Bounds() geometry.Rect {
	// Tight rect around the nodes — vertically centered, spanning ~80%
	// of the cluster width.
	const nodeH = 60
	return geometry.RectFromCenter(geometry.Pt(c.cx, c.cy), c.w*0.85, nodeH+8)
}
func (c *clusterNodes) Children() []mobject.Mobject  { return nil }
func (c *clusterNodes) Seed() int64                  { return c.seed }
func (c *clusterNodes) Style() *style.Style          { return c.Group.Style() }
func (c *clusterNodes) SetStyle(s style.Style)       { c.Group.SetStyle(s) }
func (c *clusterNodes) Position() (float64, float64) { return c.cx, c.cy }
func (c *clusterNodes) SetPosition(x, y float64)     { c.cx, c.cy = x, y }
func (c *clusterNodes) SetReveal(t float64)          { c.reveal = t }
func (c *clusterNodes) SetVisualScale(float64)       {}

func (c *clusterNodes) Render(rd render.Renderer, ctx style.Context) {
	if c.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*c.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if c.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*c.reveal)
	}
	const nodeW, nodeH = 60.0, 60.0
	spacing := (c.w*0.85 - float64(c.nodes)*nodeW) / float64(c.nodes-1)
	startX := c.cx - c.w*0.85/2
	for i := 0; i < c.nodes; i++ {
		nx := startX + float64(i)*(nodeW+spacing) + nodeW/2
		// Filled fill color rectangle so the node reads as a solid panel
		// against the cluster boundary's hatched interior.
		if eff.FillColor != nil {
			fillCol := style.ApplyOpacity(eff.FillColor, tok.OpacityScale*c.reveal)
			rd.DrawPath(geometry.RectanglePath(nx-nodeW/2, c.cy-nodeH/2, nodeW, nodeH, tok.CornerRadius/2),
				render.PathStyle{Fill: fillCol})
		}
		// Then the outline.
		var rectPath *geometry.Path
		if tok.Roughness == 0 {
			rectPath = geometry.RectanglePath(nx-nodeW/2, c.cy-nodeH/2, nodeW, nodeH, tok.CornerRadius/2)
		} else {
			opts := style.RoughOptions(eff, tok, c.seed+int64(i*17))
			opts.Roughness = tok.Roughness * 0.7
			opts.DisableMultiStroke = true
			rectPath = rough.RoughRectangle(nx-nodeW/2, c.cy-nodeH/2, nodeW, nodeH, opts)
		}
		rd.DrawPath(rectPath, stroke)
	}
}

// --- vmInnerRect ------------------------------------------------------------

// vmInnerRect draws a smaller rectangle inside the icon body, offset
// slightly toward the upper-right to convey "guest inside host."
type vmInnerRect struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newVMInnerRect(seed int64, w, h float64) *vmInnerRect {
	return &vmInnerRect{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (v *vmInnerRect) Bounds() geometry.Rect {
	innerW := v.w * 0.55
	innerH := v.h * 0.55
	off := v.w * 0.12
	return geometry.RectFromCenter(geometry.Pt(v.cx+off/2, v.cy+off/2), innerW, innerH)
}
func (v *vmInnerRect) Children() []mobject.Mobject  { return nil }
func (v *vmInnerRect) Seed() int64                  { return v.seed }
func (v *vmInnerRect) Style() *style.Style          { return v.Group.Style() }
func (v *vmInnerRect) SetStyle(s style.Style)       { v.Group.SetStyle(s) }
func (v *vmInnerRect) Position() (float64, float64) { return v.cx, v.cy }
func (v *vmInnerRect) SetPosition(x, y float64)     { v.cx, v.cy = x, y }
func (v *vmInnerRect) SetReveal(t float64)          { v.reveal = t }
func (v *vmInnerRect) SetVisualScale(float64)       {}

func (v *vmInnerRect) Render(rd render.Renderer, ctx style.Context) {
	if v.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*v.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.15
	if v.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*v.reveal)
	}
	innerW := v.w * 0.55
	innerH := v.h * 0.55
	off := v.w * 0.12
	cx := v.cx + off/2
	cy := v.cy + off/2
	// Filled inner rect in fill color so it reads against the host's
	// hatch.
	if eff.FillColor != nil {
		fillCol := style.ApplyOpacity(eff.FillColor, tok.OpacityScale*v.reveal)
		rd.DrawPath(geometry.RectanglePath(cx-innerW/2, cy-innerH/2, innerW, innerH, tok.CornerRadius),
			render.PathStyle{Fill: fillCol})
	}
	var path *geometry.Path
	if tok.Roughness == 0 {
		path = geometry.RectanglePath(cx-innerW/2, cy-innerH/2, innerW, innerH, tok.CornerRadius)
	} else {
		opts := style.RoughOptions(eff, tok, v.seed)
		opts.Roughness = tok.Roughness * 0.7
		opts.DisableMultiStroke = true
		path = rough.RoughRectangle(cx-innerW/2, cy-innerH/2, innerW, innerH, opts)
	}
	rd.DrawPath(path, stroke)
}

// --- edgeWaves --------------------------------------------------------------

// edgeWaves draws four short radiating arcs around the icon center —
// the "executes at the edge" signal-propagation cue.
type edgeWaves struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newEdgeWaves(seed int64, w, h float64) *edgeWaves {
	return &edgeWaves{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (e *edgeWaves) Bounds() geometry.Rect {
	// Waves sit at the corners of the icon body. A small halo region
	// per corner is what we want, but Bounds is a single rect so we
	// return the full corner-spanning region. Knock-out fill color
	// equals the icon's fill, so this just re-paints the same color
	// over hatch in the corner regions — visually fine.
	return geometry.RectFromCenter(geometry.Pt(e.cx, e.cy), e.w, e.h)
}
func (e *edgeWaves) Children() []mobject.Mobject  { return nil }
func (e *edgeWaves) Seed() int64                  { return e.seed }
func (e *edgeWaves) Style() *style.Style          { return e.Group.Style() }
func (e *edgeWaves) SetStyle(s style.Style)       { e.Group.SetStyle(s) }
func (e *edgeWaves) Position() (float64, float64) { return e.cx, e.cy }
func (e *edgeWaves) SetPosition(x, y float64)     { e.cx, e.cy = x, y }
func (e *edgeWaves) SetReveal(t float64)          { e.reveal = t }
func (e *edgeWaves) SetVisualScale(float64)       {}

func (e *edgeWaves) Render(rd render.Renderer, ctx style.Context) {
	if e.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*e.Group.Style())
	tok := style.TokensFor(eff)
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*e.reveal)
	ps := render.PathStyle{
		Stroke:      col,
		StrokeWidth: tok.StrokeWidthPx * 0.9,
		StrokeCap:   render.CapRound,
	}
	// Three concentric ticks fanning out from each of the icon's mid
	// edges (not corners — top, right, bottom, left). Sit OUTSIDE the
	// body, so they read on the page background; no per-tick knock-out
	// needed because they don't overlap the hatch fill.
	const tickLen = 12
	const tickStep = 8
	// edge centers (mx, my) and outward direction (dx, dy)
	edges := [][4]float64{
		{e.cx, e.cy + e.h/2, 0, 1},  // top
		{e.cx + e.w/2, e.cy, 1, 0},  // right
		{e.cx, e.cy - e.h/2, 0, -1}, // bottom
		{e.cx - e.w/2, e.cy, -1, 0}, // left
	}
	for i, edge := range edges {
		mx, my, dx, dy := edge[0], edge[1], edge[2], edge[3]
		// Perpendicular to outward direction — that's along the icon edge.
		px, py := -dy, dx
		for j := 0; j < 3; j++ {
			d := 6 + float64(j)*tickStep
			midX := mx + dx*d
			midY := my + dy*d
			a := geometry.Pt(midX+px*float64(tickLen-j*2)/2, midY+py*float64(tickLen-j*2)/2)
			b := geometry.Pt(midX-px*float64(tickLen-j*2)/2, midY-py*float64(tickLen-j*2)/2)
			rd.DrawPath(makeLine(a, b, tok, eff, e.seed+int64(i*11+j*3)), ps)
		}
	}
}
