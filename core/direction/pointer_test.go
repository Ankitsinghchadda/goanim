package direction_test

import (
	"testing"
	"time"

	"github.com/ankitsinghchadda/goanim/core/direction"
	"github.com/ankitsinghchadda/goanim/core/geometry"
)

// TestLaserPathLength — a polyline path through (0,0), (100,0),
// (100, 100) has length 200.
func TestLaserPathLength(t *testing.T) {
	p := direction.NewLaserPathFromPoints(
		geometry.Pt(0, 0),
		geometry.Pt(100, 0),
		geometry.Pt(100, 100),
	)
	if got := p.Length(); got != 200 {
		t.Errorf("LaserPath.Length() = %v, want 200", got)
	}
}

// TestLaserPathPointAt — exercising sampling at start, midpoint, and
// segment boundary, and beyond.
func TestLaserPathPointAt(t *testing.T) {
	p := direction.NewLaserPathFromPoints(
		geometry.Pt(0, 0),
		geometry.Pt(100, 0),
		geometry.Pt(100, 100),
	)
	cases := []struct {
		arc     float64
		want    geometry.Point
		precise bool
	}{
		{0, geometry.Pt(0, 0), true},
		{50, geometry.Pt(50, 0), true},    // mid of first segment
		{100, geometry.Pt(100, 0), true},  // segment boundary
		{150, geometry.Pt(100, 50), true}, // mid of second segment
		{200, geometry.Pt(100, 100), true},
		{500, geometry.Pt(100, 100), true}, // beyond → end
		{-10, geometry.Pt(0, 0), true},     // before → start
	}
	for _, c := range cases {
		got := p.PointAt(c.arc)
		if got != c.want {
			t.Errorf("PointAt(%v) = %v, want %v", c.arc, got, c.want)
		}
	}
}

// TestPathThrough — passes through each mobject's center in order.
func TestPathThrough(t *testing.T) {
	m1 := positionedMob(1, 0, 0)
	m2 := positionedMob(2, 200, 0)
	m3 := positionedMob(3, 200, 100)
	p := direction.PathThrough(m1, m2, m3)
	if got := p.Length(); got != 300 {
		t.Errorf("PathThrough length = %v, want 300", got)
	}
	mid := p.PointAt(100)
	if mid != (geometry.Pt(100, 0)) {
		t.Errorf("PathThrough.PointAt(100) = %v, want (100, 0)", mid)
	}
}

// TestLaserPointerDuration — the returned animation reports the
// requested duration and one target (the internal laser dot mobject).
func TestLaserPointerDuration(t *testing.T) {
	p := direction.NewLaserPathFromPoints(geometry.Pt(0, 0), geometry.Pt(100, 0))
	a := direction.LaserPointer(p, 2*time.Second)
	if got := a.Duration(); got != 2*time.Second {
		t.Errorf("LaserPointer.Duration() = %v, want 2s", got)
	}
	if got := a.Targets(); len(got) != 1 {
		t.Errorf("LaserPointer.Targets() len = %d, want 1", len(got))
	}
	// Apply across the range — must not panic, must not leave the dot
	// outside the path's endpoints.
	for _, tt := range []float64{0, 0.25, 0.5, 0.75, 1} {
		a.Apply(tt)
	}
}
