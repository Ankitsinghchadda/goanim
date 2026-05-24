package systemdesign

import (
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// RoutingStyle picks how an arrow draws its shaft.
type RoutingStyle uint8

const (
	RoutingStraight   RoutingStyle = iota // single straight line
	RoutingOrthogonal                     // right-angle path
	RoutingCurved                         // smooth cubic Bézier
)

// Arrow connects two mobjects. By default it auto-picks the edges
// based on the dominant axis between centers; users override with
// .From()/.To() and choose a routing style with .WithRouting().
type Arrow struct {
	*mobject.Group
	from, to mobject.Mobject
	fromSide mobject.Side // SideAuto = pick automatically
	toSide   mobject.Side
	routing  RoutingStyle
	label    string
	labelPos float64 // 0..1 along path; default 0.5
	style    style.Style
	reveal   float64
}

// NewArrow constructs an arrow between two mobjects with auto edge
// selection and straight routing.
func NewArrow(seed int64, from, to mobject.Mobject) *Arrow {
	return &Arrow{
		Group:    mobject.NewGroup(seed),
		from:     from,
		to:       to,
		fromSide: mobject.SideAuto,
		toSide:   mobject.SideAuto,
		routing:  RoutingStraight,
		labelPos: 0.5,
		reveal:   1,
	}
}

// From overrides the auto-selected source side.
func (a *Arrow) From(side mobject.Side) *Arrow { a.fromSide = side; return a }

// To overrides the auto-selected target side.
func (a *Arrow) To(side mobject.Side) *Arrow { a.toSide = side; return a }

// WithRouting selects the shaft style.
func (a *Arrow) WithRouting(r RoutingStyle) *Arrow { a.routing = r; return a }

// WithLabel attaches a label to the arrow.
func (a *Arrow) WithLabel(s string) *Arrow { a.label = s; return a }

// WithLabelPosition sets the label position along the path in [0, 1].
// Default 0.5 (midpoint).
func (a *Arrow) WithLabelPosition(t float64) *Arrow {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	a.labelPos = t
	return a
}

// WithStyle sets the per-mobject style override.
func (a *Arrow) WithStyle(s style.Style) *Arrow { a.style = s; return a }

// Style returns the per-mobject style override (in-place editable).
func (a *Arrow) Style() *style.Style { return &a.style }

// SetStyle replaces the style override.
func (a *Arrow) SetStyle(s style.Style) { a.style = s }

// SetReveal sets the arrow's reveal fraction.
func (a *Arrow) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	a.reveal = t
}

// Reveal returns the current reveal fraction.
func (a *Arrow) Reveal() float64 { return a.reveal }

// Bounds returns the bounding box of the resolved path.
func (a *Arrow) Bounds() geometry.Rect {
	p1, p2 := a.endpoints()
	return geometry.RectFromPoints(p1, p2)
}

// Endpoints returns (start, end) of the arrow with the shrink cushion
// applied — useful for packet animations.
func (a *Arrow) Endpoints() (geometry.Point, geometry.Point) {
	p1, p2 := a.endpoints()
	return p1, p2
}

// endpoints resolves the source and target attachment points.
func (a *Arrow) endpoints() (geometry.Point, geometry.Point) {
	fromB := a.from.Bounds()
	toB := a.to.Bounds()

	fs, ts := autoAttachSides(a.fromSide, a.toSide, fromB, toB)

	p1 := attachmentPoint(a.from, fs)
	p2 := attachmentPoint(a.to, ts)
	return shrinkToward(p1, p2, 6)
}

