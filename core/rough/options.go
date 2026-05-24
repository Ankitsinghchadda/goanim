package rough

import "image/color"

// FillStyle names a fill pattern.
type FillStyle uint8

const (
	// FillNone disables fill.
	FillNone FillStyle = iota
	// FillSolid is a flat color with no roughness.
	FillSolid
	// FillHatch is parallel slanted lines clipped to the shape.
	FillHatch
	// FillCrossHatch is two FillHatch layers at 90° to each other.
	FillCrossHatch
	// FillZigzag is back-and-forth lines like a child's coloring.
	FillZigzag
	// FillDots is a stippled-dot pattern.
	FillDots
)

// Options control the visual output of every rough primitive.
//
// Defaults — DefaultOptions returns a struct configured to match
// Excalidraw's out-of-the-box look. Callers may override any field.
type Options struct {
	// Seed feeds the deterministic PRNG. Same seed + same parameters →
	// byte-identical output, every time. Required to be non-zero.
	Seed int64

	// Roughness scales the random perturbation applied to every point.
	// 0 = perfectly straight; 1 = Excalidraw default; 2 = very rough.
	Roughness float64

	// Bowing controls how much "lines" deviate from straight at their
	// midpoint. 1 = Excalidraw default.
	Bowing float64

	// MaxRandomnessOffset caps the per-point jitter magnitude
	// (in user-space units). rough.js default = 2.
	MaxRandomnessOffset float64

	// StrokeWidth in user-space units.
	StrokeWidth float64

	// Stroke is the outline color. Nil disables stroking.
	Stroke color.Color

	// Fill is the interior fill color. Nil disables filling.
	Fill color.Color

	// FillStyle picks the fill pattern.
	FillStyle FillStyle

	// FillWeight is the line width of hachure/zigzag fills.
	// If <= 0, resolves to StrokeWidth/2.
	FillWeight float64

	// HachureAngle is the orientation of hachure lines, in degrees.
	// Excalidraw default = -41.
	HachureAngle float64

	// HachureGap is the spacing between hachure lines, in user-space units.
	// If <= 0, resolves to StrokeWidth*4.
	HachureGap float64

	// CurveTightness biases the Catmull-Rom→Bezier conversion used for
	// rough ellipses. 0 = standard; > 0 pulls control points inward.
	CurveTightness float64

	// CurveFitting (default 0.95) controls how closely the rough ellipse
	// tracks the ideal ellipse. The complement (1 - CurveFitting) becomes
	// the radius-perturbation magnitude.
	CurveFitting float64

	// CurveStepCount is the minimum number of bezier segments per full
	// ellipse. rough.js default = 9.
	CurveStepCount float64

	// PreserveVertices, when true, keeps polygon corners exactly at their
	// requested positions (only control points are perturbed). When false,
	// endpoints jitter freely. Excalidraw enables this for closed shapes.
	PreserveVertices bool

	// DisableMultiStroke, when true, draws each rough line as a single
	// pass instead of the signature double-stroke. Use for thin overlays.
	DisableMultiStroke bool

	// DisableMultiStrokeFill, when true, draws each fill line as a
	// single pass.
	DisableMultiStrokeFill bool
}

// DefaultOptions returns Options matching rough.js / Excalidraw defaults.
// The Seed field is set to 1 (callers should override for variety).
func DefaultOptions() Options {
	return Options{
		Seed:                1,
		Roughness:           1,
		Bowing:              1,
		MaxRandomnessOffset: 2,
		StrokeWidth:         2,
		HachureAngle:        -41,
		HachureGap:          -1, // -> StrokeWidth*4 at use-time
		CurveTightness:      0,
		CurveFitting:        0.95,
		CurveStepCount:      9,
		FillStyle:           FillNone,
		FillWeight:          -1, // -> StrokeWidth/2 at use-time
	}
}

// resolvedHachureGap returns the hachure spacing with sentinel resolution.
func (o Options) resolvedHachureGap() float64 {
	if o.HachureGap <= 0 {
		return o.StrokeWidth * 4
	}
	return o.HachureGap
}
