// Package direction provides goanim's "direction layer" — primitives
// that direct the viewer's attention rather than render new content.
//
// Where core/mobject draws diagram elements and core/animation moves
// them, this package handles pacing (Pause), framing (Camera), and
// emphasis (Spotlight, LaserPointer, Callout, Caption). Together
// they're the toolkit a creator uses to teach with a diagram —
// pausing where a narrator would explain, zooming on the component
// being discussed, pointing at the path the data takes.
//
// Direction primitives implement the core/animation.Animation
// interface. They compose with Sequence / Parallel / Stagger and
// with every other animation in the library. The composition
// guarantee — that a Spotlight during a Camera zoom, a LaserPointer
// through a panning region, etc. work without surprises — is the
// architectural contract that makes the direction layer a layer
// rather than a pile of features.
//
// The package ships:
//
//   - Pause(duration[, label])     — emit zero-change frames
//   - Camera.ZoomTo, PanTo, Reset  — frame-the-viewport animations
//   - Camera.Focus, UnFocus        — ZoomTo + dim everything else
//   - LaserPointer(path, duration) — presenter-style attention dot
//   - Pulse, Spotlight             — emphasis without movement
//   - UnderlineOn, CircleAround,
//     Callout, Caption             — sketchy-aware annotations
package direction