// autoAttachSides picks the from/target attachment sides, honoring
// any explicit override and falling back to a geometry-aware
// AutoAttach for SideAuto entries.
//
// The wrinkle this corrects: classic AutoAttach picks horizontal vs
// vertical purely on |dx| vs |dy| of centers. But when two boxes are
// close along the chosen axis, the resulting two-bend Z route ends
// up with tiny stub segments that read as glitches. Specifically,
// for a horizontal exit, if the gap between source.left and
// target.right (or source.right and target.left) is less than
// minHorizontalGap, the side stubs of the Z route are < minSegmentLen.
// In that case we promote to a vertical exit (always at least as
// natural when boxes are stacked, and never produces tiny stubs at
// the source side because the source's vertical extent is fully
// available for the long downward leg).
func autoAttachSides(fromOverride, toOverride mobject.Side, fromB, toB geometry.Rect) (mobject.Side, mobject.Side) {
	fromC := fromB.Center()
	toC := toB.Center()

	// Minimum gap between near edges along the chosen axis. Twice
	// minSegmentLen accounts for the two side stubs each needing room.
	const minAxisGap = 2 * minSegmentLen

	fs := fromOverride
	if fs == mobject.SideAuto {
		fs = mobject.AutoAttach(fromC, toC)

		// Check whether the chosen side's axis has enough edge-to-edge
		// gap to produce non-stubby segments. If not, promote to the
		// perpendicular axis.
		switch fs {
		case mobject.SideLeft:
			// Source exits left; target should be properly to the left,
			// i.e. target.right.x + minAxisGap <= source.left.x.
			if toB.Max.X+minAxisGap > fromB.Min.X {
				fs = verticalSideFor(fromC, toC)
			}
		case mobject.SideRight:
			if fromB.Max.X+minAxisGap > toB.Min.X {
				fs = verticalSideFor(fromC, toC)
			}
		case mobject.SideTop:
			if fromB.Max.Y+minAxisGap > toB.Min.Y {
				fs = horizontalSideFor(fromC, toC)
			}
		case mobject.SideBottom:
			if toB.Max.Y+minAxisGap > fromB.Min.Y {
				fs = horizontalSideFor(fromC, toC)
			}
		}
	}

	ts := toOverride
	if ts == mobject.SideAuto {
		ts = mobject.FlipSide(fs)
	}
	return fs, ts
}

// verticalSideFor returns SideTop or SideBottom depending on which
// way the target lies from the source center.
func verticalSideFor(fromC, toC geometry.Point) mobject.Side {
	if toC.Y >= fromC.Y {
		return mobject.SideTop
	}
	return mobject.SideBottom
}

// horizontalSideFor returns SideLeft or SideRight depending on which
// way the target lies from the source center.
func horizontalSideFor(fromC, toC geometry.Point) mobject.Side {
	if toC.X >= fromC.X {
		return mobject.SideRight
	}
	return mobject.SideLeft
}

// attachmentPoint queries the mobject's Attachable interface, falling
// back to bounding-box-edge midpoint if not implemented.
func attachmentPoint(m mobject.Mobject, side mobject.Side) geometry.Point {
	if a, ok := m.(mobject.Attachable); ok {
		return a.AttachmentPoint(side)
	}
	return mobject.AttachToBoundsEdge(m.Bounds(), side)
}

// pathSegments returns the polyline describing the arrow shaft for
// the resolved endpoints and routing style. Each returned segment is
// a pair of points; consecutive segments share a vertex.
//
// Orthogonal routing decisions (Phase-4 polish):
//
//   - If endpoints are nearly aligned on the routing axis (within
//     minStraightTolerance), collapse to a straight line — a 5-pixel
//     wobble doesn't deserve a 90° bend.
//   - Prefer single-bend over two-bend whenever endpoint sides allow.
//   - For perpendicular-side single-bend, place the bend point so the
//     segment that exits the source's side is the "long" one (the
//     arrow then reads as "going right with a small adjustment down,"
//     not "going down then right").
//   - Reject any bend that would produce a segment shorter than
//     minSegmentLen — such bends look glitchy. Fall back to two-bend
//     in the middle of the span when this happens.
const (
	minStraightTolerance = 6.0  // px — under this misalignment, route straight
	minSegmentLen        = 30.0 // px — segments shorter than this are forbidden
)

func (a *Arrow) pathSegments() [][2]geometry.Point {
	p1, p2 := a.endpoints()
	switch a.routing {
	case RoutingOrthogonal:
		fs, ts := a.resolvedSides()
		return orthogonalRoute(p1, p2, fs, ts)
	default:
		return [][2]geometry.Point{{p1, p2}}
	}
}

