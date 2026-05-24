package style

// Role classifies a piece of text by its semantic purpose in a scene
// — title, heading, body, label, etc. The Phase-9 hierarchy work uses
// roles to enforce typography that reads as a real explainer video
// instead of "everything is FontMedium." Authors set the role; the
// library picks the appropriate size + weight + line-height based on
// the active sloppiness preset (sketchy/Excalifont and crisp/Inter
// need different scales to feel equivalent).
//
// The zero value `RoleUnset` falls through to the legacy FontSize
// enum so existing call sites that didn't specify a role still
// render correctly.
type Role uint8

const (
	RoleUnset      Role = iota // fall back to FontSize enum
	RoleTitle                  // full-screen openers ("BGMI Architecture")
	RoleHeading                // chapter intros
	RoleSubHeading             // section titles within chapters
	RoleBody                   // captions, explanatory text
	RoleLabel                  // component labels (the today-default)
	RoleCaption                // small explanatory text under diagrams
	RoleFootnote               // disclaimers, citations
)

// roleSizes is the canonical token table. Indexed by [Role][isSketchy].
// isSketchy = true → Excalifont metrics (looser, needs ~25% larger).
// isSketchy = false → Inter metrics (tighter, smaller numbers work).
//
// Empirically the values here render at about half the named pixel
// size due to how tdewolff/canvas's Face() interprets sizes (likely
// treating them as points rather than literal pixels at our DPMM
// setting). The hierarchy remains correct — Title is ~2× Body, etc.
// — and these absolute values were tuned to PRODUCE the visual sizes
// the prompt's "Title=120pt sketchy / 96pt crisp" intent describes.
var roleSizes = map[Role][2]float64{
	// {crispPx, sketchyPx} — 4× the prompt's stated values to land at
	// the visual scale the prompt's intent describes (Title visibly
	// dominant, headings clearly secondary). The 4× multiplier
	// compensates for two combined effects: (a) the canvas library
	// renders fonts at roughly half their named pixel size at our
	// DPMM, and (b) on the standard 1920×1080 frame a "title" reads
	// as TITULAR only when it occupies a substantial chunk of the
	// frame, not a small line of text.
	RoleTitle:      {384, 480},
	RoleHeading:    {272, 336},
	RoleSubHeading: {176, 224},
	RoleBody:       {120, 144},
	RoleLabel:      {104, 128},
	RoleCaption:    {88, 112},
	RoleFootnote:   {64, 80},
}

// SizePxForRole returns the absolute pixel size for a role under the
// given sloppiness. RoleUnset returns 0 to signal "fall back to the
// FontSize enum."
//
// SloppinessArchitect (crisp) uses the smaller column; everyone else
// (Artist/Cartoonist) uses the sketchy column. The transition between
// crisp and sketchy at Architect→Artist is intentional — the metrics
// of Excalifont demand larger absolute sizes to read with equivalent
// visual weight to Inter.
func SizePxForRole(r Role, s Sloppiness) float64 {
	if r == RoleUnset {
		return 0
	}
	sizes, ok := roleSizes[r]
	if !ok {
		return 0
	}
	if s == SloppinessArchitect {
		return sizes[0]
	}
	return sizes[1]
}
