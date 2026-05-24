package mathx

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Graph is a plotted function over an Axes. It samples the function
// at `samples` points across [xMin, xMax], smooths them into a cubic
// Bezier path via Catmull-Rom tangents, and renders with the active
// style.
//
// Three behaviors keep the curve looking like part of the chart:
//
//   - Boundary clipping interpolates to the actual yMin/yMax crossing
//     instead of dropping the last in-range sample — so the curve meets
//     the chart edge cleanly.
//   - Discontinuities (Inf/NaN) start a new subpath, so the line breaks
//     where the function is undefined rather than connecting across the
//     jump.
//   - When the active style is rough (Roughness > 0) the curve's Bezier
//     control points are jittered perpendicular to the local segment,
//     scaled by the style's MaxJitter. The path stays continuous (no
//     per-segment MoveTo) so the curve reads as one drawn line, matching
//     the rough axes around it.
type Graph struct {
	*mobject.Group
	axes       *Axes
	fn         func(float64) float64
	xMin, xMax float64
	samples    int
	color      color.Color
	style      style.Style
	reveal     float64
	seed       int64
}

func (g *Graph) WithRange(xMin, xMax float64) *Graph { g.xMin = xMin; g.xMax = xMax; return g }
func (g *Graph) WithSamples(n int) *Graph            { g.samples = n; return g }
func (g *Graph) WithColor(c color.Color) *Graph      { g.color = c; return g }
func (g *Graph) WithSeed(s int64) *Graph             { g.seed = s; return g }
func (g *Graph) Bounds() geometry.Rect               { return g.axes.Bounds() }
func (g *Graph) Children() []mobject.Mobject         { return nil }
func (g *Graph) Seed() int64                         { return g.seed }
func (g *Graph) Style() *style.Style                 { return &g.style }
func (g *Graph) SetStyle(s style.Style)              { g.style = s }
func (g *Graph) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	g.reveal = t
}
func (g *Graph) Reveal() float64        { return g.reveal }
func (g *Graph) SetVisualScale(float64) {}

func (g *Graph) Render(r render.Renderer, ctx style.Context) {
	if g.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(g.style)
	tok := style.TokensFor(eff)
	col := g.color
	if col == nil {
		col = eff.StrokeColor
	}
	if g.reveal < 1 {
		col = style.ApplyOpacity(col, tok.OpacityScale*g.reveal)
	}
	stroke := render.PathStyle{
		Stroke:      col,
		StrokeWidth: tok.StrokeWidthPx * 1.2,
		StrokeCap:   render.CapRound,
		StrokeJoin:  render.JoinRound,
	}

	n := g.samples
	if n < 2 {
		n = 2
	}

	// Sample and split into subpaths at discontinuities or where the
	// curve leaves the axes' y-range. yClamp tracks whether the previous
	// sample was in range so we can interpolate to the boundary at
	// crossings rather than just dropping the out-of-range sample.
	subPaths := [][]geometry.Point{}
	current := []geometry.Point{}
	var prevX, prevY float64
	var prevState int // 0=invalid, 1=in-range, 2=above-max, 3=below-min
	flushSub := func() {
		if len(current) > 0 {
			subPaths = append(subPaths, current)
			current = nil
		}
	}
	classify := func(y float64) int {
		if math.IsNaN(y) || math.IsInf(y, 0) {
			return 0
		}
		if y > g.axes.yMax {
			return 2
		}
		if y < g.axes.yMin {
			return 3
		}
		return 1
	}
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		x := g.xMin + t*(g.xMax-g.xMin)
		y := g.fn(x)
		state := classify(y)

		if state == 0 {
			// Discontinuity ends the current subpath.
			flushSub()
			prevState = 0
			prevX, prevY = x, y
			continue
		}

		// Crossing yMin or yMax between prev and current — interpolate
		// the boundary point so the line meets the chart edge.
		if prevState != 0 && prevState != state && (prevState == 1 || state == 1) {
			bound := g.axes.yMax
			if prevState == 3 || state == 3 {
				bound = g.axes.yMin
			}
			// Linear interpolation between prev and current samples to find
			// where y = bound. Good enough at 200 samples; for coarser plots
			// the fn is sampled often enough that error stays sub-pixel.
			dy := y - prevY
			if math.Abs(dy) > 1e-12 {
				u := (bound - prevY) / dy
				if u >= 0 && u <= 1 {
					bx := prevX + u*(x-prevX)
					current = append(current, g.axes.PointAt(bx, bound))
				}
			}
			if state != 1 {
				// Crossing out — close subpath after boundary point.
				flushSub()
				prevState = state
				prevX, prevY = x, y
				continue
			}
			// Crossing in — boundary is the start of a new subpath, so
			// flush any (empty) current and start fresh with the boundary
			// point we just added.
			if len(current) > 1 {
				// Shouldn't happen — we only added one point above — but
				// guard anyway.
				current = current[len(current)-1:]
			}
		}

		if state == 1 {
			current = append(current, g.axes.PointAt(x, y))
		} else {
			// Out of range and no crossing this step — drop the subpath.
			flushSub()
		}
		prevState = state
		prevX, prevY = x, y
	}
	flushSub()

	if len(subPaths) == 0 {
		return
	}

	path := geometry.NewPath()
	for _, pts := range subPaths {
		if len(pts) < 2 {
			continue
		}
		sub := geometry.CatmullRomPath(pts)
		if tok.Roughness > 0 {
			sub = jitterCubicPath(sub, tok.MaxJitter, tok.Roughness, g.seed)
		}
		path.Append(sub)
	}

	if len(path.Cmds) == 0 {
		return
	}
	if g.reveal < 1 {
		path = geometry.PathPrefix(path, geometry.PathLength(path)*g.reveal)
	}
	r.DrawPath(path, stroke)
}

