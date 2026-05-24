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

// NewDataWarehouse renders an EXTRA-TALL cylinder with horizontal
// layer-divider lines suggesting columnar/analytical storage. Distinct
// from RelationalDB by being noticeably larger and having internal
// layer lines instead of a small grid.
func NewDataWarehouse(seed int64, label string) *icon.IconBase {
	const w, h = 240, 240
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newCylinder(seed, w, h, 30))
	ic.AddDetail(newWarehouseLayers(seed+2501, w, h, 4))
	return ic
}

// NewSearchIndex renders a cylinder with a magnifying-glass detail
// in front — Elasticsearch-style "indexed for search."
func NewSearchIndex(seed int64, label string) *icon.IconBase {
	const w, h = 220, 200
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newCylinder(seed, w, h, 28))
	ic.AddDetail(newMagnifier(seed+2601, w, h))
	return ic
}

// NewTimeSeriesDB renders a cylinder with a small sparkline running
// across its face — "data ordered by time." Sparkline is a distinct
// gesture compared to grids or graphs.
func NewTimeSeriesDB(seed int64, label string) *icon.IconBase {
	const w, h = 220, 200
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newCylinder(seed, w, h, 28))
	ic.AddDetail(newSparkline(seed+2701, w, h))
	return ic
}

// NewGraphDB renders a cylinder with a small node-edge graph (3-4
// dots connected by lines) on its face — graph-structured data.
func NewGraphDB(seed int64, label string) *icon.IconBase {
	const w, h = 220, 200
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newCylinder(seed, w, h, 28))
	ic.AddDetail(newGraphNodes(seed+2801, w, h))
	return ic
}

// --- warehouseLayers --------------------------------------------------------

// warehouseLayers draws N horizontal divider lines across the cylinder
// body, giving a multi-layered / columnar look.
type warehouseLayers struct {
	*mobject.Group
	seed   int64
	w, h   float64
	count  int
	cx, cy float64
	reveal float64
}

func newWarehouseLayers(seed int64, w, h float64, count int) *warehouseLayers {
	return &warehouseLayers{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, count: count, reveal: 1}
}

func (w *warehouseLayers) Bounds() geometry.Rect {
	// Restrict to the central front-face region of the cylinder.
	return geometry.RectFromCenter(geometry.Pt(w.cx, w.cy), w.w*0.85, w.h*0.55)
}
func (w *warehouseLayers) Children() []mobject.Mobject  { return nil }
func (w *warehouseLayers) Seed() int64                  { return w.seed }
func (w *warehouseLayers) Style() *style.Style          { return w.Group.Style() }
func (w *warehouseLayers) SetStyle(s style.Style)       { w.Group.SetStyle(s) }
func (w *warehouseLayers) Position() (float64, float64) { return w.cx, w.cy }
func (w *warehouseLayers) SetPosition(x, y float64)     { w.cx, w.cy = x, y }
func (w *warehouseLayers) SetReveal(t float64)          { w.reveal = t }
func (w *warehouseLayers) SetVisualScale(float64)       {}

func (w *warehouseLayers) Render(rd render.Renderer, ctx style.Context) {
	if w.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*w.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 0.9
	if w.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*w.reveal)
	}
	// N evenly-spaced horizontal lines across the cylinder's front face.
	lineW := w.w * 0.7
	left := w.cx - lineW/2
	right := w.cx + lineW/2
	span := w.h * 0.45
	step := span / float64(w.count+1)
	for i := 1; i <= w.count; i++ {
		y := w.cy + span/2 - float64(i)*step
		rd.DrawPath(makeLine(geometry.Pt(left, y), geometry.Pt(right, y), tok, eff, w.seed+int64(i)), stroke)
	}
}

// --- magnifier --------------------------------------------------------------

// magnifier draws a circle with a short diagonal handle —
// search-index visual cue.
type magnifier struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newMagnifier(seed int64, w, h float64) *magnifier {
	return &magnifier{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (m *magnifier) Bounds() geometry.Rect {
	r := 22.0
	return geometry.RectFromCenter(geometry.Pt(m.cx, m.cy), 2*r+24, 2*r+24)
}
func (m *magnifier) Children() []mobject.Mobject  { return nil }
func (m *magnifier) Seed() int64                  { return m.seed }
func (m *magnifier) Style() *style.Style          { return m.Group.Style() }
func (m *magnifier) SetStyle(s style.Style)       { m.Group.SetStyle(s) }
func (m *magnifier) Position() (float64, float64) { return m.cx, m.cy }
func (m *magnifier) SetPosition(x, y float64)     { m.cx, m.cy = x, y }
func (m *magnifier) SetReveal(t float64)          { m.reveal = t }
func (m *magnifier) SetVisualScale(float64)       {}

func (m *magnifier) Render(rd render.Renderer, ctx style.Context) {
	if m.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*m.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.1
	if m.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*m.reveal)
	}
	// Lens circle slightly above center; handle extends to lower-right.
	const r = 22.0
	lx := m.cx - 5
	ly := m.cy + 8
	var lens *geometry.Path
	if tok.Roughness == 0 {
		lens = geometry.EllipsePath(lx, ly, r, r)
	} else {
		opts := style.RoughOptions(eff, tok, m.seed)
		lens = rough.RoughEllipse(lx, ly, r, r, opts)
	}
	rd.DrawPath(lens, stroke)
	// Handle: from edge of lens at 45° down-right.
	cos45 := math.Sqrt2 / 2
	from := geometry.Pt(lx+r*cos45, ly-r*cos45)
	to := geometry.Pt(from.X+20, from.Y-20)
	stroke2 := stroke
	stroke2.StrokeWidth *= 1.2
	rd.DrawPath(makeLine(from, to, tok, eff, m.seed+1), stroke2)
}

