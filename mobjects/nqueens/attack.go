package nqueens

import (
	"image/color"
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// AttackDir names one of the eight rays a chess queen attacks along.
type AttackDir uint8

const (
	DirRowLeft AttackDir = iota
	DirRowRight
	DirColUp
	DirColDown
	DirDiagUpLeft
	DirDiagUpRight
	DirDiagDownLeft
	DirDiagDownRight
)

// AllAttackDirs is the complete eight-ray set — handy when you want
// to draw the queen's full attack footprint.
var AllAttackDirs = []AttackDir{
	DirRowLeft, DirRowRight, DirColUp, DirColDown,
	DirDiagUpLeft, DirDiagUpRight, DirDiagDownLeft, DirDiagDownRight,
}

var attackRayColor = color.RGBA{0xC9, 0x1F, 0x37, 0xCC} // red, slight transparency

// AttackRays is a composite mobject: one Line per direction, originating
// at a board cell and extending to the board edge. Used during the
// "rules of chess" chapter to show what a queen attacks.
//
// Implements Revealer (animation.DrawOn) by cascading reveal to each
// line — so DrawOn(rays, ...) progressively draws each ray.
type AttackRays struct {
	*mobject.Group
	seed   int64
	lines  []*mobject.Line
	reveal float64
}

// NewAttackRays builds the attack rays from cell (row, col) of board,
// going in the requested directions. Rays terminate at the board's
// outer edge.
func NewAttackRays(seed int64, board *Chessboard, row, col int, dirs []AttackDir) *AttackRays {
	cx, cy := board.CellCenter(row, col)
	bb := board.Bounds()
	rays := &AttackRays{
		Group:  mobject.NewGroup(seed),
		seed:   seed,
		reveal: 1,
	}
	for i, dir := range dirs {
		end := rayEnd(cx, cy, dir, bb)
		line := mobject.NewLine(seed+int64(i)+1, geometry.Pt(cx, cy), end)
		line.SetStyle(style.Style{
			Sloppiness:  style.SloppinessArchitect,
			StrokeStyle: style.StrokeSolid,
			StrokeWidth: style.StrokeThick,
			StrokeColor: attackRayColor,
		})
		rays.lines = append(rays.lines, line)
		rays.Group.Add(line)
	}
	return rays
}

// rayEnd computes where the ray from (cx, cy) in direction dir hits the
// inside of the board's bounding rectangle. The smallest positive t
// along the ray that crosses an X- or Y-boundary is the answer.
func rayEnd(cx, cy float64, dir AttackDir, bb geometry.Rect) geometry.Point {
	var dx, dy float64
	switch dir {
	case DirRowLeft:
		dx, dy = -1, 0
	case DirRowRight:
		dx, dy = 1, 0
	case DirColUp:
		dx, dy = 0, 1
	case DirColDown:
		dx, dy = 0, -1
	case DirDiagUpLeft:
		dx, dy = -1, 1
	case DirDiagUpRight:
		dx, dy = 1, 1
	case DirDiagDownLeft:
		dx, dy = -1, -1
	case DirDiagDownRight:
		dx, dy = 1, -1
	}
	tx := math.Inf(1)
	if dx > 0 {
		tx = (bb.Max.X - cx) / dx
	} else if dx < 0 {
		tx = (bb.Min.X - cx) / dx
	}
	ty := math.Inf(1)
	if dy > 0 {
		ty = (bb.Max.Y - cy) / dy
	} else if dy < 0 {
		ty = (bb.Min.Y - cy) / dy
	}
	t := math.Min(tx, ty)
	return geometry.Pt(cx+t*dx, cy+t*dy)
}

// SetReveal cascades to each line, so DrawOn(rays) animates a
// stagger-free progressive draw across every ray simultaneously. To
// stagger them, build separate rays per direction and Sequence/Stagger
// the DrawOn animations.
func (a *AttackRays) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	a.reveal = t
	for _, l := range a.lines {
		l.SetReveal(t)
	}
}

// Reveal returns the current reveal fraction.
func (a *AttackRays) Reveal() float64 { return a.reveal }

// Bounds returns the union of the lines' bounding boxes.
func (a *AttackRays) Bounds() geometry.Rect { return a.Group.Bounds() }

// Render propagates the group style into the child context so FadeIn /
// FadeOut on the rays cascades.
func (a *AttackRays) Render(r render.Renderer, ctx style.Context) {
	childCtx := ctx
	childCtx.SceneDefault = a.Group.Style().Merge(ctx.SceneDefault)
	for _, l := range a.lines {
		l.Render(r, childCtx)
	}
}
