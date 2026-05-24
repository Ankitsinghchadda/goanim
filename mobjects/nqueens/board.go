// Package nqueens provides the visual primitives for the
// "Explaining N-Queens" video: the chessboard grid, queen pieces,
// and attack-ray overlays.
//
// All three mobjects use the SloppinessArchitect (crisp) style so they
// render as a clean chess diagram on top of a sketchy scene background
// — matching the "Hybrid" look chosen in the plan.
package nqueens

import (
	"image/color"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Default colors for the chessboard cells. Picked to read well on the
// warm-cream sketchy background (the same RGBA{0xFF, 0xF8, 0xE1, 0xFF}
// used in cmd/bgmi_synced/main.go).
var (
	defaultLightCell = color.RGBA{0xF5, 0xE6, 0xC4, 0xFF} // warm cream-tan
	defaultDarkCell  = color.RGBA{0xB1, 0x82, 0x3F, 0xFF} // warm wood
	defaultBoardLine = color.RGBA{0x33, 0x2A, 0x1A, 0xFF} // deep brown
)

// Chessboard is an N×N grid of square cells with alternating light/dark
// fills. Row 0 is the TOP row; col 0 is the LEFT column — standard
// chess display, NOT mathematical (y-up) coordinates.
//
// The board is conceptually a single mobject for animation purposes
// (FadeIn on the board fades every cell together). Cells are exposed
// via Cell(row, col) for fine-grained pulsing / highlighting.
type Chessboard struct {
	*mobject.Group
	seed     int64
	n        int
	cellSize float64
	cx, cy   float64
	cells    []*mobject.Rectangle // row-major: index = row*n + col
}

// NewChessboard constructs an n×n board with the given per-cell pixel
// size. Centered at (0, 0) until MoveTo is called.
func NewChessboard(seed int64, n int, cellSize float64) *Chessboard {
	cb := &Chessboard{
		Group:    mobject.NewGroup(seed),
		seed:     seed,
		n:        n,
		cellSize: cellSize,
	}
	cellStyleLight := style.Style{
		Sloppiness:  style.SloppinessArchitect,
		Edges:       style.EdgesSharp,
		StrokeStyle: style.StrokeSolid,
		StrokeColor: defaultBoardLine,
		FillStyle:   style.FillSolid,
		FillColor:   defaultLightCell,
	}
	cellStyleDark := cellStyleLight
	cellStyleDark.FillColor = defaultDarkCell

	half := float64(n-1) / 2
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			x := (float64(c) - half) * cellSize
			y := (half - float64(r)) * cellSize // row 0 is the TOP (highest y)
			rect := mobject.NewRectangle(seed+int64(r*n+c)+1, cellSize, cellSize).MoveTo(x, y)
			if (r+c)%2 == 0 {
				rect.SetStyle(cellStyleLight)
			} else {
				rect.SetStyle(cellStyleDark)
			}
			cb.Group.Add(rect)
			cb.cells = append(cb.cells, rect)
		}
	}
	return cb
}

// CellCenter returns the scene-space coordinates of the center of cell
// (row, col).
func (cb *Chessboard) CellCenter(row, col int) (float64, float64) {
	half := float64(cb.n-1) / 2
	x := (float64(col) - half) * cb.cellSize
	y := (half - float64(row)) * cb.cellSize
	return cb.cx + x, cb.cy + y
}

// Cell returns the underlying Rectangle for cell (row, col) — useful
// for direction.Pulse, animation.Flash, etc.
func (cb *Chessboard) Cell(row, col int) *mobject.Rectangle {
	return cb.cells[row*cb.n+col]
}

// CellSize returns the per-cell pixel size.
func (cb *Chessboard) CellSize() float64 { return cb.cellSize }

// N returns the board's side length.
func (cb *Chessboard) N() int { return cb.n }

// Position returns the board's center.
func (cb *Chessboard) Position() (float64, float64) { return cb.cx, cb.cy }

// SetPosition moves the board (and every cell with it).
func (cb *Chessboard) SetPosition(x, y float64) {
	dx := x - cb.cx
	dy := y - cb.cy
	cb.cx, cb.cy = x, y
	for _, c := range cb.cells {
		px, py := c.Position()
		c.SetPosition(px+dx, py+dy)
	}
}

// MoveTo is the chainable form of SetPosition.
func (cb *Chessboard) MoveTo(x, y float64) *Chessboard { cb.SetPosition(x, y); return cb }

// Bounds returns the board's axis-aligned bounding box.
func (cb *Chessboard) Bounds() geometry.Rect {
	side := float64(cb.n) * cb.cellSize
	return geometry.RectFromCenter(geometry.Pt(cb.cx, cb.cy), side, side)
}

// Render iterates cells with the board's group style pushed into the
// child context's SceneDefault. This is how an Opacity set on the board
// (via FadeIn / FadeOut) cascades down to every cell.
func (cb *Chessboard) Render(r render.Renderer, ctx style.Context) {
	childCtx := ctx
	childCtx.SceneDefault = cb.Group.Style().Merge(ctx.SceneDefault)
	for _, c := range cb.cells {
		c.Render(r, childCtx)
	}
}