// resolvedSides returns the from/to sides, falling back to AutoAttach
// when SideAuto is set. Mirrors the logic in endpoints() — the two
// must agree on attachment sides so the rendered path lines up with
// the rendered endpoints.
func (a *Arrow) resolvedSides() (mobject.Side, mobject.Side) {
	return autoAttachSides(a.fromSide, a.toSide, a.from.Bounds(), a.to.Bounds())
}

// orthogonalRoute decides the bend layout for orthogonal arrows.
//
// Phase-5 rule: when EITHER segment of a single-bend route would be
// shorter than minSegmentLen, OR when the parallel-sides two-bend
// route would have a short middle segment, collapse to a straight
// line. A 90° turn that produces a 10-px stub looks like a glitch;
// a slightly angled straight line reads as a single deliberate flow.
func orthogonalRoute(p1, p2 geometry.Point, fs, ts mobject.Side) [][2]geometry.Point {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	// Collapse to straight line when the perpendicular component is small.
	if isHorizontalSide(fs) && absRouting(dy) < minStraightTolerance {
		return [][2]geometry.Point{{p1, p2}}
	}
	if isVerticalSide(fs) && absRouting(dx) < minStraightTolerance {
		return [][2]geometry.Point{{p1, p2}}
	}

	// Single-bend cases — when source and target sides are perpendicular.
	if isHorizontalSide(fs) && isVerticalSide(ts) {
		// Bend at (target.X, source.Y) — the arrow exits source horizontally,
		// then turns down/up at the target's column.
		bend := geometry.Pt(p2.X, p1.Y)
		if segLengthOK(p1, bend) && segLengthOK(bend, p2) {
			return [][2]geometry.Point{{p1, bend}, {bend, p2}}
		}
		// Bend produces a too-short segment — straighter is better.
		return [][2]geometry.Point{{p1, p2}}
	}
	if isVerticalSide(fs) && isHorizontalSide(ts) {
		// Mirror: bend at (source.X, target.Y).
		bend := geometry.Pt(p1.X, p2.Y)
		if segLengthOK(p1, bend) && segLengthOK(bend, p2) {
			return [][2]geometry.Point{{p1, bend}, {bend, p2}}
		}
		return [][2]geometry.Point{{p1, p2}}
	}

	// Two-bend fallback. For parallel sides (both horizontal or both
	// vertical) we route out, across, in — but only when the cross
	// segment is long enough to justify two right-angle turns. When
	// the source and target columns (or rows) are nearly aligned, the
	// cross segment shrinks to almost nothing and the visual reads as
	// a stutter; collapse to a straight line in that case so the
	// arrow flows as a single deliberate stroke.
	if isHorizontalSide(fs) && isHorizontalSide(ts) {
		// Middle (cross) segment is VERTICAL with length |dy|.
		if absRouting(dy) < minSegmentLen {
			return [][2]geometry.Point{{p1, p2}}
		}
		midX := (p1.X + p2.X) / 2
		b1 := geometry.Pt(midX, p1.Y)
		b2 := geometry.Pt(midX, p2.Y)
		return [][2]geometry.Point{{p1, b1}, {b1, b2}, {b2, p2}}
	}
	// Both-vertical sides: middle (cross) segment is HORIZONTAL with length |dx|.
	if absRouting(dx) < minSegmentLen {
		return [][2]geometry.Point{{p1, p2}}
	}
	midY := (p1.Y + p2.Y) / 2
	b1 := geometry.Pt(p1.X, midY)
	b2 := geometry.Pt(p2.X, midY)
	return [][2]geometry.Point{{p1, b1}, {b1, b2}, {b2, p2}}
}

