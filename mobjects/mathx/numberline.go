package mathx

import (
	"fmt"
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// NumberLine renders a horizontal axis from min to max with tick
// marks at every `step` value and optional numeric labels under
// selected ticks.
type NumberLine struct {
	*mobject.Group
	min, max float64
	step     float64
	length   float64 // pixel length of the rendered line
	labels   map[float64]string
	showAll  bool
	cx, cy   float64
	style    style.Style
	reveal   float64
}

// NewNumberLine constructs a number line from min to max with step=1.
func NewNumberLine(min, max float64) *NumberLine {
	return &NumberLine{
		Group:   mobject.NewGroup(0),
		min:     min,
		max:     max,
		step:    1,
		length:  800,
		labels:  map[float64]string{},
		showAll: true,
		reveal:  1,
	}
}

func (n *NumberLine) WithStep(s float64) *NumberLine    { n.step = s; return n }
func (n *NumberLine) WithLength(px float64) *NumberLine { n.length = px; return n }
func (n *NumberLine) MoveTo(x, y float64) *NumberLine   { n.cx, n.cy = x, y; return n }
func (n *NumberLine) SetPosition(x, y float64)          { n.cx, n.cy = x, y }
func (n *NumberLine) Position() (float64, float64)      { return n.cx, n.cy }
func (n *NumberLine) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(n.cx, n.cy), n.length, 80)
}
func (n *NumberLine) VisualBounds() geometry.Rect { return n.Bounds() }
func (n *NumberLine) Style() *style.Style         { return &n.style }
func (n *NumberLine) SetStyle(s style.Style)      { n.style = s }
func (n *NumberLine) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	n.reveal = t
}
func (n *NumberLine) Reveal() float64        { return n.reveal }
func (n *NumberLine) SetVisualScale(float64) {}

// WithLabels picks specific values to receive numeric labels under
// their tick. If empty, every step gets a label.
func (n *NumberLine) WithLabels(values ...float64) *NumberLine {
	n.showAll = false
	n.labels = map[float64]string{}
	for _, v := range values {
		n.labels[v] = numberFormat(v)
	}
	return n
}

// PointAt converts a value on the number line to a scene coordinate.
func (n *NumberLine) PointAt(value float64) geometry.Point {
	t := (value - n.min) / (n.max - n.min)
	x := n.cx - n.length/2 + t*n.length
	return geometry.Pt(x, n.cy)
}

// AddPoint marks a value with a filled dot and a label. Returns the
// dot's center so the caller can attach further elements.
func (n *NumberLine) AddPoint(value float64, labelText string) geometry.Point {
	return n.PointAt(value)
}

func (n *NumberLine) Render(r render.Renderer, ctx style.Context) {
	if n.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(n.style)
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if n.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*n.reveal)
	}

	// Main line.
	left := geometry.Pt(n.cx-n.length/2, n.cy)
	right := geometry.Pt(n.cx+n.length/2, n.cy)
	var line *geometry.Path
	if tok.Roughness == 0 {
		line = geometry.LinePath(left, right)
	} else {
		opts := style.RoughOptions(eff, tok, 0)
		line = rough.RoughLine(left, right, opts)
	}
	if n.reveal < 1 {
		line = geometry.PathPrefix(line, geometry.PathLength(line)*n.reveal)
	}
	r.DrawPath(line, stroke)

	if n.reveal < 0.95 {
		return
	}

	// Ticks and labels.
	tickH := 12.0
	for v := n.min; v <= n.max+1e-6; v += n.step {
		pt := n.PointAt(v)
		// Tick mark — short vertical line.
		r.DrawPath(geometry.LinePath(
			geometry.Pt(pt.X, pt.Y-tickH/2),
			geometry.Pt(pt.X, pt.Y+tickH/2),
		), stroke)
		// Label below.
		label, ok := n.labels[v]
		if !ok && n.showAll {
			label = numberFormat(v)
		}
		if label != "" {
			face := ctx.FontFace(eff.FontFamily)
			r.DrawText(label, pt.X, pt.Y-30, render.TextStyle{
				Face:     face,
				Size:     tok.FontSizePx * 0.55,
				Color:    eff.StrokeColor,
				Align:    render.AlignCenter,
				Baseline: render.BaselineMiddle,
			})
		}
	}

	// Arrowheads on both ends.
	const ah = 14
	leftTip := left
	rightTip := right
	r.DrawPath(geometry.LinePath(leftTip, geometry.Pt(leftTip.X+ah, leftTip.Y-ah*0.4)), stroke)
	r.DrawPath(geometry.LinePath(leftTip, geometry.Pt(leftTip.X+ah, leftTip.Y+ah*0.4)), stroke)
	r.DrawPath(geometry.LinePath(rightTip, geometry.Pt(rightTip.X-ah, rightTip.Y-ah*0.4)), stroke)
	r.DrawPath(geometry.LinePath(rightTip, geometry.Pt(rightTip.X-ah, rightTip.Y+ah*0.4)), stroke)
}

// numberFormat renders a float without unnecessary decimals.
func numberFormat(v float64) string {
	if math.Abs(v-math.Round(v)) < 1e-9 {
		return fmt.Sprintf("%d", int(math.Round(v)))
	}
	return fmt.Sprintf("%g", v)
}
