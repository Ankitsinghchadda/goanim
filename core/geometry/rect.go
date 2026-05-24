package geometry

import "math"

// Rect is an axis-aligned bounding rectangle in user space.
// Min is the lower-left corner (smallest X, smallest Y), Max is the
// upper-right corner. An empty Rect has Min.X > Max.X.
type Rect struct {
	Min, Max Point
}

// RectFromCenter constructs a rectangle of width w and height h centered at c.
func RectFromCenter(c Point, w, h float64) Rect {
	return Rect{
		Min: Point{c.X - w/2, c.Y - h/2},
		Max: Point{c.X + w/2, c.Y + h/2},
	}
}

// RectFromPoints constructs the bounding box of a sequence of points.
// Returns the empty rectangle if pts is empty.
func RectFromPoints(pts ...Point) Rect {
	if len(pts) == 0 {
		return Rect{Min: Point{1, 1}, Max: Point{-1, -1}}
	}
	r := Rect{Min: pts[0], Max: pts[0]}
	for _, p := range pts[1:] {
		if p.X < r.Min.X {
			r.Min.X = p.X
		}
		if p.Y < r.Min.Y {
			r.Min.Y = p.Y
		}
		if p.X > r.Max.X {
			r.Max.X = p.X
		}
		if p.Y > r.Max.Y {
			r.Max.Y = p.Y
		}
	}
	return r
}

// Width returns Max.X - Min.X.
func (r Rect) Width() float64 { return r.Max.X - r.Min.X }

// Height returns Max.Y - Min.Y.
func (r Rect) Height() float64 { return r.Max.Y - r.Min.Y }

// Center returns the center point of the rectangle.
func (r Rect) Center() Point {
	return Point{(r.Min.X + r.Max.X) / 2, (r.Min.Y + r.Max.Y) / 2}
}

// Empty reports whether the rectangle has non-positive width or height.
func (r Rect) Empty() bool { return r.Max.X <= r.Min.X || r.Max.Y <= r.Min.Y }

// Union returns the smallest rectangle containing both r and s.
func (r Rect) Union(s Rect) Rect {
	if r.Empty() {
		return s
	}
	if s.Empty() {
		return r
	}
	return Rect{
		Min: Point{math.Min(r.Min.X, s.Min.X), math.Min(r.Min.Y, s.Min.Y)},
		Max: Point{math.Max(r.Max.X, s.Max.X), math.Max(r.Max.Y, s.Max.Y)},
	}
}

// Inset returns r with all four sides moved inward by d. Negative d expands.
func (r Rect) Inset(d float64) Rect {
	return Rect{
		Min: Point{r.Min.X + d, r.Min.Y + d},
		Max: Point{r.Max.X - d, r.Max.Y - d},
	}
}

// Contains reports whether the point p lies within the rectangle (inclusive).
func (r Rect) Contains(p Point) bool {
	return p.X >= r.Min.X && p.X <= r.Max.X && p.Y >= r.Min.Y && p.Y <= r.Max.Y
}

// Corners returns the four corners of the rectangle in counter-clockwise
// order, starting from the lower-left.
func (r Rect) Corners() [4]Point {
	return [4]Point{
		{r.Min.X, r.Min.Y},
		{r.Max.X, r.Min.Y},
		{r.Max.X, r.Max.Y},
		{r.Min.X, r.Max.Y},
	}
}
