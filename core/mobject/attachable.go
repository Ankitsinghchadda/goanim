package mobject

import "github.com/ankitsinghchadda/goanim/core/geometry"

// Side names one of the four cardinal edges of a mobject's bounds.
// Used by arrow routing to compute attachment points.
type Side uint8

const (
	SideAuto Side = iota // routing picks the side
	SideTop
	SideRight
	SideBottom
	SideLeft
)

// Attachable is implemented by mobjects an arrow can connect to. The
// method returns the point on the mobject's boundary corresponding to
// the named side.
//
// Semantics:
//
//   - For rectangular shapes, AttachmentPoint(side) is the midpoint of
//     that bounding-box edge.
//   - For curved shapes (ellipses, cylinders), it's the intersection of
//     the shape boundary with a line from the center toward the
//     bounding-box-edge midpoint.
//   - For groups and containers, it's the midpoint of the bounding-box
//     edge — the group is treated as a rectangle.
type Attachable interface {
	AttachmentPoint(side Side) geometry.Point
}

// AttachToBoundsEdge is the exported form of boundsEdgeMidpoint, for
// implementations of AttachmentPoint outside this package.
func AttachToBoundsEdge(r geometry.Rect, side Side) geometry.Point {
	return boundsEdgeMidpoint(r, side)
}

// AttachToEllipse is the exported form of ellipseEdgeFromSide, for
// implementations of AttachmentPoint outside this package.
func AttachToEllipse(cx, cy, rx, ry float64, side Side) geometry.Point {
	return ellipseEdgeFromSide(cx, cy, rx, ry, side)
}

// boundsEdgeMidpoint returns the midpoint of the named edge of r.
// SideAuto is treated as the center.
func boundsEdgeMidpoint(r geometry.Rect, side Side) geometry.Point {
	cx, cy := (r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2
	switch side {
	case SideTop:
		return geometry.Point{X: cx, Y: r.Max.Y}
	case SideRight:
		return geometry.Point{X: r.Max.X, Y: cy}
	case SideBottom:
		return geometry.Point{X: cx, Y: r.Min.Y}
	case SideLeft:
		return geometry.Point{X: r.Min.X, Y: cy}
	}
	return geometry.Point{X: cx, Y: cy}
}

// ellipseEdgeFromSide computes the intersection of an ellipse at
// (cx, cy) with radii (rx, ry) and the line from the center toward the
// bounding-box-edge midpoint named by side.
func ellipseEdgeFromSide(cx, cy, rx, ry float64, side Side) geometry.Point {
	target := boundsEdgeMidpoint(
		geometry.RectFromCenter(geometry.Pt(cx, cy), rx*2, ry*2),
		side,
	)
	dx := target.X - cx
	dy := target.Y - cy
	if dx == 0 && dy == 0 {
		return geometry.Pt(cx, cy)
	}
	// Parametric form: scale so that (dx/rx, dy/ry) hits the unit circle.
	scale := 1 / sqrt((dx/rx)*(dx/rx)+(dy/ry)*(dy/ry))
	return geometry.Pt(cx+dx*scale, cy+dy*scale)
}

// Local sqrt to avoid importing math from this small helper file.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// Newton's method, 6 iterations — plenty for arrow geometry.
	z := x
	for i := 0; i < 6; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

// AutoAttach picks the dominant axis between a's center and target and
// returns the matching Side from a's perspective. Useful for arrows
// that want auto edge selection.
//
//	|dx| > |dy|, dx > 0 → SideRight
//	|dx| > |dy|, dx < 0 → SideLeft
//	|dy| > |dx|, dy > 0 → SideTop
//	|dy| > |dx|, dy < 0 → SideBottom
//
// Y-up convention: dy > 0 means target is above a.
func AutoAttach(from, to geometry.Point) Side {
	dx := to.X - from.X
	dy := to.Y - from.Y
	if absLocal(dx) > absLocal(dy) {
		if dx >= 0 {
			return SideRight
		}
		return SideLeft
	}
	if dy >= 0 {
		return SideTop
	}
	return SideBottom
}

// FlipSide returns the side opposite s. SideRight ↔ SideLeft, etc.
// Used by auto-routing: if A's exit is SideRight, B's entry should be
// SideLeft.
func FlipSide(s Side) Side {
	switch s {
	case SideTop:
		return SideBottom
	case SideBottom:
		return SideTop
	case SideLeft:
		return SideRight
	case SideRight:
		return SideLeft
	}
	return SideAuto
}

// SideNormal returns the outward unit-normal vector for s, in user
// space (Y-up). Used by curved-routing to set Bézier control handles.
func SideNormal(s Side) geometry.Point {
	switch s {
	case SideTop:
		return geometry.Pt(0, 1)
	case SideBottom:
		return geometry.Pt(0, -1)
	case SideLeft:
		return geometry.Pt(-1, 0)
	case SideRight:
		return geometry.Pt(1, 0)
	}
	return geometry.Pt(0, 0)
}

func absLocal(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