func segLengthOK(a, b geometry.Point) bool { return a.Distance(b) >= minSegmentLen }
func absRouting(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Render draws the arrow shaft, arrowhead, and (optionally) a label.
func (a *Arrow) Render(r render.Renderer, ctx style.Context) {
	if a.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(a.style)
	tok := style.TokensFor(eff)

	stroke := style.PathStyleStroke(eff, tok)
	stroke.DashArray = tok.DashArray

	switch a.routing {
	case RoutingCurved:
		a.renderCurved(r, eff, tok, stroke)
	default:
		a.renderSegments(r, eff, tok, stroke)
	}

	// Arrowhead at the final endpoint, aligned with the last segment's
	// direction. Always rendered solid (no dashes), and only when the
	// reveal is far enough along that the head wouldn't appear before
	// the shaft reaches the target.
	if a.reveal < 0.95 {
		return
	}
	a.renderArrowhead(r, eff, tok, stroke)

	if a.label != "" {
		a.renderLabel(r, ctx, eff, tok)
	}
}

// renderSegments handles straight and orthogonal routing.
func (a *Arrow) renderSegments(r render.Renderer, eff style.Style, tok style.Tokens, stroke render.PathStyle) {
	segs := a.pathSegments()
	totalLen := 0.0
	for _, s := range segs {
		totalLen += s[0].Distance(s[1])
	}
	if totalLen == 0 {
		return
	}
	consumed := 0.0
	for i, seg := range segs {
		segLen := seg[0].Distance(seg[1])
		// Reveal: render the segment fully if all of it lies within the
		// reveal budget; partially if reveal cuts through it.
		end := seg[1]
		if a.reveal < 1 {
			revealLen := totalLen * a.reveal
			switch {
			case consumed+segLen <= revealLen:
				// segment fully within reveal budget; end stays at seg[1]
			case consumed >= revealLen:
				return
			default:
				drawLen := revealLen - consumed
				if drawLen <= 0 {
					return
				}
				end = seg[0].Lerp(seg[1], drawLen/segLen)
			}
		}
		var path *geometry.Path
		if tok.Roughness == 0 {
			path = geometry.LinePath(seg[0], end)
		} else {
			opts := style.RoughOptions(eff, tok, a.Seed()+int64(i*100))
			opts.Bowing = tok.Bowing
			// PreserveVertices=true on bends keeps the rough lines'
			// endpoints exactly at the bend point, so the corner reads
			// as a continuous joint rather than a hard kink.
			opts.PreserveVertices = true
			// Arrows render as single-stroke shafts even in sketchy
			// mode. The rough-line double-stroke makes ORTHOGONAL
			// arrows look distinctly different from STRAIGHT arrows
			// (orthogonal accumulates 2 passes per segment, so a
			// 3-segment Z carries 6 overlapping curves vs a straight
			// arrow's 2). Single-stroke shafts read uniformly across
			// all routing styles while keeping the hand-drawn wobble.
			opts.DisableMultiStroke = true
			path = rough.RoughLine(seg[0], end, opts)
		}
		r.DrawPath(path, stroke)
		consumed += segLen
		if a.reveal < 1 && consumed >= totalLen*a.reveal {
			return
		}
	}
}

// renderCurved handles curved (cubic Bézier) routing.
func (a *Arrow) renderCurved(r render.Renderer, eff style.Style, tok style.Tokens, stroke render.PathStyle) {
	p1, p2 := a.endpoints()
	fs, ts := a.resolvedSides()
	chord := p1.Distance(p2)
	handle := chord * 0.4
	n1 := mobject.SideNormal(fs)
	n2 := mobject.SideNormal(ts)
	c1 := geometry.Pt(p1.X+n1.X*handle, p1.Y+n1.Y*handle)
	c2 := geometry.Pt(p2.X+n2.X*handle, p2.Y+n2.Y*handle)
	path := geometry.NewPath()
	path.MoveTo(p1.X, p1.Y)
	path.CurveTo(c1.X, c1.Y, c2.X, c2.Y, p2.X, p2.Y)
	if a.reveal < 1 {
		path = geometry.PathPrefix(path, geometry.PathLength(path)*a.reveal)
	}
	r.DrawPath(path, stroke)
}

// renderArrowhead draws the chevron at the final endpoint, aligned with
// the final segment's tangent direction.
func (a *Arrow) renderArrowhead(r render.Renderer, eff style.Style, tok style.Tokens, stroke render.PathStyle) {
	dir, tip := a.endTangent()
	headLen := math.Max(tok.StrokeWidthPx*8, 18)
	const headAngle = 0.45 // ~26°
	cosA, sinA := math.Cos(headAngle), math.Sin(headAngle)
	back := dir.Scale(-1)
	left := geometry.Pt(back.X*cosA-back.Y*sinA, back.X*sinA+back.Y*cosA).Scale(headLen)
	right := geometry.Pt(back.X*cosA+back.Y*sinA, -back.X*sinA+back.Y*cosA).Scale(headLen)
	tipBack := tip.Add(back.Scale(2))

	headStroke := stroke
	headStroke.DashArray = nil

	if tok.Roughness == 0 {
		r.DrawPath(geometry.LinePath(tip, tipBack.Add(left)), headStroke)
		r.DrawPath(geometry.LinePath(tip, tipBack.Add(right)), headStroke)
		return
	}
	ho := style.RoughOptions(eff, tok, a.Seed()+13)
	ho.Roughness = tok.Roughness * 0.5
	ho.DisableMultiStroke = true
	r.DrawPath(rough.RoughLine(tip, tipBack.Add(left), ho), headStroke)
	ho.Seed = a.Seed() + 29
	r.DrawPath(rough.RoughLine(tip, tipBack.Add(right), ho), headStroke)
}

// endTangent returns the unit tangent (direction) and tip point at the
// arrow's end. For orthogonal routing this is the final segment's
// direction; for curved it's the Bézier tangent; for straight it's the
// line direction.
func (a *Arrow) endTangent() (geometry.Point, geometry.Point) {
	switch a.routing {
	case RoutingCurved:
		p1, p2 := a.endpoints()
		_, ts := a.resolvedSides()
		n2 := mobject.SideNormal(ts)
		c2 := geometry.Pt(p2.X+n2.X*p1.Distance(p2)*0.4, p2.Y+n2.Y*p1.Distance(p2)*0.4)
		// Tangent at endpoint of cubic = p2 - c2 (direction into p2 from c2).
		dir := p2.Sub(c2).Normalize()
		return dir, p2
	default:
		segs := a.pathSegments()
		if len(segs) == 0 {
			return geometry.Pt(1, 0), geometry.Pt(0, 0)
		}
		last := segs[len(segs)-1]
		dir := last[1].Sub(last[0]).Normalize()
		return dir, last[1]
	}
}

// renderLabel draws the arrow label at the parametric position along
// the path, with a style-aware background.
func (a *Arrow) renderLabel(r render.Renderer, ctx style.Context, eff style.Style, tok style.Tokens) {
	pos := a.labelPositionPoint()

	// Estimate label size (Phase-2-style approximation).
	labelTok := tok
	labelTok.FontSizePx = tok.FontSizePx * 0.7 // labels smaller than node labels
	textW := float64(len(a.label)) * labelTok.FontSizePx * 0.55
	textH := labelTok.FontSizePx * 0.9
	padX := labelTok.FontSizePx * 0.4
	padY := labelTok.FontSizePx * 0.2

	bgPath := geometry.RectanglePath(
		pos.X-textW/2-padX, pos.Y-textH/2-padY,
		textW+2*padX, textH+2*padY,
		6,
	)
	bgColor := labelBackgroundColor(ctx.SceneDefault)
	if bgColor != nil {
		r.DrawPath(bgPath, render.PathStyle{Fill: bgColor})
	}
	face := ctx.FontFace(eff.FontFamily)
	r.DrawText(a.label, pos.X, pos.Y, render.TextStyle{
		Face:     face,
		Size:     labelTok.FontSizePx,
		Color:    eff.StrokeColor,
		Align:    render.AlignCenter,
		Baseline: render.BaselineMiddle,
	})
}

// labelPositionPoint computes the point along the path at parameter
// labelPos, measured by arc length.
//
// For straight: lerp(start, end, t).
// For orthogonal: walk the segments cumulatively until the target arc
// length is reached, then lerp within that segment. This produces a
// label that sits ON the path even when the path bends — e.g. an
// arrow that goes "right 100px, then down 200px" puts its midpoint
// label 50px below the bend on the vertical segment, not in the
// empty space halfway between endpoints.
// For curved: cubic-bezier arc-length walk.
func (a *Arrow) labelPositionPoint() geometry.Point {
	switch a.routing {
	case RoutingOrthogonal:
		segs := a.pathSegments()
		total := 0.0
		for _, s := range segs {
			total += s[0].Distance(s[1])
		}
		target := total * a.labelPos
		acc := 0.0
		for _, s := range segs {
			d := s[0].Distance(s[1])
			if acc+d >= target {
				return s[0].Lerp(s[1], (target-acc)/d)
			}
			acc += d
		}
		return segs[len(segs)-1][1]
	case RoutingCurved:
		// Sample the cubic Bézier at fine resolution and walk arc length.
		p1, p2 := a.endpoints()
		fs, ts := a.resolvedSides()
		chord := p1.Distance(p2)
		handle := chord * 0.4
		n1 := mobject.SideNormal(fs)
		n2 := mobject.SideNormal(ts)
		c1 := geometry.Pt(p1.X+n1.X*handle, p1.Y+n1.Y*handle)
		c2 := geometry.Pt(p2.X+n2.X*handle, p2.Y+n2.Y*handle)
		const samples = 64
		prev := p1
		var total float64
		segments := make([]struct {
			start, end geometry.Point
			len        float64
		}, samples)
		for i := 1; i <= samples; i++ {
			t := float64(i) / float64(samples)
			pt := cubicBezPoint(p1, c1, c2, p2, t)
			d := prev.Distance(pt)
			segments[i-1].start = prev
			segments[i-1].end = pt
			segments[i-1].len = d
			total += d
			prev = pt
		}
		target := total * a.labelPos
		acc := 0.0
		for _, s := range segments {
			if acc+s.len >= target {
				return s.start.Lerp(s.end, (target-acc)/s.len)
			}
			acc += s.len
		}
		return p2
	default:
		p1, p2 := a.endpoints()
		return p1.Lerp(p2, a.labelPos)
	}
}

// cubicBezPoint evaluates a cubic Bézier at parameter t.
func cubicBezPoint(p0, p1, p2, p3 geometry.Point, t float64) geometry.Point {
	u := 1 - t
	uu := u * u
	uuu := uu * u
	tt := t * t
	ttt := tt * t
	return geometry.Pt(
		uuu*p0.X+3*uu*t*p1.X+3*u*tt*p2.X+ttt*p3.X,
		uuu*p0.Y+3*uu*t*p1.Y+3*u*tt*p2.Y+ttt*p3.Y,
	)
}

// labelBackgroundColor picks a subtle backing color for the label
// based on the scene's stroke/fill palette. Returns nil if the label
// should be drawn without a background (very small label, etc.).
func labelBackgroundColor(scene style.Style) interface {
	RGBA() (uint32, uint32, uint32, uint32)
} {
	// Strategy: a pale neutral that contrasts with the stroke. We
	// avoid white-on-white by checking the scene fill color. If the
	// scene already has a pale fill, use a slightly tinted version.
	return scene.FillColor
}

func isHorizontalSide(s mobject.Side) bool { return s == mobject.SideLeft || s == mobject.SideRight }
func isVerticalSide(s mobject.Side) bool   { return s == mobject.SideTop || s == mobject.SideBottom }

// shrinkToward pulls both endpoints in toward each other by d.
func shrinkToward(p1, p2 geometry.Point, d float64) (geometry.Point, geometry.Point) {
	v := p2.Sub(p1)
	l := v.Length()
	if l <= 2*d {
		return p1, p2
	}
	u := v.Scale(1 / l)
	return p1.Add(u.Scale(d)), p2.Sub(u.Scale(d))
}
