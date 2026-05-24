package nqueens

import (
	"image/color"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

var (
	queenBodyFill   = color.RGBA{0xCC, 0x29, 0x36, 0xFF} // crimson
	queenBodyStroke = color.RGBA{0x18, 0x10, 0x10, 0xFF} // near-black
	queenLabelColor = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF} // white
)

// Queen is a chess queen piece: a filled rounded square (the "body")
// with a bold white "Q" centered. Sized to fit comfortably inside a
// chessboard cell — pass cellSize from Chessboard.CellSize().
//
// Implements the animation.Scaler interface so animation.PopIn works
// directly: PopIn(queen, ...) scales the body up from 0 → 1.
type Queen struct {
	*mobject.Group
	seed   int64
	body   *mobject.Rectangle
	label  *mobject.Text
	bodySz float64
	cx, cy float64
	scale  float64
}

// NewQueen constructs a queen sized to fit a cell of cellSize. Default
// scale = 1 (fully visible). Use PopIn(queen, ...) to animate entrance.
func NewQueen(seed int64, cellSize float64) *Queen {
	bodySz := cellSize * 0.66
	body := mobject.NewRectangle(seed, bodySz, bodySz)
	body.SetStyle(style.Style{
		Sloppiness:  style.SloppinessArchitect,
		Edges:       style.EdgesRound,
		StrokeStyle: style.StrokeSolid,
		StrokeWidth: style.StrokeThick,
		StrokeColor: queenBodyStroke,
		FillStyle:   style.FillSolid,
		FillColor:   queenBodyFill,
	})

	label := mobject.NewText(seed+7777, "Q").WithRole(style.RoleHeading)
	label.SetStyle(style.Style{
		Sloppiness:  style.SloppinessArchitect,
		StrokeColor: queenLabelColor,
		FontFamily:  style.FontSans,
		Role:        style.RoleHeading,
	})

	q := &Queen{
		Group:  mobject.NewGroup(seed),
		seed:   seed,
		body:   body,
		label:  label,
		bodySz: bodySz,
		scale:  1,
	}
	q.Group.Add(body)
	q.Group.Add(label)
	return q
}

// MoveTo positions the queen at (x, y).
func (q *Queen) MoveTo(x, y float64) *Queen { q.SetPosition(x, y); return q }

// SetPosition moves both body and label to (x, y).
func (q *Queen) SetPosition(x, y float64) {
	q.cx, q.cy = x, y
	q.body.SetPosition(x, y)
	q.label.SetPosition(x, y)
}

// Position returns the queen's center.
func (q *Queen) Position() (float64, float64) { return q.cx, q.cy }

// Bounds returns the body's bounding box (the label sits inside it).
func (q *Queen) Bounds() geometry.Rect {
	s := q.bodySz * q.scale
	return geometry.RectFromCenter(geometry.Pt(q.cx, q.cy), s, s)
}

// SetVisualScale scales the body and label together. animation.PopIn
// drives this from 0 → 1.
func (q *Queen) SetVisualScale(s float64) {
	q.scale = s
	q.body.SetVisualScale(s)
	// Text doesn't have a visual scale; we shrink it via its opacity-only
	// reveal instead so the label stays centered as the body grows.
	q.label.SetReveal(s)
}

// VisualScale returns the current scale.
func (q *Queen) VisualScale() float64 { return q.scale }

// Render propagates the group style (notably Opacity) into the child
// context — so FadeIn(queen) fades body + label together.
func (q *Queen) Render(r render.Renderer, ctx style.Context) {
	childCtx := ctx
	childCtx.SceneDefault = q.Group.Style().Merge(ctx.SceneDefault)
	q.body.Render(r, childCtx)
	q.label.Render(r, childCtx)
}
