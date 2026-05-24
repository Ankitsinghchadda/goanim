package layout

import (
	"math"
	"testing"

	"github.com/ankitsinghchadda/goanim/core/mobject"
)

func nearly(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// TestHBoxPositioning — three rectangles of different sizes laid out
// in a row with known spacing should produce known centers.
func TestHBoxPositioning(t *testing.T) {
	a := mobject.NewRectangle(1, 100, 60)
	b := mobject.NewRectangle(2, 80, 40)
	c := mobject.NewRectangle(3, 120, 50)

	hb := NewHBox(a, b, c).WithSpacing(20).MoveTo(0, 0)
	hb.Layout()

	// Total width = 100 + 80 + 120 + 2*20 = 340. Center at x=0 → left edge -170.
	// a center: -170 + 50 = -120
	// b center: -170 + 100 + 20 + 40 = -10
	// c center: -170 + 100 + 20 + 80 + 20 + 60 = 110
	wantCenters := []struct{ x, y float64 }{
		{-120, 0}, {-10, 0}, {110, 0},
	}
	for i, w := range wantCenters {
		got := []mobject.Mobject{a, b, c}[i]
		ax, ay := got.(*mobject.Rectangle).Position()
		if !nearly(ax, w.x) || !nearly(ay, w.y) {
			t.Errorf("child %d: got (%v, %v), want (%v, %v)", i, ax, ay, w.x, w.y)
		}
	}
}

// TestVBoxPositioning — Y-up convention: first child at the top.
func TestVBoxPositioning(t *testing.T) {
	a := mobject.NewRectangle(1, 100, 60)
	b := mobject.NewRectangle(2, 80, 40)

	vb := NewVBox(a, b).WithSpacing(10).MoveTo(0, 0)
	vb.Layout()

	// Total height = 60 + 40 + 10 = 110. Center at y=0 → top edge at +55.
	// a (first) center: +55 - 30 = +25
	// b (second) center: +25 - 30 - 10 - 20 = -35
	if x, y := a.Position(); !nearly(x, 0) || !nearly(y, 25) {
		t.Errorf("VBox first child: got (%v, %v), want (0, 25)", x, y)
	}
	if x, y := b.Position(); !nearly(x, 0) || !nearly(y, -35) {
		t.Errorf("VBox second child: got (%v, %v), want (0, -35)", x, y)
	}
}

// TestHBoxAlign — with VTop, all children share top edge; the row's
// effective height is max(child heights), and each child's vertical
// center is set so its top matches.
func TestHBoxAlignTop(t *testing.T) {
	a := mobject.NewRectangle(1, 60, 60) // tallest
	b := mobject.NewRectangle(2, 60, 30) // half height

	hb := NewHBox(a, b).WithSpacing(0).WithAlign(VTop).MoveTo(0, 0)
	hb.Layout()

	// maxH = 60. row top = cy + maxH/2 = 30.
	// a top at 30, a center at 30-30 = 0.
	// b top at 30, b center at 30-15 = 15.
	if _, y := a.Position(); !nearly(y, 0) {
		t.Errorf("HBox top-align a: got y=%v, want 0", y)
	}
	if _, y := b.Position(); !nearly(y, 15) {
		t.Errorf("HBox top-align b: got y=%v, want 15", y)
	}
}

// TestGridPositioning — 2 rows × 2 cols of equal rectangles.
func TestGridPositioning(t *testing.T) {
	rs := []mobject.Mobject{
		mobject.NewRectangle(1, 50, 50),
		mobject.NewRectangle(2, 50, 50),
		mobject.NewRectangle(3, 50, 50),
		mobject.NewRectangle(4, 50, 50),
	}
	g := NewGrid(2, 2, rs...).WithSpacing(10, 20).MoveTo(0, 0)
	g.Layout()

	// 2x2 with 50x50 cells, col gap 20, row gap 10.
	// totalW = 50+50+20 = 120, totalH = 50+50+10 = 110.
	// Cell centers:
	//   col 0: -60 + 25 = -35
	//   col 1: -35 + 50 + 20 = 35
	//   row 0: +55 - 25 = +30
	//   row 1: +30 - 50 - 10 = -30
	wantPositions := []struct{ x, y float64 }{
		{-35, 30}, {35, 30}, {-35, -30}, {35, -30},
	}
	for i, w := range wantPositions {
		ax, ay := rs[i].(*mobject.Rectangle).Position()
		if !nearly(ax, w.x) || !nearly(ay, w.y) {
			t.Errorf("Grid cell %d: got (%v, %v), want (%v, %v)", i, ax, ay, w.x, w.y)
		}
	}
}

// TestEmptyHBox — should not crash, returns zero bounds.
func TestEmptyHBox(t *testing.T) {
	hb := NewHBox().MoveTo(5, 7)
	b := hb.Bounds()
	if b.Width() != 0 || b.Height() != 0 {
		t.Errorf("empty HBox should have zero bounds; got w=%v h=%v", b.Width(), b.Height())
	}
}

// TestNestedComposition — a VBox of HBoxes nests correctly.
func TestNestedComposition(t *testing.T) {
	a := mobject.NewRectangle(1, 40, 20)
	b := mobject.NewRectangle(2, 40, 20)
	c := mobject.NewRectangle(3, 60, 20)

	top := NewHBox(a, b).WithSpacing(10)
	root := NewVBox(top, c).WithSpacing(20).MoveTo(0, 0)
	root.Layout()

	// outerHeight = (max child height) — top is max-of-a/b = 20, c is 20.
	// totalH = 20 + 20 + 20 = 60. center 0 → top y = 30.
	// top center y = 30 - 10 = 20.
	// c center y = 20 - 10 - 20 - 10 = -20.
	if x, y := c.Position(); !nearly(x, 0) || !nearly(y, -20) {
		t.Errorf("nested VBox second child position: got (%v, %v), want (0, -20)", x, y)
	}
	// And inside `top` HBox: total width = 40 + 40 + 10 = 90, left edge -45.
	// a center: -45 + 20 = -25
	// b center: -25 + 20 + 10 + 20 = 25
	if x, y := a.Position(); !nearly(x, -25) || !nearly(y, 20) {
		t.Errorf("nested HBox first child position: got (%v, %v), want (-25, 20)", x, y)
	}
	if x, y := b.Position(); !nearly(x, 25) || !nearly(y, 20) {
		t.Errorf("nested HBox second child position: got (%v, %v), want (25, 20)", x, y)
	}
}
