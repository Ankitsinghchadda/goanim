package netgraph

import (
	"image/color"
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Edge connects two Nodes with a straight or gently-curved line. By
// default it's undirected (no arrowhead) — call WithArrowhead(true)
// for directed edges. Edge supports a "pulse": a brighter sub-segment
// that travels along the edge from the source. Use SetPulse to drive
// the cascading-failure / signal-flow animations.
type Edge struct {
	*mobject.Group
	from, to   *Node
	style      style.Style
	reveal     float64
	curveBow   float64 // perpendicular handle length; 0 = straight
	hasHead    bool
	label      string
	labelPos   float64
	pulsePos   float64 // [0,1] center of pulse; <0 = no pulse
	pulseWidth float64 // fraction of length covered by the bright pulse
	pulseColor color.Color
}

// NewEdge constructs an undirected straight edge from a to b.
func NewEdge(seed int64, a, b *Node) *Edge {
	return &Edge{
		Group:      mobject.NewGroup(seed),
		from:       a,
		to:         b,
		reveal:     1,
		labelPos:   0.5,
		pulsePos:   -1,
		pulseWidth: 0.18,
	}
}

func (e *Edge) WithArrowhead(on bool) *Edge { e.hasHead = on; return e }
func (e *Edge) WithCurve(bow float64) *Edge { e.curveBow = bow; return e }
func (e *Edge) WithLabel(s string) *Edge    { e.label = s; return e }
func (e *Edge) WithLabelPosition(t float64) *Edge {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	e.labelPos = t
	return e
}

// SetPulse sets the center of the pulse highlight along the edge in
// [0,1]; negative values disable the pulse. Used by animations to slide
// a "signal" along the edge.
func (e *Edge) SetPulse(t float64) {
	e.pulsePos = t
}

// PulsePos returns the current pulse position (animator state).
func (e *Edge) PulsePos() float64 { return e.pulsePos }

// WithPulseColor overrides the pulse color (default: a warm red).
func (e *Edge) WithPulseColor(c color.Color) *Edge { e.pulseColor = c; return e }

func (e *Edge) WithStyle(s style.Style) *Edge { e.style = s; return e }
func (e *Edge) Style() *style.Style           { return &e.style }
func (e *Edge) SetStyle(s style.Style)        { e.style = s }
func (e *Edge) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	e.reveal = t
}
func (e *Edge) Reveal() float64        { return e.reveal }
func (e *Edge) SetVisualScale(float64) {}
func (e *Edge) Bounds() geometry.Rect {
	p1, p2 := e.endpoints()
	return geometry.RectFromPoints(p1, p2)
}

// endpoints returns the source and target attachment points on each
// node's circumference, projected toward the OTHER node's center.
func (e *Edge) endpoints() (geometry.Point, geometry.Point) {
	if e.from == nil || e.to == nil {
		return geometry.Pt(0, 0), geometry.Pt(0, 0)
	}
	fc := geometry.Pt(e.from.cx, e.from.cy)
	tc := geometry.Pt(e.to.cx, e.to.cy)
	p1 := e.from.AttachmentTowards(tc)
	p2 := e.to.AttachmentTowards(fc)
	return p1, p2
}

// shapePath returns the full geometric path of the edge (no reveal,
// no pulse) — a single straight LineTo or a single CurveTo. Used by
// Render and by pulseSegment for arc-length sampling.
func (e *Edge) shapePath() *geometry.Path {
	p1, p2 := e.endpoints()
	path := geometry.NewPath()
	path.MoveTo(p1.X, p1.Y)
	if e.curveBow == 0 {
		path.LineTo(p2.X, p2.Y)
		return path
	}
	// Perpendicular bow: shift control points to one side of the chord.
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	l := math.Hypot(dx, dy)
	if l < 1e-9 {
		path.LineTo(p2.X, p2.Y)
		return path
	}
	nx, ny := -dy/l, dx/l
	bow := e.curveBow
	c1 := geometry.Pt(p1.X+dx/3+nx*bow, p1.Y+dy/3+ny*bow)
	c2 := geometry.Pt(p1.X+2*dx/3+nx*bow, p1.Y+2*dy/3+ny*bow)
	path.CurveTo(c1.X, c1.Y, c2.X, c2.Y, p2.X, p2.Y)
	return path
}

