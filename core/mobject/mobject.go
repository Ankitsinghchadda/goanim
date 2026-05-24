package mobject

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Mobject is the goanim drawable abstraction. Every shape, group, and
// domain object implements this interface.
//
// Implementations must hold a stable Seed that is set at construction
// and never changes. The seed flows into the roughness engine to keep
// hand-drawn geometry temporally stable across frames.
//
// Style is the per-mobject override layer of the three-layer style
// chain (library → scene → mobject). Unset fields inherit at render
// time via Context.Resolve.
type Mobject interface {
	Render(r render.Renderer, ctx style.Context)
	Bounds() geometry.Rect
	Children() []Mobject
	Seed() int64
	Style() *style.Style
	SetStyle(s style.Style)
}
