package direction

import (
	"math"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/animation/easing"
	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// Pulse rhythmically scales the target between 1.0 and 1.0+amplitude,
// doing `count` full pulses across `duration`. Used to draw attention
// to a component without moving the diagram around.
//
// Default amplitude is 0.10 (gentle 10% pulse). The motion follows a
// sine wave so the pulse breathes smoothly rather than ping-ponging.
//
//	scene.Play(direction.Pulse(database, 3, 1500*time.Millisecond))
//
// Composition: Pulse calls SetVisualScale on the target — independent
// of opacity, position, and camera. It composes with FadeIn, MoveTo,
// camera moves, and spotlights without conflict.
func Pulse(target mobject.Mobject, count int, duration time.Duration) animation.Animation {
	return PulseWithAmplitude(target, count, duration, 0.10)
}

// PulseWithAmplitude is Pulse with a custom peak scale factor —
// 0.10 = 10% bigger at peak, 0.25 = 25%, etc. Above 0.5 the pulse
// reads as a bounce rather than a heartbeat.
func PulseWithAmplitude(target mobject.Mobject, count int, duration time.Duration, amplitude float64) animation.Animation {
	if count < 1 {
		count = 1
	}
	return &pulseAnim{
		target:    target,
		count:     count,
		amplitude: amplitude,
		dur:       duration,
	}
}

type pulseAnim struct {
	target    mobject.Mobject
	count     int
	amplitude float64
	dur       time.Duration
}

func (p *pulseAnim) Apply(t float64) {
	t = clamp01(t)
	// One pulse = one full sine cycle (0 → peak → 0). For `count`
	// pulses across [0, 1], the angle sweeps count * 2π.
	angle := t * float64(p.count) * 2 * math.Pi
	// sin starts at 0, rises to 1 at π/2, back to 0 at π, down to -1
	// at 3π/2, back to 0 at 2π. We want the scale to GO UP from 1.0
	// during the first half and back to 1.0 at the end — never below
	// 1.0. So use the absolute value of sin.
	scale := 1 + p.amplitude*math.Abs(math.Sin(angle))
	if s, ok := p.target.(interface{ SetVisualScale(float64) }); ok {
		s.SetVisualScale(scale)
	}
}
func (p *pulseAnim) Duration() time.Duration    { return p.dur }
func (p *pulseAnim) Targets() []mobject.Mobject { return []mobject.Mobject{p.target} }

// Spotlight dims every mobject except the target over duration. Pass
// nil as the target (or call RemoveSpotlight) to release the dim
// over duration.
//
//	scene.Play(direction.Spotlight(cam, server, 800*time.Millisecond))
//	scene.Play(direction.Pause(2*time.Second))
//	scene.Play(direction.Spotlight(cam, nil, 800*time.Millisecond))
//
// Spotlight reuses the Camera's dim machinery — the same one
// Camera.Focus uses — so it composes cleanly with active camera
// moves. A Spotlight during a Camera.ZoomTo works without conflict:
// the camera tracks the viewport, the spotlight tracks attention.
//
// Note that Spotlight needs the scene's Camera to operate (that's
// what holds the focus-target / dim-amount state read every frame).
// If your scene has no Camera attached, Spotlight is a no-op.
func Spotlight(cam *Camera, target mobject.Mobject, duration time.Duration) animation.Animation {
	if cam == nil {
		return &spotlightAnim{cam: nil, target: target, dur: duration}
	}
	return &spotlightAnim{
		cam:       cam,
		fromDim:   cam.DimAmount(),
		fromFocus: cam.FocusTarget(),
		target:    target,
		dur:       duration,
		ease:      easing.InOutCubic,
	}
}

// RemoveSpotlight is shorthand for Spotlight(cam, nil, duration) —
// reads more clearly at call sites where the intent is "release."
func RemoveSpotlight(cam *Camera, duration time.Duration) animation.Animation {
	return Spotlight(cam, nil, duration)
}

type spotlightAnim struct {
	cam       *Camera
	fromDim   float64
	fromFocus mobject.Mobject
	target    mobject.Mobject
	dur       time.Duration
	ease      animation.EasingFunc
}

func (s *spotlightAnim) Apply(t float64) {
	if s.cam == nil {
		return
	}
	t = clamp01(t)
	var toDim float64
	if s.target == nil {
		toDim = 1
	} else {
		// 0.4 is gentler than the prompt's suggested 0.2 — the
		// renderer's anti-aliased thin strokes composited at very low
		// alpha against the page color produce chromatic shifts
		// (especially visible on Cartoonist cross-hatch). 0.4 still
		// reads as "the spotlight item is clearly the subject" while
		// avoiding the worst of the color artifacts.
		toDim = 0.4
	}
	e := easing.InOutCubic
	if s.ease != nil {
		e = s.ease
	}
	f := e(t)
	s.cam.dimAmount = s.fromDim + (toDim-s.fromDim)*f
	// Set the focus target at the BEGINNING of the animation if we're
	// applying a spotlight (so the dim takes effect immediately on the
	// non-target). On removal, only clear at the end so the rest of
	// the animation still sees a focus target.
	if t == 0 {
		// snapshot in case Apply(0) is called more than once
		s.fromDim = s.cam.dimAmount
		s.fromFocus = s.cam.focusTarget
	}
	if t >= 1 {
		s.cam.focusTarget = s.target
		s.cam.dimAmount = toDim
	} else if s.target != nil {
		// Applying — make sure the target is recognized as focus from
		// the start of the animation so dimming applies to others.
		s.cam.focusTarget = s.target
	}
}
func (s *spotlightAnim) Duration() time.Duration    { return s.dur }
func (s *spotlightAnim) Targets() []mobject.Mobject { return nil }
