package netgraph

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Node is a labeled circle. Implements mobject.Attachable so edges
// connect to the perimeter (not the bounding-box corner).
type Node struct {
	*mobject.Group
	ellipse *mobject.Ellipse
	label   *mobject.Text
	radius  float64
	cx, cy  float64
	style   style.Style
}

// NewNode constructs a circular node of the given radius with a centered
// text label. The ellipse is outline-only by default so the label inside
// reads clearly; callers can override with WithStyle if they want a
// filled (e.g. solid-color) node.
func NewNode(seed int64, radius float64, labelText string) *Node {
	n := &Node{
		Group:   mobject.NewGroup(seed),
		ellipse: mobject.NewEllipse(seed, radius, radius),
		label:   mobject.NewText(seed+701, labelText),
		radius:  radius,
	}
	// Outline-only: don't let the scene's default fill (e.g. cross-hatch
	// from PresetSketchy) clobber the centered label.
	ellSty := *n.ellipse.Style()
	ellSty.FillStyle = style.FillNone
	n.ellipse.SetStyle(ellSty)
	n.MoveTo(0, 0)
	n.Group.Add(n.ellipse, n.label)
	return n
}

// MoveTo sets the node center.
func (n *Node) MoveTo(x, y float64) *Node {
	n.cx, n.cy = x, y
	n.ellipse.MoveTo(x, y)
	n.label.MoveTo(x, y)
	return n
}

// SetPosition is the imperative form used by animations.
func (n *Node) SetPosition(x, y float64) { n.MoveTo(x, y) }

// Position returns the current center.
func (n *Node) Position() (float64, float64) { return n.cx, n.cy }

// Radius returns the node's radius — useful for layout helpers.
func (n *Node) Radius() float64 { return n.radius }

// SetReveal cascades to the circle and label.
func (n *Node) SetReveal(t float64) {
	n.ellipse.SetReveal(t)
	n.label.SetReveal(t)
}

// SetVisualScale cascades to the ellipse for "pop" animations.
func (n *Node) SetVisualScale(s float64) { n.ellipse.SetVisualScale(s) }

// WithStyle sets the per-mobject style override on the ellipse (label
// inherits via the scene context).
func (n *Node) WithStyle(s style.Style) *Node {
	n.style = s
	n.ellipse.SetStyle(s)
	return n
}

// WithLabelStyle sets the per-mobject style override on the label only.
// Use for nodes whose label color/size differs from the ring stroke.
func (n *Node) WithLabelStyle(s style.Style) *Node {
	n.label.SetStyle(s)
	return n
}

// Style returns the per-mobject style override.
func (n *Node) Style() *style.Style { return &n.style }

// SetStyle replaces the style override. If the input doesn't set a
// FillStyle (zero value), preserve the ellipse's existing FillStyle so
// the outline-only default from NewNode survives stroke-color edits.
func (n *Node) SetStyle(s style.Style) {
	n.style = s
	current := *n.ellipse.Style()
	merged := s
	if merged.FillStyle == style.FillStyleUnset {
		merged.FillStyle = current.FillStyle
	}
	n.ellipse.SetStyle(merged)
}

// Bounds returns the node's bounding box.
func (n *Node) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(n.cx, n.cy), n.radius*2, n.radius*2)
}

// AttachmentPoint returns the point on the circle boundary along the
// line from the center toward the named side — so edges meet the rim,
// not the bounding-box corner.
func (n *Node) AttachmentPoint(side mobject.Side) geometry.Point {
	return mobject.AttachToEllipse(n.cx, n.cy, n.radius, n.radius, side)
}

// AttachmentTowards returns the point on the circle boundary along the
// direction from this node's center toward `target`. Used by Edge so
// the line meets the circle on the chord between centers (which looks
// correct for any angle, not just the four cardinal sides).
func (n *Node) AttachmentTowards(target geometry.Point) geometry.Point {
	dx := target.X - n.cx
	dy := target.Y - n.cy
	l := dx*dx + dy*dy
	if l < 1e-12 {
		return geometry.Pt(n.cx, n.cy)
	}
	inv := 1.0 / sqrt(l)
	return geometry.Pt(n.cx+dx*inv*n.radius, n.cy+dy*inv*n.radius)
}

func (n *Node) Render(r render.Renderer, ctx style.Context) {
	// Propagate the Node-local opacity to children. The animation
	// package's FadeIn sets opacity on the *Node's* style, but Group
	// doesn't pass that down to the ellipse/label children — without
	// this propagation the Node would appear binary (visible or not)
	// even mid-fade. We push the local opacity into each child's
	// style override at render time, then restore so repeated frames
	// stay stable.
	if n.style.Opacity != nil {
		opPtr := n.style.Opacity
		ellSty := *n.ellipse.Style()
		labSty := *n.label.Style()
		oldE := ellSty.Opacity
		oldL := labSty.Opacity
		ellSty.Opacity = opPtr
		labSty.Opacity = opPtr
		n.ellipse.SetStyle(ellSty)
		n.label.SetStyle(labSty)
		n.Group.Render(r, ctx)
		ellSty.Opacity = oldE
		labSty.Opacity = oldL
		n.ellipse.SetStyle(ellSty)
		n.label.SetStyle(labSty)
		return
	}
	n.Group.Render(r, ctx)
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 6; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
