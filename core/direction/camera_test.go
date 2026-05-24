package direction_test

import (
	"testing"
	"time"

	"github.com/ankitsinghchadda/goanim/core/direction"
	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// TestCameraIdentityByDefault — a fresh camera reports the identity
// viewport (zoom 1, origin centered). Confirms that adding a camera
// to a scene doesn't accidentally shift content.
func TestCameraIdentityByDefault(t *testing.T) {
	c := direction.NewCamera()
	cx, cy := c.Position()
	if cx != 0 || cy != 0 {
		t.Errorf("default camera position = (%v, %v), want (0, 0)", cx, cy)
	}
	if z := c.ZoomLevel(); z != 1 {
		t.Errorf("default camera zoom = %v, want 1", z)
	}
	if d := c.DimAmount(); d != 1 {
		t.Errorf("default dim amount = %v, want 1", d)
	}
	if c.FocusTarget() != nil {
		t.Errorf("default focus target != nil")
	}
}

// TestZoomToInterpolates — running ZoomTo end-to-end takes the camera
// from identity to the target's center at the target zoom.
func TestZoomToInterpolates(t *testing.T) {
	c := direction.NewCamera()
	target := positionedMob(1, 200, -50)

	anim := c.ZoomTo(target, 2.0, 1*time.Second)
	anim.Apply(0)
	cx, cy := c.Position()
	if cx != 0 || cy != 0 || c.ZoomLevel() != 1 {
		t.Errorf("t=0: pos=(%v,%v), zoom=%v; want identity", cx, cy, c.ZoomLevel())
	}
	anim.Apply(1)
	cx, cy = c.Position()
	if cx != 200 || cy != -50 {
		t.Errorf("t=1: pos=(%v, %v), want (200, -50)", cx, cy)
	}
	if z := c.ZoomLevel(); z != 2.0 {
		t.Errorf("t=1: zoom=%v, want 2.0", z)
	}
}

// TestCameraDoesNotMutateTarget — a camera move must NOT touch the
// target mobject's position. This is the architectural invariant: the
// camera transforms the viewport, not the diagram.
func TestCameraDoesNotMutateTarget(t *testing.T) {
	c := direction.NewCamera()
	target := positionedMob(1, 100, 100)

	posX, posY := target.Position()

	anim := c.ZoomTo(target, 3.0, 1*time.Second)
	for _, tt := range []float64{0, 0.25, 0.5, 0.75, 1} {
		anim.Apply(tt)
	}

	if gotX, gotY := target.Position(); gotX != posX || gotY != posY {
		t.Errorf("camera move mutated target position: was (%v, %v), got (%v, %v)",
			posX, posY, gotX, gotY)
	}
}

// TestResetReturnsIdentity — Reset takes any camera state back to
// (0, 0, 1).
func TestResetReturnsIdentity(t *testing.T) {
	c := direction.NewCamera()
	c.Cx, c.Cy, c.Zoom = 500, -200, 3.5
	anim := c.Reset(500 * time.Millisecond)
	anim.Apply(1)
	cx, cy := c.Position()
	if cx != 0 || cy != 0 || c.ZoomLevel() != 1 {
		t.Errorf("after Reset: pos=(%v,%v) zoom=%v; want identity", cx, cy, c.ZoomLevel())
	}
}

// TestFocusSetsDimAndTarget — at t=1 Focus has dim<1 and the focus
// target set; UnFocus reverses.
func TestFocusSetsDimAndTarget(t *testing.T) {
	c := direction.NewCamera()
	target := positionedMob(1, 0, 0)

	focus := c.Focus(target, 1.5, 1*time.Second)
	focus.Apply(0)
	focus.Apply(1)
	if c.FocusTarget() != target {
		t.Errorf("focus end: target should be set")
	}
	if c.DimAmount() >= 1 {
		t.Errorf("focus end: dim should be < 1, got %v", c.DimAmount())
	}

	unfocus := c.UnFocus(1 * time.Second)
	unfocus.Apply(0)
	unfocus.Apply(1)
	if c.DimAmount() != 1 {
		t.Errorf("unfocus end: dim should be 1, got %v", c.DimAmount())
	}
}

func positionedMob(seed int64, x, y float64) *posMob {
	m := &posMob{Group: mobject.NewGroup(seed)}
	m.SetPosition(x, y)
	return m
}

// posMob is a minimal Mobject that knows its position — enough to
// pass to camera animations that need a target.
type posMob struct {
	*mobject.Group
	posX, posY float64
}

func (p *posMob) Position() (float64, float64) { return p.posX, p.posY }
func (p *posMob) SetPosition(x, y float64)     { p.posX = x; p.posY = y }
