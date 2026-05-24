package animation

import (
	"time"

	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// Parallel runs all child animations concurrently for the duration of
// the longest. Each child's t parameter is independently scaled to its
// own duration.
func Parallel(anims ...Animation) Animation {
	var maxDur time.Duration
	for _, a := range anims {
		if a.Duration() > maxDur {
			maxDur = a.Duration()
		}
	}
	return &parallelAnim{anims: anims, dur: maxDur}
}

type parallelAnim struct {
	anims []Animation
	dur   time.Duration
}

func (p *parallelAnim) Apply(t float64) {
	t = clamp01(t)
	scenePos := time.Duration(t * float64(p.dur))
	for _, a := range p.anims {
		var local float64
		if a.Duration() == 0 {
			local = clamp01(t)
		} else {
			local = float64(scenePos) / float64(a.Duration())
			if local > 1 {
				local = 1
			}
		}
		a.Apply(local)
	}
}
func (p *parallelAnim) Duration() time.Duration { return p.dur }
func (p *parallelAnim) Targets() []mobject.Mobject {
	var out []mobject.Mobject
	seen := map[mobject.Mobject]bool{}
	for _, a := range p.anims {
		for _, m := range a.Targets() {
			if !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
	}
	return out
}

// Sequence runs animations one after another. Total duration is the
// sum of children.
func Sequence(anims ...Animation) Animation {
	var total time.Duration
	for _, a := range anims {
		total += a.Duration()
	}
	return &sequenceAnim{anims: anims, dur: total}
}

// Children is the Phase-8 hook the scene player uses to walk a
// Sequence's segments and apply ConstantAnim fast-pathing to each
// pause segment independently. Exposed via the SequenceLike
// interface; callers shouldn't depend on the concrete type.
type sequenceAnim struct {
	anims []Animation
	dur   time.Duration
}

// SequenceLike is implemented by sequential composites. The player
// uses this to walk top-level Sequence children and special-case
// any Pause segments. Parallel and Stagger intentionally do NOT
// implement this — their children overlap in time so the
// "render once per child" trick doesn't apply.
type SequenceLike interface {
	Animation
	Sequence() []Animation
}

func (s *sequenceAnim) Sequence() []Animation { return s.anims }

func (s *sequenceAnim) Apply(t float64) {
	t = clamp01(t)
	if s.dur == 0 {
		// Snap every child to its end state.
		for _, a := range s.anims {
			a.Apply(1)
		}
		return
	}
	scenePos := time.Duration(t * float64(s.dur))
	elapsed := time.Duration(0)
	for i, a := range s.anims {
		end := elapsed + a.Duration()
		if scenePos < end || i == len(s.anims)-1 {
			var local float64
			if a.Duration() == 0 {
				local = 1
			} else {
				local = float64(scenePos-elapsed) / float64(a.Duration())
				if local < 0 {
					local = 0
				}
				if local > 1 {
					local = 1
				}
			}
			a.Apply(local)
			// Snap earlier children to their end state so visual
			// invariants (final positions, final opacities) hold.
			for j := 0; j < i; j++ {
				s.anims[j].Apply(1)
			}
			return
		}
		elapsed = end
	}
}
func (s *sequenceAnim) Duration() time.Duration { return s.dur }
func (s *sequenceAnim) Targets() []mobject.Mobject {
	var out []mobject.Mobject
	seen := map[mobject.Mobject]bool{}
	for _, a := range s.anims {
		for _, m := range a.Targets() {
			if !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
	}
	return out
}

// Stagger runs each animation with the given inter-start delay. Child
// i starts at i*delay and ends at i*delay + anims[i].Duration(). The
// total duration is the maximum end-time across all children.
//
// For same-duration children this collapses to delay*(n-1) +
// child.Duration() (the common case). For mixed-duration children it
// correctly reports the longest finisher, which matters when callers
// compose a Stagger inside a Sequence or Parallel.
func Stagger(delay time.Duration, anims ...Animation) Animation {
	if len(anims) == 0 {
		return Wait(0)
	}
	var totalDur time.Duration
	for i, a := range anims {
		end := time.Duration(i)*delay + a.Duration()
		if end > totalDur {
			totalDur = end
		}
	}
	return &staggerAnim{anims: anims, delay: delay, dur: totalDur}
}

type staggerAnim struct {
	anims []Animation
	delay time.Duration
	dur   time.Duration
}

func (s *staggerAnim) Apply(t float64) {
	t = clamp01(t)
	scenePos := time.Duration(t * float64(s.dur))
	for i, a := range s.anims {
		startAt := time.Duration(i) * s.delay
		if scenePos < startAt {
			a.Apply(0)
			continue
		}
		elapsed := scenePos - startAt
		var local float64
		if a.Duration() == 0 {
			local = 1
		} else {
			local = float64(elapsed) / float64(a.Duration())
			if local > 1 {
				local = 1
			}
		}
		a.Apply(local)
	}
}
func (s *staggerAnim) Duration() time.Duration { return s.dur }
func (s *staggerAnim) Targets() []mobject.Mobject {
	var out []mobject.Mobject
	seen := map[mobject.Mobject]bool{}
	for _, a := range s.anims {
		for _, m := range a.Targets() {
			if !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
	}
	return out
}
