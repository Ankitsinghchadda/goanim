package direction

import (
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/animation/easing"
	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// Camera holds the viewport state — what region of the scene maps to
// the output frame — and exposes animation factories that move it.
//
// The viewport is described by two parameters:
//
//   - Center (Cx, Cy): scene-space coordinates that land at the center
//     of the output frame.
//   - Zoom: 1.0 is identity. 2.0 means each scene unit covers 2 frame
//     pixels — i.e., zoomed in 2×. Stroke widths scale with zoom (a
//     2px stroke at zoom 2.0 renders as 4 pixels in the output) to
//     match the cinematic feel of a camera move.
//
// The Camera is held by reference inside the Scene. Animations that
// move the camera (ZoomTo, PanTo, Reset, Focus) mutate Camera.Cx,
// Camera.Cy, Camera.Zoom over time. The scene player calls
// renderer.SetCamera(Cx, Cy, Zoom) at every frame so mobjects render
// through the current viewport without changing their own state.
//
// Default values (Cx=0, Cy=0, Zoom=1) reproduce the pre-Phase-7
// rendering: the scene's origin at the frame center, no zoom.
type Camera struct {
	Cx   float64
	Cy   float64
	Zoom float64

	// Dimming machinery for Focus / Spotlight. When focusTarget is
	// non-nil, mobjects whose bounds are outside the focus region get
	// an opacity factor of dimAmount; the focus target stays at 1.0.
	//
	// Stored on Camera (not Scene) so the dimming animates alongside
	// the zoom animation and tears down together via UnFocus.
	focusTarget mobject.Mobject
	dimAmount   float64
}

// NewCamera constructs a Camera with the identity viewport. Scenes
// own a Camera by default — callers usually access it via Scene.Camera
// rather than constructing one directly.
func NewCamera() *Camera {
	return &Camera{Zoom: 1, dimAmount: 1}
}

// Bounded is the interface camera animations expect from their target
// — a mobject that can report a center via Position() OR (fallback) a
// bounding rect. Most goanim mobjects satisfy this through Group's
// inherited Position method.
type Bounded interface {
	mobject.Mobject
}

// ZoomTo returns an animation that pans the camera to target's center
// and changes the zoom level to factor, over duration. Default easing
// is EaseInOutCubic — cinematic, no linear-motion feel.
//
// The animation mutates the Camera struct's Cx, Cy, Zoom fields. It
// doesn't touch the target mobject (so a ZoomTo composed in Parallel
// with a FadeIn on the same target works as expected).
func (c *Camera) ZoomTo(target mobject.Mobject, factor float64, duration time.Duration) animation.Animation {
	cx, cy := positionFor(target)
	return &camMove{
		cam:    c,
		fromCx: c.Cx, fromCy: c.Cy, fromZoom: c.Zoom,
		toCx: cx, toCy: cy, toZoom: factor,
		dur:  duration,
		ease: easing.InOutCubic,
	}
}

// PanTo pans the camera to a scene-space point without changing zoom.
func (c *Camera) PanTo(x, y float64, duration time.Duration) animation.Animation {
	return &camMove{
		cam:    c,
		fromCx: c.Cx, fromCy: c.Cy, fromZoom: c.Zoom,
		toCx: x, toCy: y, toZoom: c.Zoom,
		dur:  duration,
		ease: easing.InOutCubic,
	}
}

// Reset returns the camera to (0, 0, 1) over duration.
func (c *Camera) Reset(duration time.Duration) animation.Animation {
	return &camMove{
		cam:    c,
		fromCx: c.Cx, fromCy: c.Cy, fromZoom: c.Zoom,
		toCx: 0, toCy: 0, toZoom: 1,
		dur:  duration,
		ease: easing.InOutCubic,
	}
}

// camMove is the shared implementation behind ZoomTo / PanTo / Reset.
// Reading the start state in the factory (when the animation is
// constructed) rather than in Apply makes the animation idempotent —
// re-applying at t=0 returns to the same starting viewport every time.
type camMove struct {
	cam      *Camera
	fromCx   float64
	fromCy   float64
	fromZoom float64
	toCx     float64
	toCy     float64
	toZoom   float64
	dur      time.Duration
	ease     animation.EasingFunc
}

func (m *camMove) Apply(t float64) {
	t = clamp01(t)
	e := m.ease(t)
	m.cam.Cx = m.fromCx + (m.toCx-m.fromCx)*e
	m.cam.Cy = m.fromCy + (m.toCy-m.fromCy)*e
	m.cam.Zoom = m.fromZoom + (m.toZoom-m.fromZoom)*e
}
func (m *camMove) Duration() time.Duration    { return m.dur }
func (m *camMove) Targets() []mobject.Mobject { return nil } // camera targets nothing

// Focus is a composite: ZoomTo target at the given factor AND dim
// every other mobject to dimAmount. UnFocus reverses the dimming. The
// dim factor is stored on the Camera so the scene player can multiply
// each mobject's effective opacity at render time.
//
// Group-aware focusing: when target is a Group, all of the Group's
// children stay at full opacity. The scene player checks whether each
// mobject is the focus target OR a descendant of it.
func (c *Camera) Focus(target mobject.Mobject, factor float64, duration time.Duration) animation.Animation {
	return &focusAnim{
		cam:    c,
		target: target,
		factor: factor,
		// See attention.go Spotlight for the rationale on 0.4 vs 0.2.
		toDim: 0.4,
		dur:   duration,
		zoomAnim: &camMove{
			cam:    c,
			fromCx: c.Cx, fromCy: c.Cy, fromZoom: c.Zoom,
			toCx: 0, toCy: 0, toZoom: factor,
			dur:  duration,
			ease: easing.InOutCubic,
		},
	}
}

// FocusAt is like Focus but pans to an explicit (cx, cy) rather than
// to target's center. Phase-10 Fix 2 — DetailFocus uses this when a
// KeepVisible container is provided, so the camera can pan toward
// the target while staying within bounds that keep the container in
// view. The dim semantics are identical to Focus: target stays at
// full opacity, others dim.
func (c *Camera) FocusAt(target mobject.Mobject, cx, cy, factor float64, duration time.Duration) animation.Animation {
	return &focusAnim{
		cam:    c,
		target: target,
		factor: factor,
		toDim:  0.4,
		dur:    duration,
		// fixedCenter overrides the per-frame target-tracking pan so
		// the camera moves to the clamped (cx, cy) instead.
		fixedCx:    cx,
		fixedCy:    cy,
		fixedValid: true,
		zoomAnim: &camMove{
			cam:    c,
			fromCx: c.Cx, fromCy: c.Cy, fromZoom: c.Zoom,
			toCx: cx, toCy: cy, toZoom: factor,
			dur:  duration,
			ease: easing.InOutCubic,
		},
	}
}

// UnFocus releases the dim and returns zoom/pan to identity.
func (c *Camera) UnFocus(duration time.Duration) animation.Animation {
	return &focusAnim{
		cam:    c,
		target: nil, // unfocus
		factor: 1,
		toDim:  1,
		dur:    duration,
		zoomAnim: &camMove{
			cam:    c,
			fromCx: c.Cx, fromCy: c.Cy, fromZoom: c.Zoom,
			toCx: 0, toCy: 0, toZoom: 1,
			dur:  duration,
			ease: easing.InOutCubic,
		},
	}
}

type focusAnim struct {
	cam      *Camera
	target   mobject.Mobject
	factor   float64
	toDim    float64
	dur      time.Duration
	zoomAnim *camMove
	// fixedCenter (Phase-10 Fix 2) — when fixedValid is set, the
	// per-frame Apply uses (fixedCx, fixedCy) instead of tracking the
	// target's current position. Lets FocusAt clamp the camera so a
	// surrounding container remains in view.
	fixedCx, fixedCy float64
	fixedValid       bool
}

func (f *focusAnim) Apply(t float64) {
	t = clamp01(t)
	// Animate dim level alongside camera move. Read starting dim from
	// camera at Apply(0) so partial focus → re-focus blends naturally.
	if t == 0 {
		// Snapshot starting state. (Doing this on every t=0 call is fine
		// because Apply is idempotent.)
		f.zoomAnim.fromCx = f.cam.Cx
		f.zoomAnim.fromCy = f.cam.Cy
		f.zoomAnim.fromZoom = f.cam.Zoom
	}
	fromDim := f.cam.dimAmount
	// Compute zoom-target before the move apply so the easing uses the
	// final clamped position.
	if f.fixedValid {
		f.zoomAnim.toCx = f.fixedCx
		f.zoomAnim.toCy = f.fixedCy
	} else if f.target != nil {
		cx, cy := positionFor(f.target)
		f.zoomAnim.toCx = cx
		f.zoomAnim.toCy = cy
	}
	f.zoomAnim.Apply(t)
	// Linear dim transition is fine — the eye notices zoom timing, not
	// dim timing, so they share an easing curve via the zoom anim.
	e := easing.InOutCubic(t)
	f.cam.dimAmount = fromDim + (f.toDim-fromDim)*e
	if t >= 1 {
		f.cam.focusTarget = f.target
		f.cam.dimAmount = f.toDim
	}
}
func (f *focusAnim) Duration() time.Duration    { return f.dur }
func (f *focusAnim) Targets() []mobject.Mobject { return nil }

// FocusTarget reports the current focus target (nil when unfocused).
// Exposed for the scene player so it can decide which mobjects to dim.
func (c *Camera) FocusTarget() mobject.Mobject { return c.focusTarget }

// DimAmount reports the current opacity multiplier applied to non-
// focused mobjects (1.0 = no dim).
func (c *Camera) DimAmount() float64 {
	if c.dimAmount == 0 {
		return 1
	}
	return c.dimAmount
}

// Position returns the camera center in scene-space (the
// scene.Camera interface). Equivalent to reading (Cx, Cy).
func (c *Camera) Position() (float64, float64) { return c.Cx, c.Cy }

// ZoomLevel returns the current zoom factor (the scene.Camera
// interface). Equivalent to reading Zoom.
func (c *Camera) ZoomLevel() float64 {
	if c.Zoom == 0 {
		return 1
	}
	return c.Zoom
}

// positionFor extracts a center point from a mobject. Most mobjects
// implement Position(); for ones that don't, fall back to Bounds()
// center.
func positionFor(m mobject.Mobject) (float64, float64) {
	if p, ok := m.(interface{ Position() (float64, float64) }); ok {
		return p.Position()
	}
	c := m.Bounds().Center()
	return c.X, c.Y
}

func clamp01(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}
