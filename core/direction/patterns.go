package direction

import (
	"fmt"
	"image"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/animation/easing"
	"github.com/ankitsinghchadda/goanim/core/layout"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Phase-9 scene patterns. Each pattern is an opinionated composition
// for a distinct kind of moment in an explainer video. The library
// handles typography, positioning, and animation defaults so authors
// don't have to. Each pattern takes a Spec struct that's mostly
// optional fields — the defaults are tuned to produce
// good-looking results without configuration.
//
// Patterns currently shipped:
//   - TitleSlide       — full-screen opener
//   - ChapterIntro     — section-break intro card
//   - EstablishingShot — "here's the system at a glance"
//   - DetailFocus      — zoom + dim to highlight one component
//
// Each pattern returns when the moment is done. Author chains them
// in sequence via successive calls.

// PlayFunc is the signature scene-patterns call into the player. The
// extracted indirection lets pattern code share boilerplate around
// Play(...) / hold + check err.
type playFunc func(animation.Animation) error

func makePlay(s *scene.Scene, sink scene.FrameWriter) playFunc {
	return func(a animation.Animation) error {
		_, err := s.Play(sink, a)
		return err
	}
}

// ----------------------------------------------------------------------------
// TitleSlide — full-screen opener
// ----------------------------------------------------------------------------

// TitleSlideSpec configures a TitleSlide pattern. Title is required;
// Subtitle and Motif are optional.
type TitleSlideSpec struct {
	Title    string
	Subtitle string
	// Motif, when set, animates in behind the title and recedes when
	// the slide ends (so the title is what the viewer remembers, not
	// the motif).
	Motif Motif
	// Duration is the total time the slide is on screen including
	// enter and exit. Default 5 seconds when zero.
	Duration time.Duration
}

// TitleSlide animates a full-screen title moment: motif enters (if
// provided), title scales in dramatically, subtitle fades in below,
// hold, then motif recedes / title fades.
func TitleSlide(s *scene.Scene, sink scene.FrameWriter, spec TitleSlideSpec) error {
	if spec.Duration == 0 {
		spec.Duration = 5 * time.Second
	}
	play := makePlay(s, sink)

	// Carve up the total duration:
	//   12% — entrance (motif + title scale-in)
	//   8%  — subtitle fade-in (parallel with rest of entrance)
	//   60% — hold
	//   20% — exit (title fade + motif recede)
	entranceD := durFrac(spec.Duration, 0.12)
	subFadeD := durFrac(spec.Duration, 0.08)
	holdD := durFrac(spec.Duration, 0.60)
	exitD := durFrac(spec.Duration, 0.20)

	// Title — RoleTitle from the hierarchy system. Centered slightly
	// above frame center so the subtitle has room below.
	title := mobject.NewText(0, spec.Title).MoveTo(0, 100).WithRole(style.RoleTitle)
	zero := 0.0
	st := *title.Style()
	st.Opacity = &zero
	title.SetStyle(st)
	s.Add(title)

	var subtitle *mobject.Text
	if spec.Subtitle != "" {
		subtitle = mobject.NewText(1, spec.Subtitle).MoveTo(0, -40).WithRole(style.RoleBody)
		st2 := *subtitle.Style()
		st2.Opacity = &zero
		subtitle.SetStyle(st2)
		s.Add(subtitle)
	}

	// Motif (optional) — added FIRST so it renders behind the text.
	if spec.Motif != nil {
		s.Add(spec.Motif)
		// Bring the motif in slightly before the title's scale-in so the
		// title appears to "land" on top of it.
		if err := play(spec.Motif.EnterAnimation(entranceD)); err != nil {
			return err
		}
	}

	// Title pops in. Use a parallel of FadeIn (for opacity) + a small
	// custom scale anim if we have one; FadeIn alone gives a clean
	// presence on its own.
	titleAnims := []animation.Animation{
		animation.FadeIn(title, entranceD),
	}
	if subtitle != nil {
		// Subtitle starts 50% into the entrance so the title lands first.
		titleAnims = append(titleAnims, animation.Sequence(
			direction(durFrac(spec.Duration, 0.06)),
			animation.FadeIn(subtitle, subFadeD),
		))
	}
	if err := play(animation.Parallel(titleAnims...)); err != nil {
		return err
	}

	// Hold.
	if err := play(Pause(holdD)); err != nil {
		return err
	}

	// Exit: title + subtitle fade out. Motif recedes.
	exits := []animation.Animation{
		animation.FadeOut(title, exitD),
	}
	if subtitle != nil {
		exits = append(exits, animation.FadeOut(subtitle, exitD))
	}
	if spec.Motif != nil {
		exits = append(exits, spec.Motif.ExitAnimation(ExitDissolve, exitD))
	}
	if err := play(animation.Parallel(exits...)); err != nil {
		return err
	}

	// Remove mobjects so they don't leak into next scene.
	if subtitle != nil {
		s.Remove(subtitle)
	}
	s.Remove(title)
	if spec.Motif != nil {
		s.Remove(spec.Motif)
	}
	return nil
}

// ----------------------------------------------------------------------------
// ChapterIntro — section-break card
// ----------------------------------------------------------------------------

// ChapterSpec configures a ChapterIntro.
type ChapterSpec struct {
	Number   int    // shows as e.g. "01" or "Chapter 1"
	Title    string // chapter title
	Subtitle string // optional short description
	Duration time.Duration
}

// ChapterIntro renders a section-break with a large chapter number
// on the left and the chapter title/subtitle stacked on the right.
func ChapterIntro(s *scene.Scene, sink scene.FrameWriter, spec ChapterSpec) error {
	if spec.Duration == 0 {
		spec.Duration = 4 * time.Second
	}
	play := makePlay(s, sink)

	entranceD := durFrac(spec.Duration, 0.20)
	holdD := durFrac(spec.Duration, 0.55)
	exitD := durFrac(spec.Duration, 0.25)

	// Big chapter number left of center, aligned to title baseline at
	// y=100. Phase-10 polish — number and title share a baseline.
	// Phase-10b — instead of placing the title at a fixed x=180 (which
	// produced "01The Six Systems" overlap with the wider Excalifont
	// metrics), measure the chapter number's actual bounds and place
	// the title's left edge to the right of it with a consistent gap.
	// Position the composite (num + title) so it remains centered on
	// the canvas.
	numText := fmt.Sprintf("%02d", spec.Number)
	num := mobject.NewText(0, numText).MoveTo(0, 100).WithRole(style.RoleTitle)
	zero := 0.0
	stN := *num.Style()
	stN.Opacity = &zero
	num.SetStyle(stN)
	s.Add(num)

	title := mobject.NewText(1, spec.Title).MoveTo(0, 100).WithRole(style.RoleHeading)
	stT := *title.Style()
	stT.Opacity = &zero
	title.SetStyle(stT)
	s.Add(title)

	// Measure font metrics of both pieces, then place them with a
	// 100-px gap between them and center the composite at x=0.
	const interGap = 100.0
	nb := num.Bounds()
	tb := title.Bounds()
	numW := nb.Max.X - nb.Min.X
	titleW := tb.Max.X - tb.Min.X
	totalW := numW + interGap + titleW
	leftEdge := -totalW / 2
	num.SetPosition(leftEdge+numW/2, 100)
	title.SetPosition(leftEdge+numW+interGap+titleW/2, 100)

	// Subtitle drops well below the chapter number's bottom so the
	// huge RoleTitle "02" glyph doesn't slash through the body text,
	// and is centered horizontally on the canvas (was anchored to the
	// title's old fixed x=180, now centered for symmetry under the
	// dynamically-laid-out num/title row).
	var sub *mobject.Text
	if spec.Subtitle != "" {
		sub = mobject.NewText(2, spec.Subtitle).MoveTo(0, -160).WithRole(style.RoleBody)
		stS := *sub.Style()
		stS.Opacity = &zero
		sub.SetStyle(stS)
		s.Add(sub)
	}

	// Entrance: number first (a beat alone), then title + subtitle.
	if err := play(animation.FadeIn(num, durFrac(spec.Duration, 0.10))); err != nil {
		return err
	}
	titleAnims := []animation.Animation{
		animation.FadeIn(title, entranceD),
	}
	if sub != nil {
		titleAnims = append(titleAnims, animation.Sequence(
			Pause(durFrac(spec.Duration, 0.06)),
			animation.FadeIn(sub, durFrac(spec.Duration, 0.10)),
		))
	}
	if err := play(animation.Parallel(titleAnims...)); err != nil {
		return err
	}

	if err := play(Pause(holdD)); err != nil {
		return err
	}

	// Exit: everything fades together.
	exits := []animation.Animation{
		animation.FadeOut(num, exitD),
		animation.FadeOut(title, exitD),
	}
	if sub != nil {
		exits = append(exits, animation.FadeOut(sub, exitD))
	}
	if err := play(animation.Parallel(exits...)); err != nil {
		return err
	}

	s.Remove(num)
	s.Remove(title)
	if sub != nil {
		s.Remove(sub)
	}
	return nil
}

// ----------------------------------------------------------------------------
// EstablishingShot — "the system at a glance"
// ----------------------------------------------------------------------------

// EstablishingShotSpec configures an EstablishingShot. Components is
// the list of mobjects to lay out and reveal. Caption is the title
// text above/below them.
type EstablishingShotSpec struct {
	Components []mobject.Mobject
	Caption    string
	// CaptionAbove places the caption above the row of components.
	// Default (false) places it below.
	CaptionAbove bool
	// SafeWidth overrides the layout's auto-fit width. Default is
	// layout.DefaultSafeWidth (1728 = 1920 minus 5% on each side).
	SafeWidth float64
	Duration  time.Duration
}

// EstablishingShot lays out the given components in a horizontal row
// with auto-fit-to-width, then stagger-fades them in alongside a
// composed caption. The hold-time is the bulk of the duration so the
// viewer can absorb the layout.
func EstablishingShot(s *scene.Scene, sink scene.FrameWriter, spec EstablishingShotSpec) error {
	if spec.Duration == 0 {
		spec.Duration = 6 * time.Second
	}
	if len(spec.Components) == 0 {
		return fmt.Errorf("EstablishingShot: Components is required")
	}
	play := makePlay(s, sink)

	// Carve up the duration:
	//   25% — entrance stagger
	//   60% — hold
	//   15% — exit
	entranceTotal := durFrac(spec.Duration, 0.25)
	holdD := durFrac(spec.Duration, 0.60)
	exitD := durFrac(spec.Duration, 0.15)

	// Lay out the components via HBox with FitContent (Phase-10 Fix 4:
	// scale up modestly when the natural width is smaller than the safe
	// target, capped at 2.5×; otherwise scale down to fit). Spacing is
	// tight (40 px) so 5-6 component scenes still have room to grow,
	// not just 2-3 component scenes. Start the row at canvas-center y;
	// we'll re-center the (row+caption) composition vertically below.
	row := layout.NewHBox(spec.Components...).WithSpacing(40).Fit(layout.FitContent)
	if spec.SafeWidth > 0 {
		row.WithSafeWidth(spec.SafeWidth)
	}
	row.MoveTo(0, 0)
	s.Add(row)
	_ = row.Bounds() // trigger fit-layout pass

	// Set components hidden initially so FadeIn can ramp them.
	zero := 0.0
	for _, c := range spec.Components {
		st := *c.Style()
		st.Opacity = &zero
		c.SetStyle(st)
	}

	// Caption — RoleCaption. Tentatively positioned below (or above)
	// the row by a fixed margin; we'll re-center the composition after.
	var cap *mobject.Text
	const captionMargin = 80.0 // gap between row edge and caption baseline
	if spec.Caption != "" {
		rb := row.Bounds()
		var cy float64
		if spec.CaptionAbove {
			cy = rb.Max.Y + captionMargin
		} else {
			cy = rb.Min.Y - captionMargin
		}
		cap = mobject.NewText(0, spec.Caption).MoveTo(0, cy).WithRole(style.RoleCaption)
		stC := *cap.Style()
		stC.Opacity = &zero
		cap.SetStyle(stC)
		s.Add(cap)
	}

	// Phase-10 Fix 1: vertically center the (row + caption) composition
	// within the canvas safe area. After FitContent has set component
	// scales, measure the composite bbox and translate everything so
	// its vertical midpoint sits at scene y=0.
	{
		composite := row.Bounds()
		if cap != nil {
			composite = composite.Union(cap.Bounds())
		}
		dy := -composite.Center().Y
		if dy != 0 {
			rx, ry := row.Position()
			row.MoveTo(rx, ry+dy)
			_ = row.Bounds()
			if cap != nil {
				cpx, cpy := cap.Position()
				cap.MoveTo(cpx, cpy+dy)
			}
		}
	}

	// Stagger fade-ins. Per-child duration is short; stagger delay is
	// chosen so total = entranceTotal.
	perChildD := durFrac(spec.Duration, 0.08)
	staggerStep := time.Duration(0)
	if len(spec.Components) > 1 {
		staggerStep = (entranceTotal - perChildD) / time.Duration(len(spec.Components)-1)
	}
	fadeIns := make([]animation.Animation, 0, len(spec.Components))
	for _, c := range spec.Components {
		fadeIns = append(fadeIns, animation.FadeIn(c, perChildD))
	}
	staggered := animation.Stagger(staggerStep, fadeIns...)
	if cap != nil {
		// Caption fades in alongside the LAST third of the stagger.
		captionFade := animation.Sequence(
			Pause(entranceTotal*2/3),
			animation.FadeIn(cap, entranceTotal/3),
		)
		if err := play(animation.Parallel(staggered, captionFade)); err != nil {
			return err
		}
	} else {
		if err := play(staggered); err != nil {
			return err
		}
	}

	if err := play(Pause(holdD)); err != nil {
		return err
	}

	exits := make([]animation.Animation, 0, len(spec.Components)+1)
	for _, c := range spec.Components {
		exits = append(exits, animation.FadeOut(c, exitD))
	}
	if cap != nil {
		exits = append(exits, animation.FadeOut(cap, exitD))
	}
	if err := play(animation.Parallel(exits...)); err != nil {
		return err
	}

	for _, c := range spec.Components {
		s.Remove(c)
	}
	s.Remove(row)
	if cap != nil {
		s.Remove(cap)
	}
	return nil
}

// ----------------------------------------------------------------------------
// DetailFocus — zoom + dim + caption
// ----------------------------------------------------------------------------

// DetailFocusSpec configures a DetailFocus moment.
type DetailFocusSpec struct {
	// Target is the mobject to zoom onto. Must already be in the scene
	// (e.g. added via a prior EstablishingShot). Camera must be set on
	// the scene.
	Target mobject.Mobject
	// Caption is the explanatory text shown alongside the focused
	// component. Positioned below the target with consistent margin.
	Caption string
	// ZoomFactor is the camera zoom. Default 1.5×.
	ZoomFactor float64
	// KeepVisible, when set, is the bounding mobject (typically the
	// surrounding layout / row) whose entire extent must remain in
	// view at the chosen zoom. Phase-10 Fix 2: DetailFocus uses this
	// to clamp the camera and conservatively reduce zoom so edge
	// components in the surrounding layout aren't clipped.
	//
	// Recentering primary, conservative-zoom fallback:
	//   1. shift the camera toward Target, but only as far as still
	//      keeps KeepVisible.Bounds() inside the visible frame at the
	//      requested zoom; then
	//   2. if even centering on the layout center can't fit at the
	//      requested zoom, reduce the zoom until it does.
	KeepVisible mobject.Mobject
	Duration    time.Duration
}

// DetailFocus zooms the camera onto Target, dims surroundings, shows
// a caption explaining the component, holds, then releases.
func DetailFocus(s *scene.Scene, sink scene.FrameWriter, cam *Camera, spec DetailFocusSpec) error {
	if spec.Duration == 0 {
		spec.Duration = 4 * time.Second
	}
	if spec.ZoomFactor == 0 {
		spec.ZoomFactor = 1.5
	}
	if spec.Target == nil {
		return fmt.Errorf("DetailFocus: Target is required")
	}
	if cam == nil {
		return fmt.Errorf("DetailFocus: Camera is required")
	}
	play := makePlay(s, sink)

	enterD := durFrac(spec.Duration, 0.18)
	holdD := durFrac(spec.Duration, 0.64)
	exitD := durFrac(spec.Duration, 0.18)

	// Phase-10 Fix 2 — when KeepVisible is set, clamp zoom and camera
	// center so the surrounding layout stays in view at the chosen
	// zoom. Without this, zooming 1.7× on a left-side component shifts
	// the right edge off-canvas.
	safeFactor := spec.ZoomFactor
	safeCx, safeCy := positionFor(spec.Target)
	if spec.KeepVisible != nil {
		kb := spec.KeepVisible.Bounds()
		kw := kb.Width()
		kh := kb.Height()
		canvasW := float64(s.Width)
		canvasH := float64(s.Height)
		// Conservative zoom: don't allow the keep-visible bbox to be
		// wider/taller than the visible frame at this zoom. Visible
		// frame at zoom z = (canvasW/z, canvasH/z). To fit kw/kh:
		// z ≤ canvasW/kw and z ≤ canvasH/kh.
		const safePad = 1.05 // leave a small visual margin around the bbox
		maxZx := canvasW / (kw * safePad)
		maxZy := canvasH / (kh * safePad)
		maxZ := maxZx
		if maxZy < maxZ {
			maxZ = maxZy
		}
		if safeFactor > maxZ {
			safeFactor = maxZ
		}
		if safeFactor < 1.0 {
			safeFactor = 1.0
		}
		// Recenter primary: clamp the desired camera center so the
		// keep-visible bbox stays within the visible frame.
		halfW := (canvasW / safeFactor) / 2
		halfH := (canvasH / safeFactor) / 2
		// Cx must satisfy [kbMaxX - halfW, kbMinX + halfW]. If interval
		// is empty (keep-visible doesn't fit) we already lowered
		// safeFactor enough above; clamp anyway.
		loX := kb.Max.X - halfW
		hiX := kb.Min.X + halfW
		if loX > hiX {
			// Doesn't fit even after clamp; center on bbox center.
			loX = (kb.Min.X + kb.Max.X) / 2
			hiX = loX
		}
		if safeCx < loX {
			safeCx = loX
		}
		if safeCx > hiX {
			safeCx = hiX
		}
		loY := kb.Max.Y - halfH
		hiY := kb.Min.Y + halfH
		if loY > hiY {
			loY = (kb.Min.Y + kb.Max.Y) / 2
			hiY = loY
		}
		if safeCy < loY {
			safeCy = loY
		}
		if safeCy > hiY {
			safeCy = hiY
		}
	}

	// Caption appears below the target. Position it relative to the
	// target's bounds, but — Phase-10 polish — when KeepVisible is set
	// we center the caption x on the KeepVisible row instead of on the
	// target, so a caption under a left-side or right-side target
	// doesn't spill off the canvas edge.
	var cap *mobject.Text
	if spec.Caption != "" {
		tb := spec.Target.Bounds()
		capX := (tb.Min.X + tb.Max.X) / 2
		if spec.KeepVisible != nil {
			kb := spec.KeepVisible.Bounds()
			capX = (kb.Min.X + kb.Max.X) / 2
		}
		cap = mobject.NewText(0, spec.Caption).
			MoveTo(capX, tb.Min.Y-30).
			WithRole(style.RoleCaption)
		zero := 0.0
		stC := *cap.Style()
		stC.Opacity = &zero
		cap.SetStyle(stC)
		s.Add(cap)
	}

	// Focus is camera move (zoom + pan to safeCx/safeCy) + dim others.
	if err := play(animation.Parallel(
		cam.FocusAt(spec.Target, safeCx, safeCy, safeFactor, enterD),
		func() animation.Animation {
			if cap == nil {
				return Pause(enterD)
			}
			return animation.Sequence(
				Pause(enterD/2),
				animation.FadeIn(cap, enterD/2),
			)
		}(),
	)); err != nil {
		return err
	}

	if err := play(Pause(holdD)); err != nil {
		return err
	}

	// Exit: release dim + reset camera + fade caption.
	capOut := animation.Animation(Pause(exitD))
	if cap != nil {
		capOut = animation.FadeOut(cap, exitD)
	}
	if err := play(animation.Parallel(
		cam.UnFocus(exitD),
		capOut,
	)); err != nil {
		return err
	}

	if cap != nil {
		s.Remove(cap)
	}
	return nil
}

// ----------------------------------------------------------------------------
// helpers
// ----------------------------------------------------------------------------

// durFrac returns frac of total as a time.Duration. Used by patterns
// to carve up their total runtime into entrance / hold / exit phases.
func durFrac(total time.Duration, frac float64) time.Duration {
	return time.Duration(float64(total) * frac)
}

// direction is a tiny constant-anim wrapper used in TitleSlide's
// staggered sub-animation. Equivalent to direction.Pause(d) but
// scoped to this file to avoid the import shadow.
func direction(d time.Duration) animation.Animation { return Pause(d) }

// Unused-symbol guard — keeps the imports referenced for go vet.
var _ image.Image = nil

// Keep the easing import referenced for patterns that may inline
// tweens not already used by the base animation library.
var _ animation.EasingFunc = easing.OutCubic
