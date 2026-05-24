package systemdesign

import (
	"math"
	"testing"

	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// TestOrthogonalLabelArcLengthMidpoint constructs an L-shaped
// orthogonal arrow (horizontal 100, vertical 200) and asserts the
// label lands at the arc-length midpoint — 150 units from the start,
// 50 units down the vertical segment. The bug this guards against is
// placing the label at the chord midpoint, which would put it floating
// in empty space halfway between the endpoints rather than on the path.
func TestOrthogonalLabelArcLengthMidpoint(t *testing.T) {
	// Two rectangles positioned to force a single-bend orthogonal path
	// from source's right edge to target's top edge. The bend lands at
	// (target.X, source.Y). With source at (0, 0) of size 20×20 and
	// target at (100, -200) of size 20×20, after the shrink (6px) the
	// effective endpoints are roughly (16, 0) and (100, -184) — long
	// enough that no segment collapses to a straight line.
	src := mobject.NewRectangle(1, 20, 20).MoveTo(0, 0)
	dst := mobject.NewRectangle(2, 20, 20).MoveTo(100, -200)
	a := NewArrow(3, src, dst).
		WithRouting(RoutingOrthogonal).
		From(mobject.SideRight).
		To(mobject.SideTop)

	segs := a.pathSegments()
	if len(segs) != 2 {
		t.Fatalf("expected 2 orthogonal segments, got %d", len(segs))
	}

	// Compute total arc length and where the midpoint lands.
	var total float64
	for _, s := range segs {
		total += s[0].Distance(s[1])
	}

	pt := a.labelPositionPoint()

	// The midpoint must lie ON one of the segments. We allow a small
	// epsilon for shrink/rounding. Specifically, with the L-shape and
	// shrink, the midpoint sits on the vertical (second) segment if
	// the horizontal segment is shorter than half the total length,
	// or on the horizontal (first) segment otherwise.
	hSegLen := segs[0][0].Distance(segs[0][1])
	vSegLen := segs[1][0].Distance(segs[1][1])

	if math.Abs(total-(hSegLen+vSegLen)) > 1e-6 {
		t.Fatalf("segment lengths inconsistent with total: total=%f, h+v=%f", total, hSegLen+vSegLen)
	}

	// For a 100-by-200 L-shape, vertical >> horizontal, so the midpoint
	// is on the vertical segment.
	if hSegLen >= total/2 {
		t.Fatalf("expected vertical segment to contain midpoint; horizontal=%f, total=%f", hSegLen, total)
	}

	// Midpoint should sit on the second (vertical) segment, with
	// |pt.X - bend.X| ≈ 0 (within a couple px of rounding).
	bend := segs[0][1]
	if math.Abs(pt.X-bend.X) > 1.5 {
		t.Fatalf("label X drifted off vertical segment: pt.X=%f, bend.X=%f", pt.X, bend.X)
	}

	// Midpoint Y should equal bend.Y minus (total/2 - hSegLen), since
	// we descend from the bend by (target - hSegLen) along the vertical.
	wantY := bend.Y - (total/2 - hSegLen)
	if math.Abs(pt.Y-wantY) > 1.5 {
		t.Fatalf("label Y wrong: got %f, want %f (bend.Y=%f, hSegLen=%f, total/2=%f)",
			pt.Y, wantY, bend.Y, hSegLen, total/2)
	}
}

// TestOrthogonalLabelArcLengthMidpointTwoBend constructs a Z-shape
// (two-bend) arrow and asserts the label lands on the middle (cross)
// segment, not floating between the endpoints.
func TestOrthogonalLabelArcLengthMidpointTwoBend(t *testing.T) {
	// Source: vertical exit (bottom side). Target: vertical entry (top
	// side). Both-vertical → two-bend Z route: down, across, down.
	src := mobject.NewRectangle(1, 20, 20).MoveTo(0, 0)
	dst := mobject.NewRectangle(2, 20, 20).MoveTo(200, -300)
	a := NewArrow(3, src, dst).
		WithRouting(RoutingOrthogonal).
		From(mobject.SideBottom).
		To(mobject.SideTop)

	segs := a.pathSegments()
	if len(segs) != 3 {
		t.Fatalf("expected 3 segments for two-bend route, got %d", len(segs))
	}

	pt := a.labelPositionPoint()

	// For a Z-shape from (0,0) to (200,-300), segments are:
	//   p1→b1: vertical down ~140
	//   b1→b2: horizontal across 200 (the middle segment)
	//   b2→p2: vertical down ~140
	// Total ~480. Midpoint at ~240, which falls inside the middle
	// horizontal segment (starts at acc=140, ends at acc=340).
	mid := segs[1]
	// Label Y should equal the horizontal segment's Y (i.e. the midpoint
	// of source and target Y).
	if math.Abs(pt.Y-mid[0].Y) > 1e-6 {
		t.Fatalf("label not on cross segment: pt.Y=%f, mid.Y=%f", pt.Y, mid[0].Y)
	}
	// X should be somewhere between the segment's endpoints.
	xMin := math.Min(mid[0].X, mid[1].X)
	xMax := math.Max(mid[0].X, mid[1].X)
	if pt.X < xMin-1 || pt.X > xMax+1 {
		t.Fatalf("label X off cross segment: pt.X=%f, range=[%f, %f]", pt.X, xMin, xMax)
	}
}
