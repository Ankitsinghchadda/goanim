package layout

import (
	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// Positionable is the contract layout containers need from their
// children: a way to move a mobject to a given center. Most goanim
// mobjects satisfy this (the systemdesign domain types implement
// SetPosition; Rectangle/Ellipse/Line/Text implement MoveTo).
//
// We provide a small wrapper, posWrap, that adapts mobjects to the
// shared "set center" API regardless of which constructor pattern
// (MoveTo or SetPosition) they expose.
type Positionable interface {
	mobject.Mobject
	SetPosition(x, y float64)
}

// posMove sets the center of m to (x, y). It tries SetPosition first
// (the imperative form used by animation primitives); if the mobject
// only exposes MoveTo we still support it.
//
// If m is itself a layout container, posMove also triggers its Layout
// so the new center cascades to grandchildren. This is how nested
// containers (HBox inside VBox) stay coherent after a parent's
// Layout pass.
func posMove(m mobject.Mobject, x, y float64) {
	if p, ok := m.(Positionable); ok {
		p.SetPosition(x, y)
	} else {
		type mover interface{ MoveTo(x, y float64) }
		if m2, ok := m.(mover); ok {
			m2.MoveTo(x, y)
		}
	}
	// Cascade layout if the child is a container.
	if lo, ok := m.(layoutDriven); ok {
		lo.Layout()
	}
}

// layoutDriven is satisfied by any container that can recompute its
// children's positions on demand. Concrete containers in this package
// all satisfy it.
type layoutDriven interface{ Layout() }
