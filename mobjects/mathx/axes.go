package mathx

import (
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Axes is a 2D coordinate plane: two perpendicular number lines with
// tick marks. Used as the parent of Graph mobjects.
type Axes struct {
	*mobject.Group
	xMin, xMax, yMin, yMax float64
	xStep, yStep           float64
	width, height          float64
	cx, cy                 float64
	showGrid               bool
	showLabels             bool
	style                  style.Style
	reveal                 float64
}

// NewAxes constructs a 2D coordinate plane covering the given ranges.
// Pixel dimensions default to 800×500; override with WithSize.
func NewAxes(xMin, xMax, yMin, yMax float64) *Axes {
	return &Axes{
		Group:      mobject.NewGroup(0),
		xMin:       xMin,
		xMax:       xMax,
		yMin:       yMin,
		yMax:       yMax,
		xStep:      1,
		yStep:      1,
		width:      800,
		height:     500,
		showLabels: true,
		reveal:     1,
	}
}

func (a *Axes) WithSize(w, h float64) *Axes    { a.width, a.height = w, h; return a }
func (a *Axes) WithSteps(xs, ys float64) *Axes { a.xStep, a.yStep = xs, ys; return a }
func (a *Axes) WithGrid(show bool) *Axes       { a.showGrid = show; return a }
func (a *Axes) WithLabels(show bool) *Axes     { a.showLabels = show; return a }
func (a *Axes) MoveTo(x, y float64) *Axes      { a.cx, a.cy = x, y; return a }
func (a *Axes) SetPosition(x, y float64)       { a.cx, a.cy = x, y }
func (a *Axes) Position() (float64, float64)   { return a.cx, a.cy }
func (a *Axes) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(a.cx, a.cy), a.width, a.height)
}
func (a *Axes) VisualBounds() geometry.Rect { return a.Bounds() }
func (a *Axes) Style() *style.Style         { return &a.style }
func (a *Axes) SetStyle(s style.Style)      { a.style = s }
func (a *Axes) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	a.reveal = t
}
func (a *Axes) Reveal() float64        { return a.reveal }
func (a *Axes) SetVisualScale(float64) {}

// PointAt converts a (math-space x, y) to a scene coordinate.
func (a *Axes) PointAt(x, y float64) geometry.Point {
	xRange := a.xMax - a.xMin
	yRange := a.yMax - a.yMin
	tx := (x - a.xMin) / xRange
	ty := (y - a.yMin) / yRange
	px := a.cx - a.width/2 + tx*a.width
	py := a.cy - a.height/2 + ty*a.height
	return geometry.Pt(px, py)
}

// Plot constructs a Graph of f over the axes. Default range is the
// axes' xMin..xMax; override with Graph.WithRange.
func (a *Axes) Plot(fn func(float64) float64) *Graph {
	return &Graph{
		axes:    a,
		fn:      fn,
		xMin:    a.xMin,
		xMax:    a.xMax,
		samples: 200,
		reveal:  1,
	}
}

