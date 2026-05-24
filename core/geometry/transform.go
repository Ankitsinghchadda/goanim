package geometry

import "math"

// Transform is a 2D affine transform stored as a 3x2 matrix:
//
//	| A C E |
//	| B D F |
//	| 0 0 1 |
//
// A point (x, y) maps to (A*x + C*y + E, B*x + D*y + F).
type Transform struct {
	A, B, C, D, E, F float64
}

// Identity returns the identity transform.
func Identity() Transform {
	return Transform{A: 1, D: 1}
}

// Translate returns a translation by (dx, dy).
func Translate(dx, dy float64) Transform {
	return Transform{A: 1, D: 1, E: dx, F: dy}
}

// Scale returns a uniform scale by s (or non-uniform if sy differs).
// Scale(sx) scales both axes by sx; Scale(sx, sy) scales independently.
func Scale(sx float64, sy ...float64) Transform {
	s := sx
	if len(sy) > 0 {
		s = sy[0]
	}
	return Transform{A: sx, D: s}
}

// Rotate returns a counter-clockwise rotation by radians (in math convention,
// Y-up). To rotate around a non-origin point, compose Translate(c)·Rotate·Translate(-c).
func Rotate(radians float64) Transform {
	c, s := math.Cos(radians), math.Sin(radians)
	return Transform{A: c, B: s, C: -s, D: c}
}

// Compose returns t·u, i.e. the transform that applies u first, then t.
func (t Transform) Compose(u Transform) Transform {
	return Transform{
		A: t.A*u.A + t.C*u.B,
		B: t.B*u.A + t.D*u.B,
		C: t.A*u.C + t.C*u.D,
		D: t.B*u.C + t.D*u.D,
		E: t.A*u.E + t.C*u.F + t.E,
		F: t.B*u.E + t.D*u.F + t.F,
	}
}

// Apply transforms a point.
func (t Transform) Apply(p Point) Point {
	return Point{
		X: t.A*p.X + t.C*p.Y + t.E,
		Y: t.B*p.X + t.D*p.Y + t.F,
	}
}

// ApplyVector applies the linear part of the transform to a vector (no translation).
func (t Transform) ApplyVector(v Point) Point {
	return Point{
		X: t.A*v.X + t.C*v.Y,
		Y: t.B*v.X + t.D*v.Y,
	}
}
