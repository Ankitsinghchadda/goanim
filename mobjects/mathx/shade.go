package mathx

import (
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Shade fills the region between a Graph curve and the axes' x-axis
// (or yMin, whichever is in range) across [xMin, xMax]. Used to mark
// areas under a probability distribution, integrals, etc.
//
// Reveal grows the shaded region from left to right by clipping the
// x-range — useful for animating the tail of a bell curve filling in.
type Shade struct {
	*mobject.Group
	graph      *Graph
	xMin, xMax float64
	samples    int
	style      style.Style
	reveal     float64
}

// NewShade builds a fill region under g across [xMin, xMax]. The region
// is bounded above by g.fn and below by y=0 (clamped to the axes'
// y-range).
func NewShade(g *Graph, xMin, xMax float64) *Shade {
	return &Shade{
		Group:   mobject.NewGroup(0),
		graph:   g,
		xMin:    xMin,
		xMax:    xMax,
		samples: 120,
		reveal:  1,
	}
}

func (s *Shade) WithSamples(n int) *Shade { s.samples = n; return s }
func (s *Shade) Bounds() geometry.Rect    { return s.graph.axes.Bounds() }
func (s *Shade) Style() *style.Style      { return &s.style }
func (s *Shade) SetStyle(st style.Style)  { s.style = st }
func (s *Shade) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	s.reveal = t
}
func (s *Shade) Reveal() float64        { return s.reveal }
func (s *Shade) SetVisualScale(float64) {}
func (s *Shade) Seed() int64            { return 0 }

func (s *Shade) Render(r render.Renderer, ctx style.Context) {
	if s.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(s.style)
	if eff.FillColor == nil {
		return
	}
	tok := style.TokensFor(eff)
	ax := s.graph.axes
	baselineY := math.Max(ax.yMin, math.Min(ax.yMax, 0))

	xHi := s.xMin + (s.xMax-s.xMin)*s.reveal
	n := s.samples
	if n < 2 {
		n = 2
	}

	pts := []geometry.Point{}
	// Top curve: function samples clamped to the axes' y-range.
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		x := s.xMin + t*(xHi-s.xMin)
		y := s.graph.fn(x)
		if math.IsNaN(y) || math.IsInf(y, 0) {
			continue
		}
		if y > ax.yMax {
			y = ax.yMax
		}
		if y < ax.yMin {
			y = ax.yMin
		}
		pts = append(pts, ax.PointAt(x, y))
	}
	if len(pts) < 2 {
		return
	}
	// Baseline: walk back along y=0 (or yMin) from xHi to xMin.
	pts = append(pts, ax.PointAt(xHi, baselineY))
	pts = append(pts, ax.PointAt(s.xMin, baselineY))

	path := geometry.NewPath()
	path.MoveTo(pts[0].X, pts[0].Y)
	for i := 1; i < len(pts); i++ {
		path.LineTo(pts[i].X, pts[i].Y)
	}
	path.Close()

	fill := style.ApplyOpacity(eff.FillColor, tok.OpacityScale)
	r.DrawPath(path, render.PathStyle{Fill: fill})
}
