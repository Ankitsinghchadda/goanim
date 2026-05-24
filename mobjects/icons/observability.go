package icons

import (
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/icon"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// NewMetrics renders a rectangle with a small bar chart detail —
// "numerical monitoring." Distinct from Logs (horizontal stripes) and
// Tracing (linked nodes).
func NewMetrics(seed int64, label string) *icon.IconBase {
	const w, h = 200, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newBarChart(seed+3601, w, h))
	return ic
}

// NewLogs renders a rectangle with several short horizontal stripes —
// "log lines." Distinct from Metrics (bar chart) and BlockStorage
// (grid) by being only horizontal stripes of varying widths.
func NewLogs(seed int64, label string) *icon.IconBase {
	const w, h = 200, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newLogLines(seed+3701, w, h))
	return ic
}

// NewTracing renders a rectangle with a small horizontal "trace"
// pattern — connected nodes representing spans across time.
func NewTracing(seed int64, label string) *icon.IconBase {
	const w, h = 220, 130
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newTraceSpans(seed+3801, w, h))
	return ic
}

// NewIoTDevice renders a small sensor-like square with a wifi-signal
// arc above it — "connected device." Distinct from Client (which is
// a plain labeled rectangle) by being smaller and carrying the signal
// cue.
func NewIoTDevice(seed int64, label string) *icon.IconBase {
	const w, h = 130, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newDeviceBox(seed, w, h))
	ic.Add(newWifiArcs(seed+3901, w, h))
	return ic
}

// NewMobileClient renders a phone shape (taller than wide, with a
// small button / notch detail) — "mobile-specific endpoint." Distinct
// from Client (rectangle, wider than tall) and User (person shape).
func NewMobileClient(seed int64, label string) *icon.IconBase {
	const w, h = 110, 180
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newPhoneDetail(seed+4001, w, h))
	return ic
}

// --- barChart ---------------------------------------------------------------

// barChart draws 4 vertical bars of increasing height — bar-chart cue.
type barChart struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newBarChart(seed int64, w, h float64) *barChart {
	return &barChart{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (b *barChart) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(b.cx, b.cy), b.w*0.6, b.h*0.55)
}
func (b *barChart) Children() []mobject.Mobject  { return nil }
func (b *barChart) Seed() int64                  { return b.seed }
func (b *barChart) Style() *style.Style          { return b.Group.Style() }
func (b *barChart) SetStyle(s style.Style)       { b.Group.SetStyle(s) }
func (b *barChart) Position() (float64, float64) { return b.cx, b.cy }
func (b *barChart) SetPosition(x, y float64)     { b.cx, b.cy = x, y }
func (b *barChart) SetReveal(t float64)          { b.reveal = t }
func (b *barChart) SetVisualScale(float64)       {}

func (b *barChart) Render(rd render.Renderer, ctx style.Context) {
	if b.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*b.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.1
	if b.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*b.reveal)
	}
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*b.reveal)
	// Baseline + 4 vertical bars.
	const nBars = 4
	chartW := b.w * 0.55
	chartH := b.h * 0.5
	baseY := b.cy - chartH/2
	left := b.cx - chartW/2
	barW := chartW / float64(nBars*2)
	heights := []float64{0.3, 0.55, 0.4, 0.85}
	for i := 0; i < nBars; i++ {
		x := left + float64(i)*(chartW/float64(nBars)) + barW/2
		bh := chartH * heights[i]
		rd.DrawPath(geometry.RectanglePath(x, baseY, barW, bh, 0), render.PathStyle{Fill: col})
	}
	// Baseline.
	rd.DrawPath(makeLine(geometry.Pt(left-2, baseY), geometry.Pt(left+chartW+2, baseY), tok, eff, b.seed), stroke)
}

// --- logLines --------------------------------------------------------------

// logLines draws 4 horizontal stripes of varying widths — log-line
// cue. Differs from BlockGrid by being JUST horizontal lines (no
// vertical dividers) and from Queue dividers by being short lines
// not spanning the full width.
type logLines struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newLogLines(seed int64, w, h float64) *logLines {
	return &logLines{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (l *logLines) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(l.cx, l.cy), l.w*0.7, l.h*0.6)
}
func (l *logLines) Children() []mobject.Mobject  { return nil }
func (l *logLines) Seed() int64                  { return l.seed }
func (l *logLines) Style() *style.Style          { return l.Group.Style() }
func (l *logLines) SetStyle(s style.Style)       { l.Group.SetStyle(s) }
func (l *logLines) Position() (float64, float64) { return l.cx, l.cy }
func (l *logLines) SetPosition(x, y float64)     { l.cx, l.cy = x, y }
func (l *logLines) SetReveal(t float64)          { l.reveal = t }
func (l *logLines) SetVisualScale(float64)       {}

