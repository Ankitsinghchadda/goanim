// Package render defines the renderer abstraction used by mobjects to
// draw themselves, and provides a concrete raster backend built on
// github.com/tdewolff/canvas.
//
// The renderer accepts goanim user-space coordinates (center-origin,
// Y-up) and translates them to image space at the boundary. Strokes
// are anti-aliased, paths are kept at sub-pixel precision (never snapped
// to integers), and the rasterizer supports an optional 2× supersample
// pass to improve edge quality without changing the output resolution.
package render
