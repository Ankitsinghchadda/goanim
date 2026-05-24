package geometry

import "math"

// RectanglePath returns a clean (non-rough) rectangle path. If
// cornerRadius > 0, all four corners are rounded with quarter-circle
// arcs approximated by cubic Béziers.
func RectanglePath(x, y, w, h, cornerRadius float64) *Path {
	if cornerRadius <= 0 {
		p := NewPath()
		p.MoveTo(x, y)
		p.LineTo(x+w, y)
		p.LineTo(x+w, y+h)
		p.LineTo(x, y+h)
		p.Close()
		return p
	}
	r := cornerRadius
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}
	// Magic constant for quarter-circle bezier approximation.
	const k = 0.5522847498307936
	p := NewPath()
	p.MoveTo(x+r, y)
	p.LineTo(x+w-r, y)
	p.CurveTo(x+w-r+r*k, y, x+w, y+r-r*k, x+w, y+r)
	p.LineTo(x+w, y+h-r)
	p.CurveTo(x+w, y+h-r+r*k, x+w-r+r*k, y+h, x+w-r, y+h)
	p.LineTo(x+r, y+h)
	p.CurveTo(x+r-r*k, y+h, x, y+h-r+r*k, x, y+h-r)
	p.LineTo(x, y+r)
	p.CurveTo(x, y+r-r*k, x+r-r*k, y, x+r, y)
	p.Close()
	return p
}

// EllipsePath returns a clean ellipse path centered at (cx, cy) with
// radii (rx, ry), built from four cubic Bézier arcs.
func EllipsePath(cx, cy, rx, ry float64) *Path {
	const k = 0.5522847498307936
	p := NewPath()
	p.MoveTo(cx+rx, cy)
	p.CurveTo(cx+rx, cy+ry*k, cx+rx*k, cy+ry, cx, cy+ry)
	p.CurveTo(cx-rx*k, cy+ry, cx-rx, cy+ry*k, cx-rx, cy)
	p.CurveTo(cx-rx, cy-ry*k, cx-rx*k, cy-ry, cx, cy-ry)
	p.CurveTo(cx+rx*k, cy-ry, cx+rx, cy-ry*k, cx+rx, cy)
	p.Close()
	return p
}

// LinePath returns a simple two-point line.
func LinePath(p1, p2 Point) *Path {
	p := NewPath()
	p.MoveTo(p1.X, p1.Y)
	p.LineTo(p2.X, p2.Y)
	return p
}

// CatmullRomPath builds a smooth cubic-Bezier path that passes through
// every point in pts. The curve uses centripetal Catmull-Rom tangents
// converted to Bezier control points, which avoids cusps and overshoot
// when adjacent samples are unevenly spaced (e.g., near a steep region
// of a plotted function). Endpoints use a reflected phantom point so
// the curve is C1-continuous through every interior sample.
//
// With fewer than 2 points the path is empty; with exactly 2 it falls
// back to a straight line.
func CatmullRomPath(pts []Point) *Path {
	p := NewPath()
	n := len(pts)
	if n < 2 {
		return p
	}
	p.MoveTo(pts[0].X, pts[0].Y)
	if n == 2 {
		p.LineTo(pts[1].X, pts[1].Y)
		return p
	}
	// Build segments [i, i+1] with neighbors [i-1, i+2] for tangent context.
	// "Phantom" endpoints reflect the first/last interior tangent so end
	// curvature looks natural rather than flat.
	for i := 0; i < n-1; i++ {
		var p0, p3 Point
		p1 := pts[i]
		p2 := pts[i+1]
		if i == 0 {
			p0 = Point{X: 2*p1.X - pts[i+1].X, Y: 2*p1.Y - pts[i+1].Y}
		} else {
			p0 = pts[i-1]
		}
		if i+2 >= n {
			p3 = Point{X: 2*p2.X - pts[i].X, Y: 2*p2.Y - pts[i].Y}
		} else {
			p3 = pts[i+2]
		}
		// Catmull-Rom -> Bezier with tension 0.5 (the standard conversion).
		c1x := p1.X + (p2.X-p0.X)/6
		c1y := p1.Y + (p2.Y-p0.Y)/6
		c2x := p2.X - (p3.X-p1.X)/6
		c2y := p2.Y - (p3.Y-p1.Y)/6
		p.CurveTo(c1x, c1y, c2x, c2y, p2.X, p2.Y)
	}
	return p
}

// PathLength returns the approximate length of p, treating curves
// by flattening to ~16 segments each. Suitable for stroke-reveal math.
func PathLength(p *Path) float64 {
	total := 0.0
	const curveSubdiv = 16
	var current Point
	var subStart Point
	for _, c := range p.Cmds {
		switch c.Kind {
		case CmdMove:
			current = c.P0
			subStart = c.P0
		case CmdLine:
			total += current.Distance(c.P0)
			current = c.P0
		case CmdCurve:
			prev := current
			for i := 1; i <= curveSubdiv; i++ {
				t := float64(i) / float64(curveSubdiv)
				next := cubicBezier(current, c.P0, c.P1, c.P2, t)
				total += prev.Distance(next)
				prev = next
			}
			current = c.P2
		case CmdClose:
			total += current.Distance(subStart)
			current = subStart
		}
	}
	return total
}

