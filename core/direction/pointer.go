package direction

import (
	"image/color"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/animation/easing"
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// LaserPath is the parameterized curve a laser pointer traces. It's
// independent of geometry.Path so we can also build a path from a list
// of mobject centers via PathThrough without forcing the caller to
// touch geometry directly.
//
// The path is sampled by arc length: PointAt(0) is the start,
// PointAt(Length()) is the end. Used by the laser pointer's Apply to
// position the dot at the current arc-length progress.
type LaserPath struct {
	pts []geometry.Point
	// cumulative arc length up to each point in pts; len == len(pts).
	cum []float64
}

// NewLaserPathFromPoints builds a polyline path through the given
// points in order. Two points → straight line. Three or more →
// connected segments.
func NewLaserPathFromPoints(pts ...geometry.Point) LaserPath {
	lp := LaserPath{pts: append([]geometry.Point(nil), pts...)}
	lp.cum = make([]float64, len(pts))
	for i := 1; i < len(pts); i++ {
		lp.cum[i] = lp.cum[i-1] + pts[i].Sub(pts[i-1]).Length()
	}
	return lp
}

// PathThrough constructs a laser path passing through each mobject's
// center. Each mobject must implement Position() (Group and IconBase
// both do); for ones that don't, falls back to Bounds().Center().
//
//	scene.Play(direction.LaserPointer(
//	    direction.PathThrough(client, server, database),
//	    3*time.Second,
//	))
func PathThrough(targets ...mobject.Mobject) LaserPath {
	pts := make([]geometry.Point, 0, len(targets))
	for _, m := range targets {
		x, y := positionFor(m)
		pts = append(pts, geometry.Pt(x, y))
	}
	return NewLaserPathFromPoints(pts...)
}

// Length returns the total arc length of the path.
func (lp LaserPath) Length() float64 {
	if len(lp.cum) == 0 {
		return 0
	}
	return lp.cum[len(lp.cum)-1]
}

// PointAt returns the point at the given arc-length distance from the
// start. Clamps to the path's endpoints.
func (lp LaserPath) PointAt(arc float64) geometry.Point {
	if len(lp.pts) == 0 {
		return geometry.Point{}
	}
	if arc <= 0 {
		return lp.pts[0]
	}
	total := lp.Length()
	if arc >= total {
		return lp.pts[len(lp.pts)-1]
	}
	// Binary search would be ideal, but laser paths are short — linear
	// scan is fine.
	for i := 1; i < len(lp.cum); i++ {
		if arc <= lp.cum[i] {
			seg := lp.cum[i] - lp.cum[i-1]
			if seg == 0 {
				return lp.pts[i]
			}
			t := (arc - lp.cum[i-1]) / seg
			a, b := lp.pts[i-1], lp.pts[i]
			return geometry.Pt(a.X+(b.X-a.X)*t, a.Y+(b.Y-a.Y)*t)
		}
	}
	return lp.pts[len(lp.pts)-1]
}

// LaserPointerOptions tunes the dot's appearance. Zero values fall
// back to the package defaults: bright red dot, ~7px radius, 150px
// trail, EaseInOutCubic easing.
type LaserPointerOptions struct {
	Color       color.Color          // default: bright red (#FF3333)
	DotSize     float64              // default: 7
	TrailLength float64              // pixels of trail, default 150
	Easing      animation.EasingFunc // default: EaseInOutCubic
}

// LaserPointer animates a dot+glow+trail along the path over duration.
//
// The returned animation owns a laser-pointer mobject that gets added
// to the scene on Play (via Targets()) and renders on top of every
// other mobject because the scene appends new targets to the end of
// its mobject list (later draw = higher z-order). At t=1 the laser
// fades out so it doesn't linger as a stray red dot.
//
// The laser is intentionally style-agnostic — even at maximum scene
// sloppiness the dot is clean and the trail is solid. A wobbly laser
// pointer is the wrong visual: the laser is the PRESENTER's tool, not
// a diagram element.
func LaserPointer(path LaserPath, duration time.Duration, opts ...LaserPointerOptions) animation.Animation {
	o := LaserPointerOptions{
		Color:       color.RGBA{0xFF, 0x33, 0x33, 0xFF},
		DotSize:     7,
		TrailLength: 150,
		Easing:      easing.InOutCubic,
	}
	if len(opts) > 0 {
		if opts[0].Color != nil {
			o.Color = opts[0].Color
		}
		if opts[0].DotSize > 0 {
			o.DotSize = opts[0].DotSize
		}
		if opts[0].TrailLength > 0 {
			o.TrailLength = opts[0].TrailLength
		}
		if opts[0].Easing != nil {
			o.Easing = opts[0].Easing
		}
	}
	dot := newLaserDot(o)
	if path.Length() > 0 {
		dot.SetPositionAt(path.PointAt(0))
	}
	return &laserAnim{path: path, dot: dot, dur: duration, ease: o.Easing}
}

type laserAnim struct {
	path LaserPath
	dot  *laserDot
	dur  time.Duration
	ease animation.EasingFunc
}

func (l *laserAnim) Apply(t float64) {
	t = clamp01(t)
	if l.path.Length() == 0 {
		return
	}
	arc := l.path.Length() * l.ease(t)
	l.dot.SetArc(arc, l.path)
	// Fade out in the last 10% so the dot doesn't pop off-screen.
	if t > 0.9 {
		l.dot.opacity = 1 - (t-0.9)/0.1
	} else {
		l.dot.opacity = 1
	}
}
func (l *laserAnim) Duration() time.Duration    { return l.dur }
func (l *laserAnim) Targets() []mobject.Mobject { return []mobject.Mobject{l.dot} }

// laserDot is the mobject that renders the dot+glow+trail. Held by
// reference inside laserAnim so the animation can update its arc
// position without per-frame mobject lookup.
type laserDot struct {
	*mobject.Group
	opts    LaserPointerOptions
	cx, cy  float64
	opacity float64
	// trail stores recent positions for rendering the fading trail. We
	// keep a fixed-size ring rather than recomputing per-frame so the
	// trail looks smooth at any animation easing.
	trail    []geometry.Point
	trailMax int
}

func newLaserDot(opts LaserPointerOptions) *laserDot {
	return &laserDot{
		Group:    mobject.NewGroup(0),
		opts:     opts,
		opacity:  1,
		trailMax: 32, // 32 trail samples is plenty for visual smoothness
	}
}

// SetPositionAt jumps the dot to a point without leaving a trail
// segment. Used to seed the dot's starting position.
func (l *laserDot) SetPositionAt(p geometry.Point) {
	l.cx, l.cy = p.X, p.Y
	l.trail = nil
}

// SetArc updates the dot to follow the laser path at the given arc
// length. Records the new position in the trail buffer.
func (l *laserDot) SetArc(arc float64, path LaserPath) {
	p := path.PointAt(arc)
	l.cx, l.cy = p.X, p.Y
	l.trail = append(l.trail, p)
	if len(l.trail) > l.trailMax {
		l.trail = l.trail[len(l.trail)-l.trailMax:]
	}
}

// Bounds — the dot is a small fixed-size circle; the trail extends
// behind it. Return a generous bounding rect around the current
// position so the renderer doesn't clip the trail.
func (l *laserDot) Bounds() geometry.Rect {
	r := l.opts.TrailLength + l.opts.DotSize*3
	return geometry.RectFromCenter(geometry.Pt(l.cx, l.cy), r*2, r*2)
}
func (l *laserDot) Children() []mobject.Mobject  { return nil }
func (l *laserDot) Position() (float64, float64) { return l.cx, l.cy }
func (l *laserDot) SetPosition(x, y float64)     { l.cx, l.cy = x, y }
func (l *laserDot) SetReveal(t float64)          { l.opacity = clamp01(t) }
func (l *laserDot) SetVisualScale(float64)       {}

// Render — paints the trail (fading rear-to-front), then the glow,
// then the dot. The laser ignores style.Sloppiness intentionally:
// it's a presenter tool, not a diagram element.
func (l *laserDot) Render(rd render.Renderer, _ style.Context) {
	if l.opacity <= 0 || (l.cx == 0 && l.cy == 0 && len(l.trail) == 0) {
		return
	}
	// Trail: alpha rises from 0 at oldest sample to opts.Color's full
	// alpha at the dot. Drawn as straight line segments through the
	// trail buffer.
	tlen := len(l.trail)
	if tlen >= 2 {
		for i := 1; i < tlen; i++ {
			frac := float64(i) / float64(tlen-1) // 0 at oldest, 1 at newest
			a := style.ApplyOpacity(l.opts.Color, frac*l.opacity*0.6)
			line := geometry.LinePath(l.trail[i-1], l.trail[i])
			rd.DrawPath(line, render.PathStyle{
				Stroke:      a,
				StrokeWidth: l.opts.DotSize * 0.7 * frac,
				StrokeCap:   render.CapRound,
			})
		}
	}
	// Glow: a larger, more-transparent disc behind the dot.
	glowCol := style.ApplyOpacity(l.opts.Color, l.opacity*0.35)
	rd.DrawPath(
		geometry.EllipsePath(l.cx, l.cy, l.opts.DotSize*1.9, l.opts.DotSize*1.9),
		render.PathStyle{Fill: glowCol},
	)
	// Dot.
	dotCol := style.ApplyOpacity(l.opts.Color, l.opacity)
	rd.DrawPath(
		geometry.EllipsePath(l.cx, l.cy, l.opts.DotSize, l.opts.DotSize),
		render.PathStyle{Fill: dotCol},
	)
}
