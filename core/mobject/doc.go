// Package mobject defines the "mathematical object" abstraction goanim
// uses to compose drawings.
//
// A Mobject knows how to render itself, what its bounding box is, what
// children it contains, and what seed it uses for deterministic
// roughness. Domain-specific mobjects (system-design boxes, math
// glyphs, etc.) embed *Group and add semantic methods like MoveTo,
// SetLabel, or Connect.
//
// The seed convention is half of the temporal-stability story for
// future animations: a mobject's seed is set at construction time and
// never changes during translation/rotation/scale, so its rough
// wobble stays stable across frames even as it moves.
package mobject