// PathPrefix returns the prefix of p covering exactly `length` units
// along the path (measured by flattened arc length). Used by DrawOn-style
// animations to reveal the path progressively.
//
// The returned path preserves curve geometry where possible: a curve
// segment that's only partially covered is subdivided at the appropriate
// t-parameter using de Casteljau.
func PathPrefix(p *Path, length float64) *Path {
	if length <= 0 {
		return NewPath()
	}
	out := NewPath()
	remaining := length
	const curveSubdiv = 64
	var current Point
	var subStart Point
	emittedMove := false
	for _, c := range p.Cmds {
		if remaining <= 0 {
			break
		}
		switch c.Kind {
		case CmdMove:
			out.MoveTo(c.P0.X, c.P0.Y)
			emittedMove = true
			current = c.P0
			subStart = c.P0
		case CmdLine:
			d := current.Distance(c.P0)
			if d <= remaining {
				out.LineTo(c.P0.X, c.P0.Y)
				remaining -= d
				current = c.P0
			} else {
				t := remaining / d
				p := current.Lerp(c.P0, t)
				out.LineTo(p.X, p.Y)
				remaining = 0
				current = p
			}
		case CmdCurve:
			// Walk the curve in small steps until we exhaust remaining length.
			prev := current
			done := false
			for i := 1; i <= curveSubdiv; i++ {
				t := float64(i) / float64(curveSubdiv)
				next := cubicBezier(current, c.P0, c.P1, c.P2, t)
				step := prev.Distance(next)
				if step <= remaining {
					remaining -= step
					prev = next
					continue
				}
				// Subdivide at this t.
				sub := subdivideCubic(current, c.P0, c.P1, c.P2, t-1.0/float64(curveSubdiv)+remaining/step*(1.0/float64(curveSubdiv)))
				out.CurveTo(sub.c1.X, sub.c1.Y, sub.c2.X, sub.c2.Y, sub.end.X, sub.end.Y)
				remaining = 0
				prev = sub.end
				done = true
				break
			}
			if !done {
				// Whole curve fits.
				out.CurveTo(c.P0.X, c.P0.Y, c.P1.X, c.P1.Y, c.P2.X, c.P2.Y)
			}
			current = prev
			// outer-loop top-of-iteration check (line above the switch)
			// will exit when remaining drops to 0.
		case CmdClose:
			d := current.Distance(subStart)
			if d <= remaining {
				out.Close()
				remaining -= d
				current = subStart
			} else {
				t := remaining / d
				p := current.Lerp(subStart, t)
				out.LineTo(p.X, p.Y)
				remaining = 0
				current = p
			}
		}
	}
	_ = emittedMove
	return out
}

func cubicBezier(p0, p1, p2, p3 Point, t float64) Point {
	u := 1 - t
	uu := u * u
	uuu := uu * u
	tt := t * t
	ttt := tt * t
	return Point{
		X: uuu*p0.X + 3*uu*t*p1.X + 3*u*tt*p2.X + ttt*p3.X,
		Y: uuu*p0.Y + 3*uu*t*p1.Y + 3*u*tt*p2.Y + ttt*p3.Y,
	}
}

// subdivCubic is the curve up to t — start point + new c1,c2,end.
type subdivCubic struct {
	c1, c2, end Point
}

func subdivideCubic(p0, p1, p2, p3 Point, t float64) subdivCubic {
	// de Casteljau
	q0 := p0.Lerp(p1, t)
	q1 := p1.Lerp(p2, t)
	q2 := p2.Lerp(p3, t)
	r0 := q0.Lerp(q1, t)
	r1 := q1.Lerp(q2, t)
	end := r0.Lerp(r1, t)
	return subdivCubic{c1: q0, c2: r0, end: end}
}

// PointAlongPath returns the point at `length` units along p
// (length-parameterized).
func PointAlongPath(p *Path, length float64) Point {
	if length <= 0 {
		// Return first MoveTo point if present.
		for _, c := range p.Cmds {
			if c.Kind == CmdMove {
				return c.P0
			}
		}
		return Point{}
	}
	const curveSubdiv = 64
	remaining := length
	var current Point
	var subStart Point
	for _, c := range p.Cmds {
		switch c.Kind {
		case CmdMove:
			current = c.P0
			subStart = c.P0
		case CmdLine:
			d := current.Distance(c.P0)
			if d >= remaining {
				return current.Lerp(c.P0, remaining/math.Max(d, 1e-9))
			}
			remaining -= d
			current = c.P0
		case CmdCurve:
			prev := current
			for i := 1; i <= curveSubdiv; i++ {
				t := float64(i) / float64(curveSubdiv)
				next := cubicBezier(current, c.P0, c.P1, c.P2, t)
				step := prev.Distance(next)
				if step >= remaining {
					return prev.Lerp(next, remaining/math.Max(step, 1e-9))
				}
				remaining -= step
				prev = next
			}
			current = prev
		case CmdClose:
			d := current.Distance(subStart)
			if d >= remaining {
				return current.Lerp(subStart, remaining/math.Max(d, 1e-9))
			}
			remaining -= d
			current = subStart
		}
	}
	return current
}
