package rough

import (
	"testing"

	"github.com/ankitsinghchadda/goanim/core/geometry"
)

func samePath(t *testing.T, a, b *geometry.Path) {
	t.Helper()
	if len(a.Cmds) != len(b.Cmds) {
		t.Fatalf("cmd count differs: %d vs %d", len(a.Cmds), len(b.Cmds))
	}
	for i, c := range a.Cmds {
		d := b.Cmds[i]
		if c != d {
			t.Fatalf("cmd %d differs: %#v vs %#v", i, c, d)
		}
	}
}

// TestLineDeterminism — same seed, same options, identical output.
func TestLineDeterminism(t *testing.T) {
	o := DefaultOptions()
	o.Seed = 12345
	p1 := RoughLine(geometry.Pt(-100, 0), geometry.Pt(100, 0), o)
	p2 := RoughLine(geometry.Pt(-100, 0), geometry.Pt(100, 0), o)
	samePath(t, p1, p2)
}

func TestLineDifferentSeeds(t *testing.T) {
	o := DefaultOptions()
	o.Seed = 1
	p1 := RoughLine(geometry.Pt(-100, 0), geometry.Pt(100, 0), o)
	o.Seed = 2
	p2 := RoughLine(geometry.Pt(-100, 0), geometry.Pt(100, 0), o)
	// They should NOT be identical.
	identical := true
	if len(p1.Cmds) != len(p2.Cmds) {
		identical = false
	} else {
		for i, c := range p1.Cmds {
			if c != p2.Cmds[i] {
				identical = false
				break
			}
		}
	}
	if identical {
		t.Fatalf("expected different output for different seeds; got identical")
	}
}

func TestRectangleDeterminism(t *testing.T) {
	o := DefaultOptions()
	o.Seed = 99
	a := RoughRectangle(-50, -50, 100, 100, o)
	b := RoughRectangle(-50, -50, 100, 100, o)
	samePath(t, a, b)
}

func TestPolygonDeterminism(t *testing.T) {
	o := DefaultOptions()
	o.Seed = 42
	pts := []geometry.Point{{X: 0, Y: 0}, {X: 100, Y: 0}, {X: 50, Y: 87}}
	a := RoughPolygon(pts, o)
	b := RoughPolygon(pts, o)
	samePath(t, a, b)
}

func TestEllipseDeterminism(t *testing.T) {
	o := DefaultOptions()
	o.Seed = 77
	a := RoughEllipse(0, 0, 80, 50, o)
	b := RoughEllipse(0, 0, 80, 50, o)
	samePath(t, a, b)
}

func TestHatchDeterminism(t *testing.T) {
	o := DefaultOptions()
	o.Seed = 7
	o.HachureGap = 8
	poly := RectToPolygon(0, 0, 100, 60)
	a := Hatch(poly, o)
	b := Hatch(poly, o)
	samePath(t, a, b)
}

func TestZigzagDeterminism(t *testing.T) {
	o := DefaultOptions()
	o.Seed = 11
	o.HachureGap = 8
	poly := RectToPolygon(0, 0, 100, 60)
	a := Zigzag(poly, o)
	b := Zigzag(poly, o)
	samePath(t, a, b)
}

// TestRNGMatchesLehmer — sanity check: known seeded PRNG output.
func TestRNGMatchesLehmer(t *testing.T) {
	r := newRNG(1)
	// Lehmer (Park-Miller) seed=1, multiplier=48271 mod 2^32:
	// state_1 = 48271, state_2 = 48271*48271 mod 2^32 (int32 overflow).
	got1 := r.next()
	want1 := float64(48271) / float64(1<<31)
	if got1 != want1 {
		t.Fatalf("first draw: got %v want %v", got1, want1)
	}
}
