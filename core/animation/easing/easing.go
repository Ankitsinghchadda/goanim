// Package easing provides curve functions used to shape animations.
//
// An easing function maps a normalized time t in [0, 1] to a
// normalized progress value (also typically in [0, 1], though Back
// and Spring overshoot intentionally).
//
// Choosing the right easing is a design decision. EaseOutCubic and
// EaseOutBack feel intentional for most "appearance" animations;
// reserve Linear for path traversal and motion-along; avoid InOut for
// short snappy effects (it feels mechanical).
package easing

import "math"

// Func is the easing signature.
type Func func(t float64) float64

// Linear returns t unchanged.
func Linear(t float64) float64 { return clamp01(t) }

// --- Quad (t^2) ---

func InQuad(t float64) float64    { t = clamp01(t); return t * t }
func OutQuad(t float64) float64   { t = clamp01(t); return 1 - (1-t)*(1-t) }
func InOutQuad(t float64) float64 { return inOut(t, InQuad, OutQuad) }

// --- Cubic (t^3) ---

func InCubic(t float64) float64    { t = clamp01(t); return t * t * t }
func OutCubic(t float64) float64   { t = clamp01(t); u := 1 - t; return 1 - u*u*u }
func InOutCubic(t float64) float64 { return inOut(t, InCubic, OutCubic) }

// --- Quart (t^4) ---

func InQuart(t float64) float64    { t = clamp01(t); return t * t * t * t }
func OutQuart(t float64) float64   { t = clamp01(t); u := 1 - t; return 1 - u*u*u*u }
func InOutQuart(t float64) float64 { return inOut(t, InQuart, OutQuart) }

// --- Expo ---

func InExpo(t float64) float64 {
	t = clamp01(t)
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*t-10)
}

func OutExpo(t float64) float64 {
	t = clamp01(t)
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

func InOutExpo(t float64) float64 { return inOut(t, InExpo, OutExpo) }

// --- Back (overshoot — great for snappy appearances) ---

const backOvershoot = 1.70158

func InBack(t float64) float64 {
	t = clamp01(t)
	c1 := backOvershoot
	c3 := c1 + 1
	return c3*t*t*t - c1*t*t
}

func OutBack(t float64) float64 {
	t = clamp01(t)
	c1 := backOvershoot
	c3 := c1 + 1
	u := t - 1
	return 1 + c3*u*u*u + c1*u*u
}

func InOutBack(t float64) float64 {
	t = clamp01(t)
	c1 := backOvershoot
	c2 := c1 * 1.525
	if t < 0.5 {
		return (math.Pow(2*t, 2) * ((c2+1)*2*t - c2)) / 2
	}
	return (math.Pow(2*t-2, 2)*((c2+1)*(2*t-2)+c2) + 2) / 2
}

// Spring returns an easing function modeling a critically- or
// under-damped harmonic oscillator. Parameters follow Framer Motion's
// convention: higher stiffness → faster motion; higher damping →
// less overshoot; higher mass → slower motion.
//
// A typical "bouncy spring" is Spring(180, 12, 1). A "snappy" spring
// is Spring(260, 26, 1). A "calm" spring is Spring(100, 20, 1).
//
// The returned function is normalized so f(0)=0 and f(1)≈1 (the
// oscillation has decayed sufficiently).
func Spring(stiffness, damping, mass float64) Func {
	if stiffness <= 0 {
		stiffness = 1
	}
	if damping <= 0 {
		damping = 0.0001
	}
	if mass <= 0 {
		mass = 1
	}
	// ω0 = sqrt(k/m), ζ = c / (2 sqrt(k m))
	omega0 := math.Sqrt(stiffness / mass)
	zeta := damping / (2 * math.Sqrt(stiffness*mass))
	const T = 1.0 // we map normalized t∈[0,1] to physical t = t * T

	return func(t float64) float64 {
		t = clamp01(t)
		x := t * T
		switch {
		case zeta < 1:
			// Underdamped
			omegaD := omega0 * math.Sqrt(1-zeta*zeta)
			env := math.Exp(-zeta * omega0 * x)
			osc := math.Cos(omegaD*x) + (zeta*omega0/omegaD)*math.Sin(omegaD*x)
			return 1 - env*osc
		case zeta == 1:
			// Critically damped
			return 1 - (1+omega0*x)*math.Exp(-omega0*x)
		default:
			// Overdamped
			r := math.Sqrt(zeta*zeta - 1)
			a := -omega0 * (zeta - r)
			b := -omega0 * (zeta + r)
			return 1 - ((math.Exp(a*x))*(zeta+r)-(math.Exp(b*x))*(zeta-r))/(2*r)
		}
	}
}

// --- helpers ---

func inOut(t float64, in, out Func) float64 {
	t = clamp01(t)
	if t < 0.5 {
		return in(2*t) / 2
	}
	return 0.5 + out(2*t-1)/2
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
