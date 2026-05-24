package systemdesign

import (
	"image/color"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Packet is a small labeled marker — used to animate data flowing
// between nodes (e.g. "GET /user" hopping from Client to Database).
//
// It renders as a rounded rectangle with a centered label. Style flows
// through the normal three-layer chain; for typical use the per-packet
// style overrides StrokeColor and FillColor to make it visually
// distinct from the nodes.
type Packet struct {
	*mobject.Group
	cx, cy float64
	w, h   float64
	label  *mobject.Text
	style  style.Style
	reveal float64
}

// NewPacket constructs a packet with the given inline label. Packets
// default to a warm amber palette so they pop against any scene
// preset; users can override via WithStyle.
//
// Style philosophy: the packet pins the attributes that define a packet
// (colors, fill, the small font size). It leaves FontFamily *unset* so
// the label inherits from the scene default, matching the surrounding
// labels in any preset (e.g. a packet in PresetCrisp uses the same
// sans-serif Inter as the Client/Server labels).
func NewPacket(seed int64, labelText string) *Packet {
	p := &Packet{
		Group: mobject.NewGroup(seed),
		w:     170,
		h:     54,
		label: mobject.NewText(seed+777, labelText),
		style: style.Style{
			StrokeColor: color.RGBA{0xB4, 0x53, 0x09, 0xFF}, // dark amber
			FillColor:   color.RGBA{0xFE, 0xF3, 0xC7, 0xFF}, // light amber
			FillStyle:   style.FillSolid,
			// FontFamily intentionally unset — inherit from scene.
		},
		reveal: 1,
	}
	// Label inherits FontFamily and StrokeColor from the scene; only
	// the small size and a slightly darker label color are pinned (the
	// label sits inside a pale fill, so it needs to read against amber).
	p.label.SetStyle(style.Style{
		FontSize:    style.FontSmall,
		StrokeColor: color.RGBA{0x7A, 0x2E, 0x0E, 0xFF},
		// FontFamily intentionally unset — inherit from scene.
	})
	p.Group.Add(p.label)
	return p
}

// MoveTo sets the center.
func (p *Packet) MoveTo(x, y float64) *Packet {
	p.cx, p.cy = x, y
	p.label.MoveTo(x, y)
	return p
}

// SetPosition is the imperative form of MoveTo (used by animations).
func (p *Packet) SetPosition(x, y float64) {
	p.cx, p.cy = x, y
	p.label.SetPosition(x, y)
}

// Position returns the current center.
func (p *Packet) Position() (float64, float64) { return p.cx, p.cy }

// SetReveal cascades the reveal fraction.
func (p *Packet) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	p.reveal = t
	p.label.SetReveal(t)
}

// SetVisualScale — packets don't currently scale, but the method is
// here to satisfy animation.Scaler so PopIn works.
func (p *Packet) SetVisualScale(s float64) {
	// We map "scale" to "opacity" for packets — popping in scales
	// visually like a fade-in for small markers.
	if s < 0 {
		s = 0
	}
	p.reveal = s
	p.label.SetReveal(s)
}

// Bounds returns the packet's rectangle.
func (p *Packet) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(p.cx, p.cy), p.w, p.h)
}

// WithStyle sets the per-packet style override.
func (p *Packet) WithStyle(s style.Style) *Packet { p.style = s; return p }

// Style returns the per-packet style (in-place editable).
func (p *Packet) Style() *style.Style { return &p.style }

// SetStyle replaces the style override.
func (p *Packet) SetStyle(s style.Style) { p.style = s }

// Render draws the packet (rounded rect + label).
//
// Packets carry their own per-mobject style with stroke/fill defaults
// (set in NewPacket) so they remain visible across every scene preset.
// Reveal is binary: invisible at reveal=0, fully drawn otherwise. This
// avoids the partial-alpha rendering quirks seen with closed paths.
func (p *Packet) Render(r render.Renderer, ctx style.Context) {
	if p.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(p.style)
	tok := style.TokensFor(eff)

	x0 := p.cx - p.w/2
	y0 := p.cy - p.h/2
	outline := geometry.RectanglePath(x0, y0, p.w, p.h, 12)

	// Fill.
	if eff.FillColor != nil {
		r.DrawPath(outline, render.PathStyle{Fill: eff.FillColor})
	}

	// Outline.
	ps := style.PathStyleStroke(eff, tok)
	r.DrawPath(outline, ps)

	// Label.
	p.label.Render(r, ctx)
}