func (l *logLines) Render(rd render.Renderer, ctx style.Context) {
	if l.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*l.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.4
	if l.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*l.reveal)
	}
	const nLines = 4
	areaW := l.w * 0.6
	areaH := l.h * 0.5
	left := l.cx - areaW/2
	step := areaH / float64(nLines-1)
	widths := []float64{0.9, 0.55, 0.8, 0.4}
	for i := 0; i < nLines; i++ {
		y := l.cy + areaH/2 - float64(i)*step
		w := areaW * widths[i]
		p1 := geometry.Pt(left, y)
		p2 := geometry.Pt(left+w, y)
		rd.DrawPath(makeLine(p1, p2, tok, eff, l.seed+int64(i)), stroke)
	}
}

// --- traceSpans ------------------------------------------------------------

// traceSpans draws a 3-segment horizontal pipeline with dots at each
// hop — distributed trace cue.
type traceSpans struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newTraceSpans(seed int64, w, h float64) *traceSpans {
	return &traceSpans{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (t *traceSpans) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(t.cx, t.cy), t.w*0.75, t.h*0.5)
}
func (t *traceSpans) Children() []mobject.Mobject  { return nil }
func (t *traceSpans) Seed() int64                  { return t.seed }
func (t *traceSpans) Style() *style.Style          { return t.Group.Style() }
func (t *traceSpans) SetStyle(s style.Style)       { t.Group.SetStyle(s) }
func (t *traceSpans) Position() (float64, float64) { return t.cx, t.cy }
func (t *traceSpans) SetPosition(x, y float64)     { t.cx, t.cy = x, y }
func (t *traceSpans) SetReveal(r float64)          { t.reveal = r }
func (t *traceSpans) SetVisualScale(float64)       {}

func (t *traceSpans) Render(rd render.Renderer, ctx style.Context) {
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
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*t.reveal)
	// 4 nodes on a horizontal line, connected by short line segments.
	nodes := []geometry.Point{
		{X: t.cx - 60, Y: t.cy + 10},
		{X: t.cx - 20, Y: t.cy + 10},
		{X: t.cx + 20, Y: t.cy + 10},
		{X: t.cx + 60, Y: t.cy + 10},
	}
	for i := 0; i < len(nodes)-1; i++ {
		rd.DrawPath(makeLine(nodes[i], nodes[i+1], tok, eff, t.seed+int64(i)), stroke)
	}
	const dotR = 5.0
	for _, n := range nodes {
		rd.DrawPath(geometry.EllipsePath(n.X, n.Y, dotR, dotR), render.PathStyle{Fill: col})
	}
	// Three short span-bars below the line, of varying widths.
	spans := []struct{ x1, x2, y float64 }{
		{t.cx - 60, t.cx - 22, t.cy - 18},
		{t.cx - 14, t.cx + 18, t.cy - 28},
		{t.cx + 22, t.cx + 60, t.cy - 22},
	}
	stroke2 := stroke
	stroke2.StrokeWidth *= 1.6
	for i, sp := range spans {
		rd.DrawPath(makeLine(geometry.Pt(sp.x1, sp.y), geometry.Pt(sp.x2, sp.y), tok, eff, t.seed+int64(100+i)), stroke2)
	}
}

// --- deviceBox + wifiArcs --------------------------------------------------

// deviceBox is a small filled rectangle with a small antenna dot on top.
type deviceBox struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	scale  float64
	reveal float64
}

func newDeviceBox(seed int64, w, h float64) *deviceBox {
	return &deviceBox{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, scale: 1, reveal: 1}
}

func (d *deviceBox) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(d.cx, d.cy-d.h*0.15), d.w*0.7, d.h*0.5)
}
func (d *deviceBox) Children() []mobject.Mobject  { return nil }
func (d *deviceBox) Seed() int64                  { return d.seed }
func (d *deviceBox) Style() *style.Style          { return d.Group.Style() }
func (d *deviceBox) SetStyle(s style.Style)       { d.Group.SetStyle(s) }
func (d *deviceBox) Position() (float64, float64) { return d.cx, d.cy }
func (d *deviceBox) SetPosition(x, y float64)     { d.cx, d.cy = x, y }
func (d *deviceBox) SetReveal(t float64)          { d.reveal = t }
func (d *deviceBox) SetVisualScale(s float64)     { d.scale = s }

func (d *deviceBox) Render(rd render.Renderer, ctx style.Context) {
	if d.reveal <= 0 {
		return
	}
	// Delegate to a small mobject.Rectangle for the body, with the
	// rectangle SIZED smaller than the icon body so the wifi arcs above
	// have room.
	dw := d.w * 0.7
	dh := d.h * 0.5
	r := mobject.NewRectangle(d.seed, dw, dh).MoveTo(d.cx, d.cy-d.h*0.15)
	r.SetReveal(d.reveal)
	r.Render(rd, ctx)
}

