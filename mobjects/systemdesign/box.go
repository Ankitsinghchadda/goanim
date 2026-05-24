package systemdesign

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// box is the rectangular base shared by Client and Server. It carries
// a Mobject Rectangle for the body and a Text label centered inside.
type box struct {
	rect   *mobject.Rectangle
	label  *mobject.Text
	cx, cy float64
}

func newBox(seed int64, w, h float64, labelText string) *box {
	b := &box{
		rect:  mobject.NewRectangle(seed, w, h),
		label: mobject.NewText(seed+777, labelText),
	}
	b.moveTo(0, 0)
	return b
}

func (b *box) moveTo(x, y float64) {
	b.cx, b.cy = x, y
	b.rect.MoveTo(x, y)
	b.label.MoveTo(x, y)
}

// Client is a labeled rectangle representing an end-user device.
type Client struct {
	*mobject.Group
	box *box
}

// NewClient constructs a Client with the given label.
func NewClient(seed int64, labelText string) *Client {
	c := &Client{Group: mobject.NewGroup(seed)}
	c.box = newBox(seed, 300, 200, labelText)
	c.Group.Add(c.box.rect, c.box.label)
	return c
}

// MoveTo sets the center.
func (c *Client) MoveTo(x, y float64) *Client { c.box.moveTo(x, y); return c }

// SetPosition is the imperative form of MoveTo (used by animations).
func (c *Client) SetPosition(x, y float64) { c.box.moveTo(x, y) }

// Position returns the current center.
func (c *Client) Position() (float64, float64) { return c.box.cx, c.box.cy }

// SetReveal cascades the reveal fraction to children.
func (c *Client) SetReveal(t float64) {
	c.box.rect.SetReveal(t)
	c.box.label.SetReveal(t)
}

// SetVisualScale cascades a uniform scale to the rect.
func (c *Client) SetVisualScale(s float64) { c.box.rect.SetVisualScale(s) }

// WithStyle sets the per-mobject style override.
func (c *Client) WithStyle(s style.Style) *Client {
	c.box.rect.SetStyle(s)
	return c
}

// Bounds returns the rectangle's bounding box.
func (c *Client) Bounds() geometry.Rect { return c.box.rect.Bounds() }

// Server is a labeled rectangle with a "rack lines" decoration in the
// top-right corner — a small visual cue distinguishing it from a Client.
type Server struct {
	*mobject.Group
	box    *box
	reveal float64
}

// NewServer constructs a Server with the given label.
func NewServer(seed int64, labelText string) *Server {
	s := &Server{Group: mobject.NewGroup(seed), reveal: 1}
	s.box = newBox(seed, 320, 220, labelText)
	s.Group.Add(s.box.rect, s.box.label)
	return s
}

// MoveTo sets the center.
func (s *Server) MoveTo(x, y float64) *Server { s.box.moveTo(x, y); return s }

// SetPosition is the imperative form of MoveTo.
func (s *Server) SetPosition(x, y float64) { s.box.moveTo(x, y) }

// Position returns the current center.
func (s *Server) Position() (float64, float64) { return s.box.cx, s.box.cy }

// SetReveal cascades the reveal fraction to children.
func (s *Server) SetReveal(t float64) {
	s.box.rect.SetReveal(t)
	s.box.label.SetReveal(t)
	s.reveal = clampMobjectLocal(t)
}

// SetVisualScale cascades a uniform scale to the rect.
func (s *Server) SetVisualScale(v float64) { s.box.rect.SetVisualScale(v) }

// WithStyle sets the per-mobject style override on the rect (label inherits).
func (s *Server) WithStyle(st style.Style) *Server {
	s.box.rect.SetStyle(st)
	return s
}

// Bounds returns the rectangle's bounding box.
func (s *Server) Bounds() geometry.Rect { return s.box.rect.Bounds() }

func clampMobjectLocal(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

// Render delegates to the embedded group then overlays the rack-lines.
func (s *Server) Render(r render.Renderer, ctx style.Context) {
	s.Group.Render(r, ctx)
	if s.reveal <= 0 {
		return
	}
	// Draw three short "rack" lines in the top-right corner. Style
	// follows the resolved style.
	eff := ctx.Resolve(*s.box.rect.Style())
	tok := style.TokensFor(eff)

	bb := s.box.rect.Bounds()
	rightX := bb.Max.X - 22
	topY := bb.Max.Y - 24

	for i := 0; i < 3; i++ {
		y := topY - float64(i)*12
		p1 := geometry.Pt(rightX-44, y)
		p2 := geometry.Pt(rightX, y)
		ps := style.PathStyleStroke(eff, tok)
		ps.StrokeWidth = ps.StrokeWidth * 0.7
		var path *geometry.Path
		if tok.Roughness == 0 {
			path = geometry.LinePath(p1, p2)
		} else {
			opts := style.RoughOptions(eff, tok, s.Seed()+int64(1001+i*17))
			opts.Roughness = tok.Roughness * 0.6
			opts.DisableMultiStroke = true
			path = rough.RoughLine(p1, p2, opts)
		}
		// Rack lines: opacity-fade with reveal (lines are short — truncating
		// would look glitchy at this scale).
		tokRack := tok
		tokRack.OpacityScale = tok.OpacityScale * s.reveal
		ps.Stroke = style.ApplyOpacity(eff.StrokeColor, tokRack.OpacityScale)
		r.DrawPath(path, ps)
	}
}
