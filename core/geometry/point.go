package geometry

import "math"

// Point is a 2D point or vector in user-space coordinates.
//
// User space is center-origin: (0, 0) is the middle of the canvas and Y
// increases upward. The renderer is responsible for translating to image
// space (top-left origin, Y down) at the boundary.
type Point struct {
	X, Y float64
}

// Pt is a convenience constructor.
func Pt(x, y float64) Point { return Point{X: x, Y: y} }

// Add returns p + q.
func (p Point) Add(q Point) Point { return Point{p.X + q.X, p.Y + q.Y} }

// Sub returns p - q.
func (p Point) Sub(q Point) Point { return Point{p.X - q.X, p.Y - q.Y} }

// Scale returns p * s.
func (p Point) Scale(s float64) Point { return Point{p.X * s, p.Y * s} }

// Dot returns p · q.
func (p Point) Dot(q Point) float64 { return p.X*q.X + p.Y*q.Y }

// Length returns |p|.
func (p Point) Length() float64 { return math.Hypot(p.X, p.Y) }

// LengthSq returns |p|^2.
func (p Point) LengthSq() float64 { return p.X*p.X + p.Y*p.Y }

// Distance returns |p - q|.
func (p Point) Distance(q Point) float64 { return math.Hypot(p.X-q.X, p.Y-q.Y) }

// Normalize returns p / |p|, or the zero point if p is the zero point.
func (p Point) Normalize() Point {
	l := p.Length()
	if l == 0 {
		return Point{}
	}
	return Point{p.X / l, p.Y / l}
}

// Perp returns a vector perpendicular to p (rotated 90° counter-clockwise
// in math convention, where Y points up).
func (p Point) Perp() Point { return Point{-p.Y, p.X} }

// Lerp returns the linear interpolation between p and q at parameter t.
// t=0 returns p; t=1 returns q.
func (p Point) Lerp(q Point, t float64) Point {
	return Point{p.X + (q.X-p.X)*t, p.Y + (q.Y-p.Y)*t}
}
