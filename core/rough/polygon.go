package rough

import "github.com/ankitsinghchadda/goanim/core/geometry"

// RoughRectangle returns a rough-style axis-aligned rectangle with its
// lower-left corner at (x, y) and the given width and height.
//
// Each of the four edges is drawn as an independent RoughLine, so the
// corners visibly disagree — which is exactly the Excalidraw "corners
// don't quite meet" signature.
func RoughRectangle(x, y, w, h float64, opts Options) *geometry.Path {
	pts := []geometry.Point{
		{X: x, Y: y},
		{X: x + w, Y: y},
		{X: x + w, Y: y + h},
		{X: x, Y: y + h},
	}
	return RoughPolygon(pts, opts)
}

// RoughRectangleCentered returns a rough rectangle centered at (cx, cy).
func RoughRectangleCentered(cx, cy, w, h float64, opts Options) *geometry.Path {
	return RoughRectangle(cx-w/2, cy-h/2, w, h, opts)
}

// RoughPolygon returns a rough-style closed polygon through pts.
// Edges are drawn in order; the polygon is closed by drawing a final
// edge from pts[len-1] back to pts[0].
func RoughPolygon(pts []geometry.Point, opts Options) *geometry.Path {
	if len(pts) < 2 {
		return geometry.NewPath()
	}
	r := newRNG(opts.Seed)
	out := geometry.NewPath()
	for i := 0; i < len(pts); i++ {
		p1 := pts[i]
		p2 := pts[(i+1)%len(pts)]
		roughLineInto(out, p1, p2, opts, r, false)
		if !opts.DisableMultiStroke {
			roughLineInto(out, p1, p2, opts, r, true)
		}
	}
	return out
}

// RoughPolyline returns a rough-style open polyline through pts (no
// closing edge).
func RoughPolyline(pts []geometry.Point, opts Options) *geometry.Path {
	if len(pts) < 2 {
		return geometry.NewPath()
	}
	r := newRNG(opts.Seed)
	out := geometry.NewPath()
	for i := 0; i < len(pts)-1; i++ {
		p1 := pts[i]
		p2 := pts[i+1]
		roughLineInto(out, p1, p2, opts, r, false)
		if !opts.DisableMultiStroke {
			roughLineInto(out, p1, p2, opts, r, true)
		}
	}
	return out
}