// --- sparkline --------------------------------------------------------------

// sparkline draws a small zigzagging line across the cylinder's face —
// time-series cue.
type sparkline struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newSparkline(seed int64, w, h float64) *sparkline {
	return &sparkline{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (s *sparkline) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(s.cx, s.cy), s.w*0.7, s.h*0.45)
}
func (s *sparkline) Children() []mobject.Mobject  { return nil }
func (s *sparkline) Seed() int64                  { return s.seed }
func (s *sparkline) Style() *style.Style          { return s.Group.Style() }
func (s *sparkline) SetStyle(st style.Style)      { s.Group.SetStyle(st) }
func (s *sparkline) Position() (float64, float64) { return s.cx, s.cy }
func (s *sparkline) SetPosition(x, y float64)     { s.cx, s.cy = x, y }
func (s *sparkline) SetReveal(t float64)          { s.reveal = t }
func (s *sparkline) SetVisualScale(float64)       {}

func (s *sparkline) Render(rd render.Renderer, ctx style.Context) {
	if s.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*s.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.2
	if s.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*s.reveal)
	}
	// Six points forming a sparkline pattern: low, high, mid, high,
	// low, high. Across the cylinder's front face.
	lineW := s.w * 0.65
	amp := s.h * 0.12
	left := s.cx - lineW/2
	step := lineW / 5
	ys := []float64{-0.6, 0.8, -0.2, 0.5, -0.4, 0.7}
	prev := geometry.Pt(left, s.cy+ys[0]*amp)
	for i := 1; i < len(ys); i++ {
		next := geometry.Pt(left+float64(i)*step, s.cy+ys[i]*amp)
		rd.DrawPath(makeLine(prev, next, tok, eff, s.seed+int64(i)), stroke)
		prev = next
	}
}

// --- graphNodes -------------------------------------------------------------

// graphNodes draws 4 small filled dots connected by a few lines —
// graph-structured-data cue.
type graphNodes struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newGraphNodes(seed int64, w, h float64) *graphNodes {
	return &graphNodes{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (g *graphNodes) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(g.cx, g.cy), g.w*0.6, g.h*0.5)
}
func (g *graphNodes) Children() []mobject.Mobject  { return nil }
func (g *graphNodes) Seed() int64                  { return g.seed }
func (g *graphNodes) Style() *style.Style          { return g.Group.Style() }
func (g *graphNodes) SetStyle(s style.Style)       { g.Group.SetStyle(s) }
func (g *graphNodes) Position() (float64, float64) { return g.cx, g.cy }
func (g *graphNodes) SetPosition(x, y float64)     { g.cx, g.cy = x, y }
func (g *graphNodes) SetReveal(t float64)          { g.reveal = t }
func (g *graphNodes) SetVisualScale(float64)       {}

func (g *graphNodes) Render(rd render.Renderer, ctx style.Context) {
	if g.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*g.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if g.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*g.reveal)
	}
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*g.reveal)
	// Four nodes in a kite arrangement with three connecting edges.
	nodes := []geometry.Point{
		geometry.Pt(g.cx-30, g.cy+18), // left
		geometry.Pt(g.cx+30, g.cy+18), // right
		geometry.Pt(g.cx, g.cy+40),    // top
		geometry.Pt(g.cx, g.cy-10),    // bottom-center
	}
	// Edges first, so dots sit on top.
	edges := [][2]int{{0, 2}, {1, 2}, {0, 3}, {1, 3}, {2, 3}}
	for i, e := range edges {
		rd.DrawPath(makeLine(nodes[e[0]], nodes[e[1]], tok, eff, g.seed+int64(i)), stroke)
	}
	const dotR = 5.5
	for _, n := range nodes {
		rd.DrawPath(geometry.EllipsePath(n.X, n.Y, dotR, dotR), render.PathStyle{Fill: col})
	}
}
