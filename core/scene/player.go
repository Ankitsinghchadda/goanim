package scene

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// FrameWriter receives one image per scene frame from Play / PlayStill.
// The video encoder and "dump as PNGs" sinks both satisfy this.
type FrameWriter interface {
	WriteFrame(img image.Image) error
}

// FrameWriterFunc adapts an ordinary function to FrameWriter.
type FrameWriterFunc func(img image.Image) error

func (f FrameWriterFunc) WriteFrame(img image.Image) error { return f(img) }

// Play advances the scene clock by anim.Duration(), applying the
// animation at each frame and rendering all live mobjects.
//
// Animations are stateful in the sense that they mutate the targeted
// mobject(s) at each Apply call. The Player drives the clock; it does
// not snapshot state between frames.
//
// Play special-cases two animation shapes for throughput:
//
//   - animation.ConstantAnim (e.g. direction.Pause) renders one frame
//     and writes it N times rather than re-rendering identical
//     content. A 2-second pause skips ~119 redundant rasterizes.
//   - animation.SequenceLike (Sequence) is walked child-by-child so
//     any Pause INSIDE a Sequence gets the same fast path.
//
// Parallel and Stagger don't get this treatment because their
// children overlap in time — the "render once" optimization only
// applies when a span of time is owned exclusively by one
// zero-change animation.
func (s *Scene) Play(sink FrameWriter, anim animation.Animation) (int, error) {
	if s.Renderer == nil {
		return 0, fmt.Errorf("scene: Play requires a Renderer (set via WithRenderer)")
	}
	// Ensure all animation targets are in the live mobject set.
	for _, t := range anim.Targets() {
		s.ensureLive(t)
	}
	return s.playAny(sink, anim)
}

// playAny dispatches one animation through the right code path:
// constant-fast-path for Pause, sequence-walk for Sequence, generic
// per-frame loop otherwise.
func (s *Scene) playAny(sink FrameWriter, anim animation.Animation) (int, error) {
	if _, ok := anim.(animation.ConstantAnim); ok {
		return s.runConstantFrames(sink, anim.Duration())
	}
	if seq, ok := anim.(animation.SequenceLike); ok {
		// Snap any preceding Sequence children to their end state by
		// applying the WHOLE Sequence at t=1, then walk children one
		// at a time. The end-state snap is necessary because each
		// child's animation expects to start from where the previous
		// child left off (e.g. MoveTo captures the current position
		// in its factory).
		//
		// We use the original Sequence's Apply to drive the snap
		// because it knows how to handle sub-children that are
		// themselves composites.
		total := 0
		for _, child := range seq.Sequence() {
			for _, t := range child.Targets() {
				s.ensureLive(t)
			}
			n, err := s.playAny(sink, child)
			total += n
			if err != nil {
				return total, err
			}
		}
		return total, nil
	}
	return s.runFrames(sink, anim.Duration(), func(t float64) {
		anim.Apply(t)
	})
}

// runConstantFrames is the "scene state doesn't change" fast path. It
// renders ONE frame and emits the same image bytes for every tick of
// the duration. The caller's FrameWriter implementation is
// responsible for handling repeated identical writes — VideoEncoder
// writes them straight through to ffmpeg, which compresses them
// near-instantly because the inter-frame delta is zero.
func (s *Scene) runConstantFrames(sink FrameWriter, total time.Duration) (int, error) {
	if s.FPS <= 0 {
		s.FPS = 60
	}
	numFrames := int(float64(s.FPS) * total.Seconds())
	if numFrames < 1 && total > 0 {
		numFrames = 1
	}
	if numFrames == 0 {
		return 0, nil
	}
	// Render once.
	s.Renderer.BeginFrame(s.Width, s.Height, s.bgColor())
	s.applyCameraToRenderer()
	s.renderAllMobjects()
	img := s.Renderer.Image()
	for i := 0; i < numFrames; i++ {
		if err := sink.WriteFrame(img); err != nil {
			return i, fmt.Errorf("write frame %d: %w", i, err)
		}
	}
	return numFrames, nil
}

// PlayStill emits `duration` worth of frames with no animation. Useful
// for "hold the final state" sections at the end of a video.
func (s *Scene) PlayStill(sink FrameWriter, duration time.Duration) (int, error) {
	if s.Renderer == nil {
		return 0, fmt.Errorf("scene: PlayStill requires a Renderer")
	}
	return s.runFrames(sink, duration, func(float64) {})
}

func (s *Scene) ensureLive(m mobject.Mobject) {
	if m == nil {
		return
	}
	for _, x := range s.Mobjects {
		if x == m {
			return
		}
	}
	s.Mobjects = append(s.Mobjects, m)
}

func (s *Scene) runFrames(sink FrameWriter, total time.Duration, applyAt func(t float64)) (int, error) {
	if s.FPS <= 0 {
		s.FPS = 60
	}
	fps := s.FPS
	numFrames := int(float64(fps) * total.Seconds())
	if numFrames < 1 && total > 0 {
		numFrames = 1
	}
	for i := 0; i < numFrames; i++ {
		t := 1.0
		if numFrames > 1 {
			t = float64(i) / float64(numFrames-1)
		}
		applyAt(t)
		s.Renderer.BeginFrame(s.Width, s.Height, s.bgColor())
		s.applyCameraToRenderer()
		s.renderAllMobjects()
		if err := sink.WriteFrame(s.Renderer.Image()); err != nil {
			return i, fmt.Errorf("write frame %d: %w", i, err)
		}
	}
	return numFrames, nil
}

// applyCameraToRenderer pushes the current camera viewport into the
// renderer. Called once per frame, after BeginFrame (which resets the
// camera to identity).
func (s *Scene) applyCameraToRenderer() {
	if s.Camera == nil {
		return
	}
	cx, cy := s.Camera.Position()
	s.Renderer.SetCamera(cx, cy, s.Camera.ZoomLevel())
}

// renderAllMobjects walks the live mobject set and renders each one.
// When a camera focus is active, mobjects that aren't the focus target
// (or descendants of it) get an opacity multiplier on their style
// context — that's how Focus dims surroundings without mutating
// per-mobject Style.Opacity.
func (s *Scene) renderAllMobjects() {
	baseCtx := s.Context()
	dim, focusTarget := 1.0, mobject.Mobject(nil)
	if s.Camera != nil {
		dim = s.Camera.DimAmount()
		focusTarget = s.Camera.FocusTarget()
	}
	for _, m := range s.Mobjects {
		ctx := baseCtx
		if focusTarget != nil && dim < 1 && !sameOrDescendant(m, focusTarget) {
			ctx.OpacityMultiplier = dim
		} else {
			ctx.OpacityMultiplier = 0 // sentinel: no multiplier applied
		}
		m.Render(s.Renderer, ctx)
	}
}

// sameOrDescendant reports whether m is the focus target itself or
// a descendant of it (e.g., a child of a focused Group). Used by the
// dim logic to keep group children fully lit when their Group is the
// focus target.
func sameOrDescendant(m, target mobject.Mobject) bool {
	if m == target {
		return true
	}
	// Walk target's children if it's a Group-like aggregate.
	if children, ok := target.(interface {
		Children() []mobject.Mobject
	}); ok {
		for _, c := range children.Children() {
			if sameOrDescendant(m, c) {
				return true
			}
		}
	}
	return false
}

func (s *Scene) bgColor() color.Color {
	if s.BgColor == nil {
		return color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	}
	return s.BgColor
}
