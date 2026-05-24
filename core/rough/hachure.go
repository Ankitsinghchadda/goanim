package rough

import (
	"math"
	"sort"

	"github.com/ankitsinghchadda/goanim/core/geometry"
)

// hachureLines computes parallel scan-fill lines through a closed
// polygon at the requested angle and spacing. Lines are returned in
// world (unrotated) coordinates. The algorithm is the classic active-
// edge-table polygon scan-fill, generalized over a rotation.
//
// angleDeg is the orientation of the hachure lines; gap is the spacing.
// Internally we rotate the polygon by angleDeg+90° so that the hachure
// direction becomes horizontal (scan lines are parallel to the X axis),
// run the scan fill, and rotate the resulting line segments back.
func hachureLines(polygon []geometry.Point, angleDeg, gap float64) [][2]geometry.Point {
	if gap <= 0 {
		gap = 0.1
	}
	rot := (angleDeg + 90) * math.Pi / 180
	cos, sin := math.Cos(rot), math.Sin(rot)
	rotPts := make([]geometry.Point, len(polygon))
	for i, p := range polygon {
		rotPts[i] = geometry.Pt(p.X*cos-p.Y*sin, p.X*sin+p.Y*cos)
	}

	lines := straightHachureLines(rotPts, gap)

	// Rotate result lines back by -rot.
	cosI, sinI := math.Cos(-rot), math.Sin(-rot)
	out := make([][2]geometry.Point, len(lines))
	for i, l := range lines {
		out[i][0] = geometry.Pt(l[0].X*cosI-l[0].Y*sinI, l[0].X*sinI+l[0].Y*cosI)
		out[i][1] = geometry.Pt(l[1].X*cosI-l[1].Y*sinI, l[1].X*sinI+l[1].Y*cosI)
	}
	return out
}

type edge struct {
	yMin, yMax float64
	x          float64 // current x at scan-line y
	islope     float64 // dx/dy
}

// straightHachureLines runs an integer-stepping active-edge-table scan
// fill across the polygon in its current orientation (hachure direction
// = horizontal). Returns horizontal segments.
//
// rough.js uses gap as the visual spacing and steps y by 1, emitting a
// line every `gap` iterations — that yields effective vertical spacing
// of `gap`. We do the same.
func straightHachureLines(polygon []geometry.Point, gap float64) [][2]geometry.Point {
	pts := polygon
	if len(pts) == 0 {
		return nil
	}
	// Ensure closed (rough.js appends p0 if missing).
	if pts[0] != pts[len(pts)-1] {
		pts = append(append([]geometry.Point{}, pts...), pts[0])
	}
	var edges []edge
	for i := 0; i+1 < len(pts); i++ {
		p1, p2 := pts[i], pts[i+1]
		if p1.Y == p2.Y {
			continue
		}
		yMin := math.Min(p1.Y, p2.Y)
		yMax := math.Max(p1.Y, p2.Y)
		var x float64
		if yMin == p1.Y {
			x = p1.X
		} else {
			x = p2.X
		}
		edges = append(edges, edge{
			yMin: yMin, yMax: yMax, x: x,
			islope: (p2.X - p1.X) / (p2.Y - p1.Y),
		})
	}
	if len(edges) == 0 {
		return nil
	}
	sort.SliceStable(edges, func(i, j int) bool {
		if edges[i].yMin != edges[j].yMin {
			return edges[i].yMin < edges[j].yMin
		}
		if edges[i].x != edges[j].x {
			return edges[i].x < edges[j].x
		}
		return edges[i].yMax < edges[j].yMax
	})

	var active []edge
	var lines [][2]geometry.Point
	y := edges[0].yMin
	iteration := 0
	// Step y by 1 unit at a time; emit every gap iterations.
	gapInt := int(math.Round(math.Max(gap, 1)))
	if gapInt < 1 {
		gapInt = 1
	}
	for {
		// Promote edges whose yMin <= y.
		for len(edges) > 0 && edges[0].yMin <= y {
			active = append(active, edges[0])
			edges = edges[1:]
		}
		// Drop edges whose yMax <= y.
		active = filterEdges(active, y)
		if len(active) == 0 && len(edges) == 0 {
			break
		}
		// Sort active by current x.
		sort.SliceStable(active, func(i, j int) bool {
			return active[i].x < active[j].x
		})
		if iteration%gapInt == 0 {
			for i := 0; i+1 < len(active); i += 2 {
				lines = append(lines, [2]geometry.Point{
					{X: active[i].x, Y: y},
					{X: active[i+1].x, Y: y},
				})
			}
		}
		y++
		for i := range active {
			active[i].x += active[i].islope
		}
		iteration++
	}
	return lines
}

func filterEdges(active []edge, y float64) []edge {
	out := active[:0]
	for _, e := range active {
		if e.yMax > y {
			out = append(out, e)
		}
	}
	return out
}

// Hatch emits a hachure (parallel sketched lines) fill of polygon.
// The fill lines are themselves drawn with rough.js-style double strokes
// when opts.DisableMultiStrokeFill is false, so the fill has the same
// hand-drawn quality as the outline.
func Hatch(polygon []geometry.Point, opts Options) *geometry.Path {
	gap := opts.resolvedHachureGap()
	lines := hachureLines(polygon, opts.HachureAngle, gap)
	r := newRNG(opts.Seed)
	out := geometry.NewPath()
	for _, l := range lines {
		roughLineInto(out, l[0], l[1], opts, r, false)
		if !opts.DisableMultiStrokeFill {
			roughLineInto(out, l[0], l[1], opts, r, true)
		}
	}
	return out
}

