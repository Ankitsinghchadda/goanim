package mobject

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// shapeCache memoizes generated rough geometry for a mobject. The
// cached path is always centered at the origin — callers transform it
// to the current position at render time. This is the load-bearing
// piece of temporal stability: translating a mobject across the screen
// reuses the SAME cached rough geometry, so the wobble doesn't change.
//
// The cache key encodes everything that affects the geometry: seed,
// dimensions, and the style attributes that influence rough generation
// (sloppiness/roughness, edges, stroke width, hachure gap).
type shapeCache struct {
	key  cacheKey
	path *geometry.Path
}

type cacheKey struct {
	seed          int64
	w, h          float64
	roughness     float64
	bowing        float64
	maxJitter     float64
	strokeWidth   float64
	preserveVerts bool
}

func (c *shapeCache) invalidate() { c.path = nil }

func (c *shapeCache) buildKey(seed int64, w, h float64, _ style.Style, tok style.Tokens, preserveVerts bool) cacheKey {
	return cacheKey{
		seed: seed, w: w, h: h,
		roughness: tok.Roughness, bowing: tok.Bowing,
		maxJitter: tok.MaxJitter, strokeWidth: tok.StrokeWidthPx,
		preserveVerts: preserveVerts,
	}
}

// roughRect returns a cached rough rectangle path centered at the
// origin. The path is regenerated only when the cache key changes.
func (c *shapeCache) roughRect(seed int64, w, h float64, eff style.Style, tok style.Tokens) *geometry.Path {
	k := c.buildKey(seed, w, h, eff, tok, eff.Edges == style.EdgesSharp)
	if c.path != nil && c.key == k {
		return c.path
	}
	opts := style.RoughOptions(eff, tok, seed)
	c.path = rough.RoughRectangleCentered(0, 0, w, h, opts)
	c.key = k
	return c.path
}

// roughEllipse returns a cached rough ellipse path centered at origin.
func (c *shapeCache) roughEllipse(seed int64, rx, ry float64, eff style.Style, tok style.Tokens) *geometry.Path {
	k := c.buildKey(seed, rx*2, ry*2, eff, tok, eff.Edges == style.EdgesSharp)
	if c.path != nil && c.key == k {
		return c.path
	}
	opts := style.RoughOptions(eff, tok, seed)
	c.path = rough.RoughEllipse(0, 0, rx, ry, opts)
	c.key = k
	return c.path
}