// jitterCubicPath displaces each cubic-Bezier control point perpendicular
// to the segment from start to end, producing a "drawn" wobble while
// keeping the path continuous (no per-segment MoveTo). Anchor points are
// preserved so the curve still passes through every sample.
//
// The PRNG is seeded by `seed` so the same Graph wobbles the same way
// frame-to-frame (the rough engine uses this same trick for all the
// other primitives — stable jitter across the animation).
func jitterCubicPath(p *geometry.Path, maxJitter, roughness float64, seed int64) *geometry.Path {
	if maxJitter <= 0 || roughness <= 0 {
		return p
	}
	src := rand.NewPCG(uint64(seed)+0x9E3779B97F4A7C15, 0xBF58476D1CE4E5B9)
	rng := rand.New(src)
	// Amplitude tuned to match the rough engine's line wobble visually
	// when the underlying mathx Sloppiness is light/normal: a small
	// fraction of MaxJitter, so the curve reads as hand-drawn rather
	// than noisy.
	amp := maxJitter * roughness * 0.35
	out := geometry.NewPath()
	var current geometry.Point
	for _, c := range p.Cmds {
		switch c.Kind {
		case geometry.CmdMove:
			out.MoveTo(c.P0.X, c.P0.Y)
			current = c.P0
		case geometry.CmdLine:
			out.LineTo(c.P0.X, c.P0.Y)
			current = c.P0
		case geometry.CmdCurve:
			// Perpendicular to the chord (current -> c.P2) so wobble
			// shows as bowing, not as randomness along the curve.
			dx := c.P2.X - current.X
			dy := c.P2.Y - current.Y
			l := math.Hypot(dx, dy)
			var nx, ny float64
			if l > 1e-9 {
				nx = -dy / l
				ny = dx / l
			}
			j1 := (rng.Float64()*2 - 1) * amp
			j2 := (rng.Float64()*2 - 1) * amp
			out.CurveTo(
				c.P0.X+nx*j1, c.P0.Y+ny*j1,
				c.P1.X+nx*j2, c.P1.Y+ny*j2,
				c.P2.X, c.P2.Y,
			)
			current = c.P2
		case geometry.CmdClose:
			out.Close()
		}
	}
	return out
}
