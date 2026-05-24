package mobject

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Group is a Mobject composed of children. It renders each child in
// order and unions their bounds. Domain mobjects (Server, Database,
// ...) embed *Group and add semantic methods on top.
type Group struct {
	seed     int64
	children []Mobject
	style    style.Style
}

// NewGroup constructs an empty group with the given seed.
func NewGroup(seed int64) *Group { return &Group{seed: seed} }

// Add appends one or more children.
func (g *Group) Add(m ...Mobject) *Group {
	g.children = append(g.children, m...)
	return g
}

// SetChildren replaces all children.
func (g *Group) SetChildren(m []Mobject) { g.children = m }

// Render draws every child in insertion order. The style context is
// inherited as-is — groups don't introduce a new scope (set a per-child
// style if you need to override).
func (g *Group) Render(r render.Renderer, ctx style.Context) {
	for _, c := range g.children {
		c.Render(r, ctx)
	}
}

// Bounds returns the union of all child bounds.
func (g *Group) Bounds() geometry.Rect {
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
	if first {
		return geometry.Rect{Min: geometry.Pt(1, 1), Max: geometry.Pt(-1, -1)}
	}
	return b
}

// Children returns the child list (live — do not mutate).
func (g *Group) Children() []Mobject { return g.children }

// Seed returns the group's stable random seed.
func (g *Group) Seed() int64 { return g.seed }

// AttachmentPoint returns the midpoint of the named edge of the group's
// bounding box. Subtypes (Server, Database) that need shape-aware
// attachment override this.
func (g *Group) AttachmentPoint(side Side) geometry.Point {
	return boundsEdgeMidpoint(g.Bounds(), side)
}

// Style returns the group's style override (pointer for in-place edits).
func (g *Group) Style() *style.Style { return &g.style }

// SetStyle replaces the group's style override.
func (g *Group) SetStyle(s style.Style) { g.style = s }
