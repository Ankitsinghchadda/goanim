package mobject

import (
	"math"
	"testing"

	"github.com/ankitsinghchadda/goanim/core/geometry"
)

func nearMobj(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// TestRectangleAttachment — midpoints of each edge.
func TestRectangleAttachment(t *testing.T) {
	r := NewRectangle(1, 200, 100).MoveTo(50, 30)

	cases := []struct {
		side        Side
		x, y        float64
		description string
	}{
		{SideTop, 50, 80, "top"},
		{SideRight, 150, 30, "right"},
		{SideBottom, 50, -20, "bottom"},
		{SideLeft, -50, 30, "left"},
	}
	for _, c := range cases {
		p := r.AttachmentPoint(c.side)
		if !nearMobj(p.X, c.x) || !nearMobj(p.Y, c.y) {
			t.Errorf("%s edge: got (%v, %v), want (%v, %v)", c.description, p.X, p.Y, c.x, c.y)
		}
	}
}

// TestEllipseAttachment — points must lie on the ellipse boundary.
func TestEllipseAttachment(t *testing.T) {
	e := NewEllipse(1, 100, 50).MoveTo(0, 0)
	p := e.AttachmentPoint(SideRight)
	if !nearMobj(p.X, 100) || !nearMobj(p.Y, 0) {
		t.Errorf("right: got (%v, %v), want (100, 0)", p.X, p.Y)
	}
	p = e.AttachmentPoint(SideTop)
	if !nearMobj(p.X, 0) || !nearMobj(p.Y, 50) {
		t.Errorf("top: got (%v, %v), want (0, 50)", p.X, p.Y)
	}
}

func TestAutoAttach(t *testing.T) {
	cases := []struct {
		dx, dy float64
		want   Side
	}{
		{100, 0, SideRight},
		{-100, 0, SideLeft},
		{0, 100, SideTop},
		{0, -100, SideBottom},
		{200, 50, SideRight},   // dominant horizontal +
		{-200, 50, SideLeft},   // dominant horizontal -
		{50, 200, SideTop},     // dominant vertical +
		{50, -200, SideBottom}, // dominant vertical -
	}
	for _, c := range cases {
		got := AutoAttach(geometry.Pt(0, 0), geometry.Pt(c.dx, c.dy))
		if got != c.want {
			t.Errorf("AutoAttach(%v,%v): got %v, want %v", c.dx, c.dy, got, c.want)
		}
	}
}
