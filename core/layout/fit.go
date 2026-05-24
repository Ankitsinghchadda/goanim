// Phase-9 auto-fit composition. FitMode is a hint on a layout
// container that says "fit yourself to the canvas" instead of "use
// child natural sizes." The container measures the natural total
// extent, computes a uniform scale factor, and applies it via
// SetVisualScale to each child (and uses the scaled extent for child
// positioning).
//
// The motivating problem: in the BGMI deliverable the CDN scene had
// 6 region clusters laid out at fixed widths, total ~2160px, on a
// 1920px canvas — and they clipped off-screen. FitMode.FitToWidth
// scales them down to fit. FitContent does the opposite — a 2-icon
// scene with vast empty space gets its icons scaled UP (capped) so
// they don't look lost.

package layout

// FitMode picks how a layout container scales its children to fit
// the target canvas region. The zero value (FitFixed) preserves
// Phase-1 behavior — use child natural sizes as-is.
type FitMode uint8

const (
	// FitFixed uses child natural sizes; the layout may overflow.
	// (Pre-Phase-9 behavior.) A composition-validation warning will
	// fire at render time if the layout actually overflows.
	FitFixed FitMode = iota

	// FitToWidth scales children uniformly so the layout's total
	// width matches the safe target width. If children naturally fit,
	// no scaling is applied.
	FitToWidth

	// FitContent scales children UP (capped at safeMax, default 1.6)
	// when the natural width is much smaller than the safe target.
	// Use when you have only 1-3 components and the default sizes
	// look lost in the canvas.
	FitContent

	// FitToCanvas scales to fit both width and height (preserves
	// aspect ratio of the row). Mostly useful for very tall content.
	FitToCanvas
)

// DefaultSafeWidth is the safe-width target when WithSafeWidth isn't
// called. Sized to match the standard 1920px canvas minus 5%
// horizontal safe area on each side. Layouts using a different
// canvas should set this explicitly.
const DefaultSafeWidth = 1920.0 * 0.9

// Default min/max scale caps. These prevent Fit from producing
// extreme sizes — a 30-icon row capped at 0.5× scale is still
// usefully sized; a 2-icon row scales up to 2.5× max so a
// 2-component EstablishingShot fills the canvas instead of floating
// in it (Phase-10 Fix 4). FitToWidth still uses the lower
// DefaultSafeMax via its own clamp because it's about preventing
// overflow, not filling.
const (
	DefaultSafeMin    = 0.5
	DefaultSafeMax    = 1.6
	DefaultFillMaxUp  = 2.5 // FitContent / FitFill upper bound.
	DefaultSafeHeight = 1080.0 * 0.78
)