func (a *Axes) Render(r render.Renderer, ctx style.Context) {
	if a.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(a.style)
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if a.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*a.reveal)
	}

	// Pick where the axes cross. If 0 is in range, axes go through the
	// origin; otherwise they hug the minimum row/column so the chart
	// still has visible reference lines on its edge.
	xAxisY := math.Max(a.yMin, math.Min(a.yMax, 0))
	yAxisX := math.Max(a.xMin, math.Min(a.xMax, 0))

	// Grid renders first so it sits behind the axes and ticks.
	if a.showGrid {
		a.renderGrid(r, eff, tok)
	}

	xLeft := a.PointAt(a.xMin, xAxisY)
	xRight := a.PointAt(a.xMax, xAxisY)
	yTop := a.PointAt(yAxisX, a.yMax)
	yBot := a.PointAt(yAxisX, a.yMin)

	drawLine := func(p1, p2 geometry.Point, seed int64) {
		var path *geometry.Path
		if tok.Roughness == 0 {
			path = geometry.LinePath(p1, p2)
		} else {
			opts := style.RoughOptions(eff, tok, seed)
			path = rough.RoughLine(p1, p2, opts)
		}
		if a.reveal < 1 {
			path = geometry.PathPrefix(path, geometry.PathLength(path)*a.reveal)
		}
		r.DrawPath(path, stroke)
	}
	drawLine(xLeft, xRight, 1001)
	drawLine(yBot, yTop, 1002)

	// Ticks fade with reveal rather than popping at a threshold, so the
	// draw-on animation reads as one continuous gesture.
	tickOpacity := tok.OpacityScale * smoothstep(0.4, 0.95, a.reveal)
	if tickOpacity <= 0 {
		return
	}
	tickStroke := stroke
	tickStroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tickOpacity)
	tickH := 10.0
	face := ctx.FontFace(eff.FontFamily)
	labelSize := tok.FontSizePx * 0.55
	labelColor := style.ApplyOpacity(eff.StrokeColor, tickOpacity)

	tickSeed := int64(2001)
	drawTick := func(p1, p2 geometry.Point) {
		var path *geometry.Path
		if tok.Roughness == 0 {
			path = geometry.LinePath(p1, p2)
		} else {
			opts := style.RoughOptions(eff, tok, tickSeed)
			opts.DisableMultiStroke = true // ticks are tiny; one pass is plenty
			path = rough.RoughLine(p1, p2, opts)
		}
		tickSeed++
		r.DrawPath(path, tickStroke)
	}

	for v := math.Ceil(a.xMin/a.xStep) * a.xStep; v <= a.xMax+1e-6; v += a.xStep {
		pt := a.PointAt(v, xAxisY)
		drawTick(
			geometry.Pt(pt.X, pt.Y-tickH/2),
			geometry.Pt(pt.X, pt.Y+tickH/2),
		)
		if a.showLabels && !nearZero(v) {
			r.DrawText(numberFormat(v), pt.X, pt.Y-tickH-labelSize*0.6, render.TextStyle{
				Face:     face,
				Size:     labelSize,
				Color:    labelColor,
				Align:    render.AlignCenter,
				Baseline: render.BaselineMiddle,
			})
		}
	}
	for v := math.Ceil(a.yMin/a.yStep) * a.yStep; v <= a.yMax+1e-6; v += a.yStep {
		pt := a.PointAt(yAxisX, v)
		drawTick(
			geometry.Pt(pt.X-tickH/2, pt.Y),
			geometry.Pt(pt.X+tickH/2, pt.Y),
		)
		if a.showLabels && !nearZero(v) {
			r.DrawText(numberFormat(v), pt.X-tickH-labelSize*0.5, pt.Y, render.TextStyle{
				Face:     face,
				Size:     labelSize,
				Color:    labelColor,
				Align:    render.AlignRight,
				Baseline: render.BaselineMiddle,
			})
		}
	}
}

// renderGrid draws faint dashed lines at every step on both axes. Drawn
// before the axis lines so it sits visually behind them. Kept crisp even
// when the chart is rough — wobbly grid lines fight with the curve for
// attention.
func (a *Axes) renderGrid(r render.Renderer, eff style.Style, tok style.Tokens) {
	gridOpacity := tok.OpacityScale * smoothstep(0.2, 0.9, a.reveal) * 0.25
	if gridOpacity <= 0 {
		return
	}
	gridStyle := render.PathStyle{
		Stroke:      style.ApplyOpacity(eff.StrokeColor, gridOpacity),
		StrokeWidth: math.Max(1, tok.StrokeWidthPx*0.5),
		StrokeCap:   render.CapButt,
		DashArray:   []float64{4, 6},
	}
	for v := math.Ceil(a.xMin/a.xStep) * a.xStep; v <= a.xMax+1e-6; v += a.xStep {
		if nearZero(v) {
			continue // origin handled by main axis
		}
		top := a.PointAt(v, a.yMax)
		bot := a.PointAt(v, a.yMin)
		r.DrawPath(geometry.LinePath(top, bot), gridStyle)
	}
	for v := math.Ceil(a.yMin/a.yStep) * a.yStep; v <= a.yMax+1e-6; v += a.yStep {
		if nearZero(v) {
			continue
		}
		left := a.PointAt(a.xMin, v)
		right := a.PointAt(a.xMax, v)
		r.DrawPath(geometry.LinePath(left, right), gridStyle)
	}
}

func smoothstep(a, b, t float64) float64 {
	if t <= a {
		return 0
	}
	if t >= b {
		return 1
	}
	x := (t - a) / (b - a)
	return x * x * (3 - 2*x)
}

func nearZero(v float64) bool { return math.Abs(v) < 1e-9 }
