package layout

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Stack overlays children at the same anchor position. Children render
// in insertion order — the last child is on top. The anchor controls
// where the overlay aligns (default: center).
//
// Useful for badges, callouts, or composing multiple visuals at the
// same spot.
type Stack struct {
	*mobject.Group
	cx, cy   float64
	children []mobject.Mobject
	halign   HorizontalAlign
	valign   VerticalAlign
	dirty    bool
}

func NewStack(children ...mobject.Mobject) *Stack {
	s := &Stack{
		Group:    mobject.NewGroup(0),
		children: append([]mobject.Mobject{}, children...),
		halign:   HCenter,
		valign:   VMiddle,
		dirty:    true,
	}
	for _, c := range children {
		s.Group.Add(c)
	}
	return s
}

func (s *Stack) WithAlign(h HorizontalAlign, v VerticalAlign) *Stack {
	s.halign, s.valign = h, v
	s.dirty = true
	return s
}
func (s *Stack) MoveTo(x, y float64) *Stack   { s.cx, s.cy = x, y; s.dirty = true; return s }
func (s *Stack) SetPosition(x, y float64)     { s.cx, s.cy = x, y; s.dirty = true }
func (s *Stack) Position() (float64, float64) { return s.cx, s.cy }
func (s *Stack) Children() []mobject.Mobject  { return s.children }
func (s *Stack) Seed() int64                  { return 0 }

// Layout places every child at the same anchored position. Per the
// anchor, children's centers are offset by half their bounds from the
// anchor edge.
func (s *Stack) Layout() {
	s.dirty = false
	for _, c := range s.children {
		b := c.Bounds()
		// Default to center; offset by anchor.
		dx, dy := 0.0, 0.0
		switch s.halign {
		case HLeft:
			dx = b.Width() / 2
		case HRight:
			dx = -b.Width() / 2
		}
		switch s.valign {
		case VTop:
			dy = -b.Height() / 2
		case VBottom:
			dy = b.Height() / 2
		}
		posMove(c, s.cx+dx, s.cy+dy)
	}
}

func (s *Stack) Bounds() geometry.Rect {
	if s.dirty {
		s.Layout()
	}
	if len(s.children) == 0 {
		return geometry.RectFromCenter(geometry.Pt(s.cx, s.cy), 0, 0)
	}
	var b geometry.Rect
	first := true
	for _, c := range s.children {
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

func (s *Stack) Render(r render.Renderer, ctx style.Context) {
	if s.dirty {
		s.Layout()
	}
	for _, c := range s.children {
		c.Render(r, ctx)
	}
}

// Padding wraps a single child with padding on all four sides. It
// reports an enlarged bounding box (child bounds + padding) but
// doesn't move the child — the child sits at the same position as it
// would without the padding wrapper.
type Padding struct {
	*mobject.Group
	child                  mobject.Mobject
	padT, padR, padB, padL float64
}

// NewPadding wraps child with top/right/bottom/left padding.
func NewPadding(child mobject.Mobject, top, right, bottom, left float64) *Padding {
	p := &Padding{
		Group: mobject.NewGroup(0),
		child: child,
		padT:  top, padR: right, padB: bottom, padL: left,
	}
	p.Group.Add(child)
	return p
}

// NewPaddingAll wraps child with uniform padding on all four sides.
func NewPaddingAll(child mobject.Mobject, all float64) *Padding {
	return NewPadding(child, all, all, all, all)
}

func (p *Padding) Bounds() geometry.Rect {
	b := p.child.Bounds()
	b.Min.X -= p.padL
	b.Min.Y -= p.padB
	b.Max.X += p.padR
	b.Max.Y += p.padT
	return b
}

func (p *Padding) Render(r render.Renderer, ctx style.Context) {
	p.child.Render(r, ctx)
}
func (p *Padding) Children() []mobject.Mobject { return []mobject.Mobject{p.child} }
func (p *Padding) Seed() int64                 { return p.child.Seed() }
func (p *Padding) SetPosition(x, y float64)    { posMove(p.child, x, y) }

// MoveTo repositions the padded child (which forwards to the wrapped child).
func (p *Padding) MoveTo(x, y float64) *Padding { posMove(p.child, x, y); return p }

// AlignTo positions child relative to a reference mobject with the
// given anchor. Common uses: a label below a node; a badge in a corner.
func AlignTo(child mobject.Mobject, ref mobject.Mobject, anchor Anchor, gap float64) {
	refB := ref.Bounds()
	chB := child.Bounds()
	refC := refB.Center()
	var x, y float64
	switch anchor {
	case AnchorAbove:
		x = refC.X
		y = refB.Max.Y + gap + chB.Height()/2
	case AnchorBelow:
		x = refC.X
		y = refB.Min.Y - gap - chB.Height()/2
	case AnchorLeftOf:
		x = refB.Min.X - gap - chB.Width()/2
		y = refC.Y
	case AnchorRightOf:
		x = refB.Max.X + gap + chB.Width()/2
		y = refC.Y
	case AnchorCenter:
		x = refC.X
		y = refC.Y
	}
	posMove(child, x, y)
}