func (e *Edge) Render(r render.Renderer, ctx style.Context) {
	if e.reveal <= 0 || e.from == nil || e.to == nil {
		return
	}
	eff := ctx.Resolve(e.style)
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeCap = render.CapRound

	path := e.shapePath()
	if tok.Roughness > 0 && e.curveBow == 0 {
		p1, p2 := e.endpoints()
		opts := style.RoughOptions(eff, tok, e.Seed()+777)
		opts.PreserveVertices = true
		opts.DisableMultiStroke = true
		path = rough.RoughLine(p1, p2, opts)
	}
	drawPath := path
	if e.reveal < 1 {
		drawPath = geometry.PathPrefix(path, geometry.PathLength(path)*e.reveal)
	}
	r.DrawPath(drawPath, stroke)

	// Pulse: a bright short sub-segment of the path centered at pulsePos.
	if e.pulsePos >= 0 && e.reveal >= 0.99 {
		total := geometry.PathLength(path)
		c := e.pulseColor
		if c == nil {
			c = color.RGBA{0xE5, 0x39, 0x35, 0xFF}
		}
		w := e.pulseWidth
		if w <= 0 {
			w = 0.18
		}
		start := math.Max(0, e.pulsePos-w/2)
		end := math.Min(1, e.pulsePos+w/2)
		if end > start {
			full := geometry.PathPrefix(path, total*end)
			// Drop the first `start*total` of arc length by sampling the
			// reversed remainder; simpler approach: just use PathPrefix
			// twice and let the difference be the visible bright slice.
			head := geometry.PathPrefix(path, total*start)
			pulse := pathDifference(full, head)
			r.DrawPath(pulse, render.PathStyle{
				Stroke:      c,
				StrokeWidth: stroke.StrokeWidth * 2.2,
				StrokeCap:   render.CapRound,
			})
		}
	}

	if e.hasHead && e.reveal >= 0.95 {
		e.renderArrowhead(r, eff, tok)
	}
	if e.label != "" && e.reveal >= 0.95 {
		e.renderLabel(r, ctx, eff, tok)
	}
}

// pathDifference returns the portion of `full` that comes after the
// arc length of `head`. Works because PathPrefix preserves MoveTo at
// the start: we splice the second-to-last point of head as the new
// MoveTo of the result. Cheap and correct for line/curve subpaths.
func pathDifference(full, head *geometry.Path) *geometry.Path {
	if len(head.Cmds) == 0 {
		return full
	}
	// Drop the prefix commands; for the last command of head, use its
	// endpoint as the MoveTo of the remainder.
	out := geometry.NewPath()
	cut := len(head.Cmds)
	if cut > len(full.Cmds) {
		cut = len(full.Cmds)
	}
	// Find the start point of the remainder = endpoint of head's last cmd.
	last := head.Cmds[len(head.Cmds)-1]
	var sx, sy float64
	switch last.Kind {
	case geometry.CmdMove, geometry.CmdLine:
		sx, sy = last.P0.X, last.P0.Y
	case geometry.CmdCurve:
		sx, sy = last.P2.X, last.P2.Y
	}
	out.MoveTo(sx, sy)
	for i := cut; i < len(full.Cmds); i++ {
		out.Cmds = append(out.Cmds, full.Cmds[i])
	}
	return out
}

func (e *Edge) renderArrowhead(r render.Renderer, eff style.Style, tok style.Tokens) {
	path := e.shapePath()
	total := geometry.PathLength(path)
	if total < 4 {
		return
	}
	// Sample tip + just-before-tip to get a tangent.
	tip := geometry.PointAlongPath(path, total)
	pre := geometry.PointAlongPath(path, total-2)
	dx := tip.X - pre.X
	dy := tip.Y - pre.Y
	l := math.Hypot(dx, dy)
	if l < 1e-9 {
		return
	}
	ux, uy := dx/l, dy/l
	headLen := math.Max(tok.StrokeWidthPx*7, 16)
	const headAng = 0.45
	cosA, sinA := math.Cos(headAng), math.Sin(headAng)
	lx := -ux*cosA - uy*sinA
	ly := -uy*cosA + ux*sinA
	rx := -ux*cosA + uy*sinA
	ry := -uy*cosA - ux*sinA
	left := geometry.Pt(tip.X+lx*headLen, tip.Y+ly*headLen)
	right := geometry.Pt(tip.X+rx*headLen, tip.Y+ry*headLen)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.DashArray = nil
	stroke.StrokeCap = render.CapRound
	r.DrawPath(geometry.LinePath(tip, left), stroke)
	r.DrawPath(geometry.LinePath(tip, right), stroke)
}

func (e *Edge) renderLabel(r render.Renderer, ctx style.Context, eff style.Style, tok style.Tokens) {
	path := e.shapePath()
	total := geometry.PathLength(path)
	pos := geometry.PointAlongPath(path, total*e.labelPos)
	face := ctx.FontFace(eff.FontFamily)
	size := tok.FontSizePx * 0.72
	// Background pill: small pad so text isn't sitting on the line.
	padX := size * 0.6
	padY := size * 0.25
	textW := float64(len(e.label)) * size * 0.55
	bg := geometry.RectanglePath(
		pos.X-textW/2-padX, pos.Y-size/2-padY,
		textW+2*padX, size+2*padY,
		6,
	)
	if ctx.SceneDefault.FillColor != nil {
		r.DrawPath(bg, render.PathStyle{Fill: ctx.SceneDefault.FillColor})
	}
	r.DrawText(e.label, pos.X, pos.Y, render.TextStyle{
		Face:     face,
		Size:     size,
		Color:    eff.StrokeColor,
		Align:    render.AlignCenter,
		Baseline: render.BaselineMiddle,
	})
}

// Children satisfies Mobject when the consumer treats Edge as a leaf.
func (e *Edge) Children() []mobject.Mobject { return nil }
