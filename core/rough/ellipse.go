package rough

import (
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
)

// RoughEllipse returns a rough-style ellipse centered at (cx, cy) with
// x-radius rx and y-radius ry. The ellipse is approximated by a closed
// Catmull-Rom spline through perturbed sample points on the parametric
// circle; the result is a smooth, sketchy oval.
//
// When DisableMultiStroke is false (the default), two overlay strokes
// with different jitter magnitudes are drawn — the rough.js signature
// look for hand-drawn circles.
func RoughEllipse(cx, cy, rx, ry float64, opts Options) *geometry.Path {
	r := newRNG(opts.Seed)
	params := generateEllipseParams(rx*2, ry*2, opts, r)

	out := geometry.NewPath()
	pts1, _ := computeEllipsePoints(params.increment, cx, cy, params.rx, params.ry, 1.0, params.overlap, opts, r)
	curveInto(out, pts1, opts)

	if !opts.DisableMultiStroke && opts.Roughness != 0 {
		pts2, _ := computeEllipsePoints(params.increment, cx, cy, params.rx, params.ry, 1.5, 0, opts, r)
		curveInto(out, pts2, opts)
	}
	return out
}

type ellipseParams struct {
	increment float64
	rx, ry    float64
	overlap   float64
}

// generateEllipseParams matches rough.js's perimeter-based step count
// and randomizes the radii slightly. It also draws the "overlap"
// parameter used to make the closed spline wrap smoothly.
//
// PRNG draw order (counted from generateEllipseParams entry):
//
//  1. rx jitter
//  2. ry jitter
//  3. inner _offset(0.4, 1) draw for overlap
//  4. outer _offset(0.1, ...) draw for overlap
func generateEllipseParams(width, height float64, o Options, r *rng) ellipseParams {
	hw := width / 2
	hh := height / 2
	psq := math.Sqrt(2 * math.Pi * math.Sqrt((hw*hw+hh*hh)/2))
	stepCount := math.Ceil(math.Max(o.CurveStepCount, (o.CurveStepCount/math.Sqrt(200))*psq))
	increment := 2 * math.Pi / stepCount
	rx := math.Abs(hw)
	ry := math.Abs(hh)
	curveFitRand := 1 - o.CurveFitting

	rx += offsetSym(rx*curveFitRand, o.Roughness, 1, r)
	ry += offsetSym(ry*curveFitRand, o.Roughness, 1, r)

	// rough.js: increment * _offset(0.1, _offset(0.4, 1, o), o)
	inner := offsetRange(0.4, 1, o.Roughness, 1, r)
	overlap := increment * offsetRange(0.1, inner, o.Roughness, 1, r)
	return ellipseParams{increment: increment, rx: rx, ry: ry, overlap: overlap}
}

// computeEllipsePoints emits sample points around the ellipse, each
// independently perturbed in x and y by ±offset * roughness.
//
// The returned slice contains the full point list (with leading and
// trailing wrap-around padding so the spline closes smoothly). For each
// emitted point, the PRNG is consumed exactly twice: x-jitter, then
// y-jitter. (At the very start, one additional draw is consumed for
// radOffset.)
func computeEllipsePoints(increment, cx, cy, rx, ry, offset, overlap float64, o Options, r *rng) ([]geometry.Point, []geometry.Point) {
	var all, core []geometry.Point
	if o.Roughness == 0 {
		// Smooth path — fine sampling, no jitter.
		inc := increment / 4
		all = append(all, geometry.Pt(cx+rx*math.Cos(-inc), cy+ry*math.Sin(-inc)))
		for a := 0.0; a <= 2*math.Pi; a += inc {
			p := geometry.Pt(cx+rx*math.Cos(a), cy+ry*math.Sin(a))
			core = append(core, p)
			all = append(all, p)
		}
		all = append(all, geometry.Pt(cx+rx*math.Cos(0), cy+ry*math.Sin(0)))
		all = append(all, geometry.Pt(cx+rx*math.Cos(inc), cy+ry*math.Sin(inc)))
		return all, core
	}

	radOffset := offsetSym(0.5, o.Roughness, 1, r) - math.Pi/2

	// Leading "pre-point" at 0.9 * radius to give the closed spline a tangent.
	all = append(all, geometry.Pt(
		offsetSym(offset, o.Roughness, 1, r)+cx+0.9*rx*math.Cos(radOffset-increment),
		offsetSym(offset, o.Roughness, 1, r)+cy+0.9*ry*math.Sin(radOffset-increment),
	))
	endAngle := 2*math.Pi + radOffset - 0.01
	for a := radOffset; a < endAngle; a += increment {
		p := geometry.Pt(
			offsetSym(offset, o.Roughness, 1, r)+cx+rx*math.Cos(a),
			offsetSym(offset, o.Roughness, 1, r)+cy+ry*math.Sin(a),
		)
		core = append(core, p)
		all = append(all, p)
	}
	// Trailing wrap-around: three more points to close the spline cleanly.
	all = append(all, geometry.Pt(
		offsetSym(offset, o.Roughness, 1, r)+cx+rx*math.Cos(radOffset+2*math.Pi+overlap*0.5),
		offsetSym(offset, o.Roughness, 1, r)+cy+ry*math.Sin(radOffset+2*math.Pi+overlap*0.5),
	))
	all = append(all, geometry.Pt(
		offsetSym(offset, o.Roughness, 1, r)+cx+0.98*rx*math.Cos(radOffset+overlap),
		offsetSym(offset, o.Roughness, 1, r)+cy+0.98*ry*math.Sin(radOffset+overlap),
	))
	all = append(all, geometry.Pt(
		offsetSym(offset, o.Roughness, 1, r)+cx+0.9*rx*math.Cos(radOffset+overlap*0.5),
		offsetSym(offset, o.Roughness, 1, r)+cy+0.9*ry*math.Sin(radOffset+overlap*0.5),
	))
	return all, core
}

// curveInto appends a Catmull-Rom spline through pts to out, converted
// to cubic Bezier segments. curveTightness in opts pulls control points
// in or out: tightness 0 → standard Catmull-Rom; > 0 → straighter; < 0
// → loopier.
func curveInto(out *geometry.Path, pts []geometry.Point, opts Options) {
	n := len(pts)
	if n == 0 {
		return
	}
	if n == 2 {
		// Two points → just a line.
		out.MoveTo(pts[0].X, pts[0].Y)
		out.LineTo(pts[1].X, pts[1].Y)
		return
	}
	if n == 3 {
		out.MoveTo(pts[1].X, pts[1].Y)
		out.CurveTo(pts[1].X, pts[1].Y, pts[2].X, pts[2].Y, pts[2].X, pts[2].Y)
		return
	}
	s := 1 - opts.CurveTightness
	out.MoveTo(pts[1].X, pts[1].Y)
	for i := 1; i < n-2; i++ {
		b0 := pts[i]
		b1 := geometry.Pt(
			pts[i].X+(s*pts[i+1].X-s*pts[i-1].X)/6,
			pts[i].Y+(s*pts[i+1].Y-s*pts[i-1].Y)/6,
		)
		b2 := geometry.Pt(
			pts[i+1].X+(s*pts[i].X-s*pts[i+2].X)/6,
			pts[i+1].Y+(s*pts[i].Y-s*pts[i+2].Y)/6,
		)
		b3 := pts[i+1]
		_ = b0
		out.CurveTo(b1.X, b1.Y, b2.X, b2.Y, b3.X, b3.Y)
	}
}
