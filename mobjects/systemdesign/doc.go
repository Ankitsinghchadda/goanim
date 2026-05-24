// Package systemdesign provides domain-specific mobjects for
// system-architecture diagrams: Client, Server, Database, Arrow.
//
// Each shape is a *mobject.Group with semantic helpers — MoveTo,
// SetLabel — and renders the outline, fill, label, and any decorative
// elements (rack lines on Server, the cylinder shape on Database, etc.)
// in a single coherent style. Arrows are aware of their endpoints'
// geometry: they attach to the bounding rectangle (for Client/Server)
// or the bounding ellipse (for Database) rather than the center, so
// they look like they're "connecting" rather than "pointing at".
package systemdesign
