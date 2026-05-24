package rough

import (
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
)

// RoughLine returns a Path that draws a single hand-drawn line between
// p1 and p2.
//
// The result is the rough.js "double-stroke": two slightly-bowed cubic
// Bezier curves drawn over the same segment, the second with half the
// jitter magnitude. The two passes share a single PRNG stream so the
// stroke is reproducible for a given Seed.
//
// Set Options.DisableMultiStroke to draw only one pass.
func RoughLine(p1, p2 geometry.Point, opts Options) *geometry.Path {
	r := newRNG(opts.Seed)
	out := geometry.NewPath()
	roughLineInto(out, p1, p2, opts, r, false)
	if !opts.DisableMultiStroke {
		roughLineInto(out, p1, p2, opts, r, true)
	}
	return out
}

// roughLineInto appends a single rough-line subpath to out. The overlay
// flag selects the second pass (half-magnitude jitter).
//
// The PRNG draw order is load-bearing for determinism — every change
// here is a behavior change. Order:
//
//  1. divergePoint (1 draw)
//  2. midDispX, midDispY (2 draws)
//  3. MOVE jitter x, MOVE jitter y (2 draws) — skipped if preserveVertices
//  4. cp1x, cp1y (2 draws)
//  5. cp2x, cp2y (2 draws)
//  6. endX, endY (2 draws) — skipped if preserveVertices
func roughLineInto(out *geometry.Path, p1, p2 geometry.Point, o Options, r *rng, overlay bool) {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	lengthSq := dx*dx + dy*dy
	length := math.Sqrt(lengthSq)

	// Length-based roughness damping — long lines look stiffer.
	var roughGain float64
	switch {
	case length < 200:
		roughGain = 1
	case length > 500:
		roughGain = 0.4
	default:
		roughGain = -0.0016668*length + 1.233334
	}

	// Cap the per-point jitter so it doesn't dominate short lines.
	offset := o.MaxRandomnessOffset
	if offset*offset*100 > lengthSq {
		offset = length / 10
	}
	halfOffset := offset / 2

	// asymmetric midpoint along the line
	divergePoint := 0.2 + r.next()*0.2

	// "Bowing": rotate the line vector 90° (in JS-y-down coords this is
	// (y2-y1, x1-x2)) and scale by bowing * MaxRandomnessOffset / 200,
	// then jitter both components.
	midDispX := o.Bowing * o.MaxRandomnessOffset * (p2.Y - p1.Y) / 200
	midDispY := o.Bowing * o.MaxRandomnessOffset * (p1.X - p2.X) / 200
	midDispX = offsetSym(midDispX, o.Roughness, roughGain, r)
	midDispY = offsetSym(midDispY, o.Roughness, roughGain, r)

	jit := offset
	if overlay {
		jit = halfOffset
	}

	startX, startY := p1.X, p1.Y
	// When PreserveVertices is set, leave the endpoint exact (no
	// jitter). rough.js does the same — and crucially does NOT consume
	// any random draws in that branch, so the PRNG state stays aligned
	// across implementations.
	if !o.PreserveVertices {
		startX += offsetSym(jit, o.Roughness, roughGain, r)
		startY += offsetSym(jit, o.Roughness, roughGain, r)
	}
	out.MoveTo(startX, startY)

	c1x := midDispX + p1.X + dx*divergePoint + offsetSym(jit, o.Roughness, roughGain, r)
	c1y := midDispY + p1.Y + dy*divergePoint + offsetSym(jit, o.Roughness, roughGain, r)
	c2x := midDispX + p1.X + 2*dx*divergePoint + offsetSym(jit, o.Roughness, roughGain, r)
	c2y := midDispY + p1.Y + 2*dy*divergePoint + offsetSym(jit, o.Roughness, roughGain, r)

	endX, endY := p2.X, p2.Y
	if !o.PreserveVertices {
		endX += offsetSym(jit, o.Roughness, roughGain, r)
		endY += offsetSym(jit, o.Roughness, roughGain, r)
	}
	out.CurveTo(c1x, c1y, c2x, c2y, endX, endY)
}

// offsetSym returns a random value in [-x, +x] * roughness * roughGain.
// (rough.js: _offsetOpt(x, ops, gain).)
func offsetSym(x, roughness, roughGain float64, r *rng) float64 {
	if roughness == 0 {
		return 0
	}
	return roughness * roughGain * (r.next()*(2*x) - x)
}

// offsetRange returns a random value in [min, max] * roughness * roughGain.
// (rough.js: _offset(min, max, ops, gain).)
func offsetRange(min, max, roughness, roughGain float64, r *rng) float64 {
	if roughness == 0 {
		return (min + max) / 2
	}
	return roughness * roughGain * (r.next()*(max-min) + min)
}