// wifiArcs draws three concentric arcs above the device box —
// signal-propagation cue.
type wifiArcs struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newWifiArcs(seed int64, w, h float64) *wifiArcs {
	return &wifiArcs{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (w *wifiArcs) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(w.cx, w.cy+w.h*0.22), w.w, w.h*0.4)
}
func (w *wifiArcs) Children() []mobject.Mobject  { return nil }
func (w *wifiArcs) Seed() int64                  { return w.seed }
func (w *wifiArcs) Style() *style.Style          { return w.Group.Style() }
func (w *wifiArcs) SetStyle(s style.Style)       { w.Group.SetStyle(s) }
func (w *wifiArcs) Position() (float64, float64) { return w.cx, w.cy }
func (w *wifiArcs) SetPosition(x, y float64)     { w.cx, w.cy = x, y }
func (w *wifiArcs) SetReveal(t float64)          { w.reveal = t }
func (w *wifiArcs) SetVisualScale(float64)       {}

func (w *wifiArcs) Render(rd render.Renderer, ctx style.Context) {
	if w.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*w.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.2
	if w.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*w.reveal)
	}
	// Three concentric upward-opening arcs centered on the device top.
	// Approximate each arc with a polyline along the upper half.
	originX := w.cx
	originY := w.cy + w.h*0.1
	radii := []float64{16, 26, 36}
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*w.reveal)
	for ai, r := range radii {
		const samples = 14
		var prev geometry.Point
		for i := 0; i <= samples; i++ {
			// Theta sweeps from 135° (left-up) to 45° (right-up).
			theta := math.Pi*3/4 - float64(i)*(math.Pi/2)/samples
			px := originX + r*math.Cos(theta)
			py := originY + r*math.Sin(theta)
			cur := geometry.Pt(px, py)
			if i > 0 {
				rd.DrawPath(geometry.LinePath(prev, cur), render.PathStyle{
					Stroke: col, StrokeWidth: stroke.StrokeWidth, StrokeCap: render.CapRound,
				})
			}
			prev = cur
			_ = ai
		}
	}
}

// --- phoneDetail -----------------------------------------------------------

// phoneDetail draws a small horizontal slot near the top (speaker) and
// a small circle near the bottom (home button) — the universal phone
// cues that distinguish it from a plain rectangle.
type phoneDetail struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newPhoneDetail(seed int64, w, h float64) *phoneDetail {
	return &phoneDetail{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (p *phoneDetail) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(p.cx, p.cy), p.w*0.65, p.h*0.85)
}
func (p *phoneDetail) Children() []mobject.Mobject  { return nil }
func (p *phoneDetail) Seed() int64                  { return p.seed }
func (p *phoneDetail) Style() *style.Style          { return p.Group.Style() }
func (p *phoneDetail) SetStyle(s style.Style)       { p.Group.SetStyle(s) }
func (p *phoneDetail) Position() (float64, float64) { return p.cx, p.cy }
func (p *phoneDetail) SetPosition(x, y float64)     { p.cx, p.cy = x, y }
func (p *phoneDetail) SetReveal(t float64)          { p.reveal = t }
func (p *phoneDetail) SetVisualScale(float64)       {}

func (p *phoneDetail) Render(rd render.Renderer, ctx style.Context) {
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
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*p.reveal)
	// Speaker slot near top.
	slotY := p.cy + p.h/2 - 18
	rd.DrawPath(makeLine(geometry.Pt(p.cx-14, slotY), geometry.Pt(p.cx+14, slotY), tok, eff, p.seed), stroke)
	// Home button (filled circle near bottom).
	btnY := p.cy - p.h/2 + 18
	rd.DrawPath(geometry.EllipsePath(p.cx, btnY, 6, 6), render.PathStyle{Fill: col})
	// Inner screen rectangle outline — short of the edges, so it reads
	// "this is a phone with screen and chrome."
	const margin = 10.0
	screenTop := p.cy + p.h/2 - 30
	screenBot := p.cy - p.h/2 + 30
	screenLeft := p.cx - p.w/2 + margin
	screenRight := p.cx + p.w/2 - margin
	rd.DrawPath(makeLine(geometry.Pt(screenLeft, screenTop), geometry.Pt(screenRight, screenTop), tok, eff, p.seed+1), stroke)
	rd.DrawPath(makeLine(geometry.Pt(screenRight, screenTop), geometry.Pt(screenRight, screenBot), tok, eff, p.seed+2), stroke)
	rd.DrawPath(makeLine(geometry.Pt(screenRight, screenBot), geometry.Pt(screenLeft, screenBot), tok, eff, p.seed+3), stroke)
	rd.DrawPath(makeLine(geometry.Pt(screenLeft, screenBot), geometry.Pt(screenLeft, screenTop), tok, eff, p.seed+4), stroke)
}
