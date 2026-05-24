package layout

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Grid arranges children in a rows×cols matrix, in row-major order.
//
// Row and column gaps are configurable. Per-cell alignment defaults to
// center; a cell wider/taller than its content places its content
// according to the cell-align attributes.
type Grid struct {
	*mobject.Group
	cx, cy   float64
	rows     int
	cols     int
	children []mobject.Mobject
	rowGap   float64
	colGap   float64
	halign   HorizontalAlign
	valign   VerticalAlign
	dirty    bool
}

// NewGrid constructs a rows×cols grid with children placed in row-major
// order. Extra cells (when len(children) < rows*cols) are left empty;
// extra children beyond rows*cols are ignored.
func NewGrid(rows, cols int, children ...mobject.Mobject) *Grid {
	g := &Grid{
		Group:    mobject.NewGroup(0),
		rows:     rows,
		cols:     cols,
		children: append([]mobject.Mobject{}, children...),
		halign:   HCenter,
		valign:   VMiddle,
		dirty:    true,
	}
	for _, c := range children {
		g.Group.Add(c)
	}
	return g
}

func (g *Grid) WithSpacing(rowGap, colGap float64) *Grid {
	g.rowGap, g.colGap = rowGap, colGap
	g.dirty = true
	return g
}

func (g *Grid) WithCellAlign(h HorizontalAlign, v VerticalAlign) *Grid {
	g.halign, g.valign = h, v
	g.dirty = true
	return g
}

func (g *Grid) MoveTo(x, y float64) *Grid    { g.cx, g.cy = x, y; g.dirty = true; return g }
func (g *Grid) SetPosition(x, y float64)     { g.cx, g.cy = x, y; g.dirty = true }
func (g *Grid) Position() (float64, float64) { return g.cx, g.cy }
func (g *Grid) Children() []mobject.Mobject  { return g.children }
func (g *Grid) Seed() int64                  { return 0 }

// Layout positions every child in its row×col cell. Column widths are
// the max child width in that column; row heights similarly.
func (g *Grid) Layout() {
	g.dirty = false
	if len(g.children) == 0 || g.rows == 0 || g.cols == 0 {
		return
	}

	colWidths := make([]float64, g.cols)
	rowHeights := make([]float64, g.rows)
	for idx, c := range g.children {
		if idx >= g.rows*g.cols {
			break
		}
		r := idx / g.cols
		col := idx % g.cols
		b := c.Bounds()
		if b.Width() > colWidths[col] {
			colWidths[col] = b.Width()
		}
		if b.Height() > rowHeights[r] {
			rowHeights[r] = b.Height()
		}
	}

	totalW := g.colGap * float64(g.cols-1)
	for _, w := range colWidths {
		totalW += w
	}
	totalH := g.rowGap * float64(g.rows-1)
	for _, h := range rowHeights {
		totalH += h
	}

	// Top-left of the grid (Y-up).
	startX := g.cx - totalW/2
	startY := g.cy + totalH/2

	// Precompute the X center of each column and Y center of each row.
	colCenters := make([]float64, g.cols)
	x := startX
	for c := 0; c < g.cols; c++ {
		colCenters[c] = x + colWidths[c]/2
		x += colWidths[c] + g.colGap
	}
	rowCenters := make([]float64, g.rows)
	y := startY
	for r := 0; r < g.rows; r++ {
		rowCenters[r] = y - rowHeights[r]/2
		y -= rowHeights[r] + g.rowGap
	}

	for idx, c := range g.children {
		if idx >= g.rows*g.cols {
			break
		}
		r := idx / g.cols
		col := idx % g.cols
		b := c.Bounds()
		cellCX := colCenters[col]
		cellCY := rowCenters[r]
		// Cell alignment offset (when child is smaller than the cell).
		offX := 0.0
		offY := 0.0
		switch g.halign {
		case HLeft:
			offX = -(colWidths[col]/2 - b.Width()/2)
		case HRight:
			offX = +(colWidths[col]/2 - b.Width()/2)
		}
		switch g.valign {
		case VTop:
			offY = +(rowHeights[r]/2 - b.Height()/2)
		case VBottom:
			offY = -(rowHeights[r]/2 - b.Height()/2)
		}
		posMove(c, cellCX+offX, cellCY+offY)
	}
}

func (g *Grid) Bounds() geometry.Rect {
	if g.dirty {
		g.Layout()
	}
	if len(g.children) == 0 {
		return geometry.RectFromCenter(geometry.Pt(g.cx, g.cy), 0, 0)
	}
	var b geometry.Rect
	first := true
	for _, c := range g.children {
		cb := c.Bounds()
		if cb.Empty() {
			continue
		}
		if first {
			b = cb
			first = false
		} else {
			b = b.Union(cb)
		}
	}
	return b
}

func (g *Grid) Render(r render.Renderer, ctx style.Context) {
	if g.dirty {
		g.Layout()
	}
	for _, c := range g.children {
		c.Render(r, ctx)
	}
}
