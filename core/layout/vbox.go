package layout

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// VBox arranges children in a vertical column.
//
// Children appear top → bottom at uniform horizontal alignment (default
// HCenter = aligned centers). Spacing is added between consecutive
// children. Y-up convention: the first child is rendered at the
// container's top, the last at the bottom.
type VBox struct {
	*mobject.Group
	cx, cy                 float64
	children               []mobject.Mobject
	spacing                float64
	halign                 HorizontalAlign
	padT, padR, padB, padL float64
	dirty                  bool
}

// NewVBox builds a vertical column of children.
func NewVBox(children ...mobject.Mobject) *VBox {
	v := &VBox{
		Group:    mobject.NewGroup(0),
		children: append([]mobject.Mobject{}, children...),
		halign:   HCenter,
		dirty:    true,
	}
	for _, c := range children {
		v.Group.Add(c)
	}
	return v
}

func (v *VBox) WithSpacing(px float64) *VBox { v.spacing = px; v.dirty = true; return v }
func (v *VBox) WithAlign(h HorizontalAlign) *VBox {
	v.halign = h
	v.dirty = true
	return v
}
func (v *VBox) WithPadding(top, right, bottom, left float64) *VBox {
	v.padT, v.padR, v.padB, v.padL = top, right, bottom, left
	v.dirty = true
	return v
}
func (v *VBox) WithPaddingAll(p float64) *VBox { return v.WithPadding(p, p, p, p) }

func (v *VBox) MoveTo(x, y float64) *VBox    { v.cx, v.cy = x, y; v.dirty = true; return v }
func (v *VBox) SetPosition(x, y float64)     { v.cx, v.cy = x, y; v.dirty = true }
func (v *VBox) Position() (float64, float64) { return v.cx, v.cy }
func (v *VBox) Children() []mobject.Mobject  { return v.children }
func (v *VBox) Seed() int64                  { return 0 }

// Layout positions children top → bottom centered horizontally.
func (v *VBox) Layout() {
	v.dirty = false
	if len(v.children) == 0 {
		return
	}

	widths := make([]float64, len(v.children))
	heights := make([]float64, len(v.children))
	for i, c := range v.children {
		b := c.Bounds()
		widths[i] = b.Width()
		heights[i] = b.Height()
	}
	totalH := v.spacing * float64(len(v.children)-1)
	for _, h := range heights {
		totalH += h
	}
	maxW := 0.0
	for _, w := range widths {
		if w > maxW {
			maxW = w
		}
	}

	// Y-up: first child renders at the TOP of the column. Top edge =
	// cy + totalH/2.
	topY := v.cy + totalH/2
	for i, c := range v.children {
		cy := topY - heights[i]/2
		var cx float64
		switch v.halign {
		case HLeft:
			cx = v.cx - maxW/2 + widths[i]/2
		case HRight:
			cx = v.cx + maxW/2 - widths[i]/2
		default: // HCenter
			cx = v.cx
		}
		posMove(c, cx, cy)
		topY -= heights[i] + v.spacing
	}
}

func (v *VBox) Bounds() geometry.Rect {
	if v.dirty {
		v.Layout()
	}
	if len(v.children) == 0 {
		return geometry.RectFromCenter(geometry.Pt(v.cx, v.cy), 0, 0)
	}
	var b geometry.Rect
	first := true
	for _, c := range v.children {
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
	b.Min.X -= v.padL
	b.Min.Y -= v.padB
	b.Max.X += v.padR
	b.Max.Y += v.padT
	return b
}

func (v *VBox) Render(r render.Renderer, ctx style.Context) {
	if v.dirty {
		v.Layout()
	}
	for _, c := range v.children {
		c.Render(r, ctx)
	}
}