// CrossHatch emits a hatched fill at two perpendicular angles, giving
// a denser "scribble" look.
func CrossHatch(polygon []geometry.Point, opts Options) *geometry.Path {
	a := Hatch(polygon, opts)
	o2 := opts
	o2.HachureAngle = opts.HachureAngle + 90
	o2.Seed = opts.Seed + 7
	b := Hatch(polygon, o2)
	a.Append(b)
	return a
}

// Zigzag emits a zigzag fill: parallel hachure scan lines whose
// start-points are offset by ±gap/2 along the hachure direction,
// producing the V-shaped pattern.
func Zigzag(polygon []geometry.Point, opts Options) *geometry.Path {
	gap := opts.resolvedHachureGap()
	if gap <= 0 {
		gap = 0.1
	}
	lines := hachureLines(polygon, opts.HachureAngle, gap)

	rad := opts.HachureAngle * math.Pi / 180
	dgx := gap * 0.5 * math.Cos(rad)
	dgy := gap * 0.5 * math.Sin(rad)

	r := newRNG(opts.Seed)
	out := geometry.NewPath()
	for _, l := range lines {
		if l[0].Distance(l[1]) == 0 {
			continue
		}
		up := geometry.Pt(l[0].X-dgx, l[0].Y+dgy)
		down := geometry.Pt(l[0].X+dgx, l[0].Y-dgy)
		roughLineInto(out, up, l[1], opts, r, false)
		roughLineInto(out, down, l[1], opts, r, false)
	}
	return out
}

// Dots emits a grid of small filled circles inside the polygon, jittered
// by the seeded PRNG so they look hand-stippled rather than mechanical.
//
// The result is a Path containing one closed cubic-Bezier circle per dot;
// render with a solid Fill (no stroke) for the proper stippled look.
func Dots(polygon []geometry.Point, opts Options) *geometry.Path {
	if len(polygon) < 3 {
		return geometry.NewPath()
	}
	bb := geometry.RectFromPoints(polygon...)
	gap := opts.resolvedHachureGap()
	if gap < 4 {
		gap = 4
	}
	radius := math.Max(opts.StrokeWidth*0.7, 1.2)

	r := newRNG(opts.Seed)
	out := geometry.NewPath()

	// Walk a grid covering bb, jitter each cell's center, and emit a
	// circle if the jittered point is inside the polygon.
	for y := bb.Min.Y + gap/2; y < bb.Max.Y; y += gap {
		for x := bb.Min.X + gap/2; x < bb.Max.X; x += gap {
			jx := x + offsetSym(gap*0.25, opts.Roughness, 1, r)
			jy := y + offsetSym(gap*0.25, opts.Roughness, 1, r)
			if !pointInPolygon(geometry.Pt(jx, jy), polygon) {
				continue
			}
			out.Append(geometry.EllipsePath(jx, jy, radius, radius))
		}
	}
	return out
}

// pointInPolygon — standard even-odd ray-cast test.
func pointInPolygon(p geometry.Point, poly []geometry.Point) bool {
	inside := false
	n := len(poly)
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		yi, yj := poly[i].Y, poly[j].Y
		if (yi > p.Y) == (yj > p.Y) {
			continue
		}
		xCross := (poly[j].X-poly[i].X)*(p.Y-yi)/(yj-yi) + poly[i].X
		if p.X < xCross {
			inside = !inside
		}
	}
	return inside
}

// SolidFill returns a closed-polygon Path that, when rendered with a
// non-nil Fill, gives a flat-color shape. No roughness is applied to
// the outline so the fill stays inside the visual shape; the rough
// stroke is meant to be drawn on top.
func SolidFill(polygon []geometry.Point) *geometry.Path {
	if len(polygon) < 3 {
		return geometry.NewPath()
	}
	p := geometry.NewPath()
	p.MoveTo(polygon[0].X, polygon[0].Y)
	for _, pt := range polygon[1:] {
		p.LineTo(pt.X, pt.Y)
	}
	p.Close()
	return p
}

// EllipseToPolygon samples an ellipse outline into a polygon suitable
// for hachure fills. The sampling is fine enough that small fill gaps
// near the boundary aren't visually noticeable.
func EllipseToPolygon(cx, cy, rx, ry float64, steps int) []geometry.Point {
	if steps < 8 {
		steps = 8
	}
	pts := make([]geometry.Point, steps)
	for i := 0; i < steps; i++ {
		a := 2 * math.Pi * float64(i) / float64(steps)
		pts[i] = geometry.Pt(cx+rx*math.Cos(a), cy+ry*math.Sin(a))
	}
	return pts
}

// RectToPolygon converts the four corners of a rectangle to a polygon.
func RectToPolygon(x, y, w, h float64) []geometry.Point {
	return []geometry.Point{
		{X: x, Y: y},
		{X: x + w, Y: y},
		{X: x + w, Y: y + h},
		{X: x, Y: y + h},
	}
}
