// Package layout provides compositional positioning primitives —
// HBox, VBox, Grid, Stack, Padding, AlignTo — that let users build
// diagrams declaratively without computing pixel coordinates.
//
// Containers are themselves mobjects: they can be added to scenes,
// translated, and rendered. The container's own appearance is
// invisible — it just positions children. Style is transparent: a
// container does not override its children's styles; children
// continue to inherit from scene defaults.
//
// Layout is computed lazily: positions are recomputed whenever
// Bounds() or Render() is called. Setting a property via the fluent
// API (WithSpacing, WithAlign, etc.) does not eagerly re-layout, so
// chained construction stays cheap.
//
// Layouts compose recursively. A VBox containing HBoxes containing
// individual mobjects is the standard idiom:
//
//	// Two rows of nodes, with a "load balancer" feeding into a row
//	// of three servers:
//	diagram := layout.NewVBox(
//	    lb,
//	    layout.NewHBox(s1, s2, s3).WithSpacing(40),
//	).WithSpacing(80)
//	scene.Add(diagram)
//
// Center-origin coordinates: (0, 0) is the canvas center, Y points up.
// A container's "position" is the center of its bounding box.
package layout
