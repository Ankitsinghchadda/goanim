// Package latex compiles LaTeX math formulas to goanim Path
// primitives via tdewolff/canvas's pure-Go LaTeX parser, with
// on-disk caching for fast iteration.
//
// Unlike a typical LaTeX pipeline (latex → DVI → SVG → paths), this
// implementation does NOT shell out to a system LaTeX installation.
// canvas.ParseLaTeX is a pure-Go implementation of a subset of TeX
// math layout (via star-tex.org/x/tex), so the library has zero
// external dependencies and works on any platform that runs Go.
//
// Limitations (Phase-4 known):
//   - Subset of LaTeX math is supported (no \begin{align} blocks, no
//     custom packages). Stick to inline math expressions.
//   - The returned Path is a single composite — individual glyphs are
//     not separately addressable. We approximate per-glyph access in
//     mathx.Equation by splitting at MoveTo boundaries.
//   - Computer Modern is the font.
//
// Caching: results are cached by SHA-256 of the source string in
// GOANIM_CACHE_DIR (default ~/.cache/goanim/latex/). Cache hits skip
// the parse entirely and load the serialized path.
package latex

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/tdewolff/canvas"
)

// Compile parses a LaTeX math formula and returns it as a
// geometry.Path (consumable by the standard goanim render pipeline).
//
// The returned path is centered at the origin and scaled so its
// height matches `heightPx`. Width is derived from the formula's
// natural aspect ratio.
//
// On first compile, the result is cached on disk. Subsequent calls
// with the same `source` return from cache without re-parsing.
func Compile(source string, heightPx float64) (*geometry.Path, error) {
	if heightPx <= 0 {
		heightPx = 60
	}
	key := cacheKey(source)

	if cached, ok := readCache(key); ok {
		return scaleAndCenter(cached, heightPx), nil
	}

	cp, err := canvas.ParseLaTeX(source)
	if err != nil {
		return nil, fmt.Errorf("latex: parse %q: %w", source, err)
	}

	gp := canvasPathToGeometry(cp)
	_ = writeCache(key, gp) // best-effort; cache errors don't fail compile
	return scaleAndCenter(gp, heightPx), nil
}

// ClusterGlyphs groups the subpaths of p (split at MoveTo) into
// visual glyph clusters using bounding-box geometry. Each returned
// path approximates one visual glyph: closed counters (the inside of
// "O", "a", "e", "p"), accents (the dot on "i"/"j", umlauts) are
// merged with their bodies; horizontally disjoint letters and stacked
// operator parts of comparable size (e.g. the two bars of "=") stay
// separate.
//
// Heuristics applied per pair (existing-cluster bbox, new subpath bbox):
//
//  1. Containment — one bbox lies wholly inside the other (closed
//     counters always nest this way).
//  2. Accent above/below — strong horizontal overlap, the smaller piece
//     is no wider than the larger (rules out wide fraction bars over a
//     narrower numerator), the smaller piece is much shorter than the
//     larger (rules out same-size stacks like "="), and the vertical
//     gap is small relative to the taller piece.
//
// This is a geometric approximation, not real glyph awareness; it will
// over- or under-merge in edge cases (notably equals/divide stay split,
// and large fraction layouts may split per row). Original draw order
// is preserved — clusters appear in the order their first subpath was
// added.
func ClusterGlyphs(p *geometry.Path) []*geometry.Path {
	raw := SubPaths(p)
	if len(raw) <= 1 {
		return raw
	}
	type cluster struct {
		parts []*geometry.Path
		bbox  geometry.Rect
	}
	var clusters []*cluster
	for _, sp := range raw {
		b := sp.Bounds()
		if b.Empty() {
			continue
		}
		merged := false
		for _, c := range clusters {
			if shouldMergeGlyph(b, c.bbox) {
				c.parts = append(c.parts, sp)
				c.bbox = c.bbox.Union(b)
				merged = true
				break
			}
		}
		if !merged {
			clusters = append(clusters, &cluster{
				parts: []*geometry.Path{sp},
				bbox:  b,
			})
		}
	}
	out := make([]*geometry.Path, 0, len(clusters))
	for _, c := range clusters {
		gp := geometry.NewPath()
		for _, part := range c.parts {
			gp.Append(part)
		}
		out = append(out, gp)
	}
	return out
}

// shouldMergeGlyph reports whether a subpath with bbox b should be
// merged into an existing cluster with bbox a.
func shouldMergeGlyph(a, b geometry.Rect) bool {
	if a.Empty() || b.Empty() {
		return false
	}
	// Rule 1: full containment — closed counters (inside of "O", "a"…).
	if rectContains(a, b) || rectContains(b, a) {
		return true
	}
	aw, ah := a.Width(), a.Height()
	bw, bh := b.Width(), b.Height()
	if aw <= 0 || bw <= 0 || ah <= 0 || bh <= 0 {
		return false
	}
	xOverlap := math.Max(0, math.Min(a.Max.X, b.Max.X)-math.Max(a.Min.X, b.Min.X))
	minW := math.Min(aw, bw)
	if xOverlap/minW < 0.7 {
		return false
	}
	// Identify the shorter piece (the candidate accent).
	var smallW, smallH, largeW, largeH float64
	if ah <= bh {
		smallW, smallH, largeW, largeH = aw, ah, bw, bh
	} else {
		smallW, smallH, largeW, largeH = bw, bh, aw, ah
	}
	// An accent isn't wider than its body — this is what separates an
	// "i" dot (narrow on narrow stem) from a fraction bar (wide over a
	// narrower numerator).
	if smallW > 1.1*largeW {
		return false
	}
	// Same-size stacks (e.g. the two bars of "=") shouldn't fuse —
	// require the accent to be visibly smaller than the body.
	if smallH > 0.4*largeH {
		return false
	}
	// Vertical gap must be small relative to the body height.
	yGap := math.Max(0, math.Max(a.Min.Y, b.Min.Y)-math.Min(a.Max.Y, b.Max.Y))
	return yGap < 0.5*largeH
}

func rectContains(outer, inner geometry.Rect) bool {
	return outer.Min.X <= inner.Min.X && outer.Min.Y <= inner.Min.Y &&
		outer.Max.X >= inner.Max.X && outer.Max.Y >= inner.Max.Y
}

// SubPaths splits a composite path at MoveTo boundaries — each
// subpath roughly corresponds to one stroke/contour, not necessarily
// one visual glyph (an "O" yields two subpaths: outer outline and
// counter). For per-glyph grouping use ClusterGlyphs.
func SubPaths(p *geometry.Path) []*geometry.Path {
	var out []*geometry.Path
	var cur *geometry.Path
	for _, c := range p.Cmds {
		if c.Kind == geometry.CmdMove {
			if cur != nil && len(cur.Cmds) > 0 {
				out = append(out, cur)
			}
			cur = geometry.NewPath()
		}
		if cur == nil {
			cur = geometry.NewPath()
		}
		cur.Cmds = append(cur.Cmds, c)
	}
	if cur != nil && len(cur.Cmds) > 0 {
		out = append(out, cur)
	}
	return out
}

// scaleAndCenter normalizes the path to the requested height and
// centers it at the origin.
func scaleAndCenter(p *geometry.Path, heightPx float64) *geometry.Path {
	b := p.Bounds()
	srcH := b.Height()
	srcW := b.Width()
	if srcH <= 0 || srcW <= 0 {
		return p
	}
	scale := heightPx / srcH
	cx := (b.Min.X + b.Max.X) / 2
	cy := (b.Min.Y + b.Max.Y) / 2
	t := geometry.Translate(-cx*scale, -cy*scale).Compose(geometry.Scale(scale))
	return p.Transform(t)
}

// canvasPathToGeometry converts a canvas.Path (the result of
// canvas.ParseLaTeX) into a goanim geometry.Path.
//
// canvas's data format encodes each segment as
// [cmd, ...args, cmd]
// where the leading and trailing cmd are identical (the trailing copy
// makes reverse iteration possible). Segment lengths:
//
//	MoveTo (1.0)  → 4 floats: cmd, x, y, cmd
//	LineTo (2.0)  → 4 floats: cmd, x, y, cmd
//	QuadTo (4.0)  → 6 floats: cmd, cx, cy, x, y, cmd
//	CubeTo (8.0)  → 8 floats: cmd, c1x, c1y, c2x, c2y, x, y, cmd
//	ArcTo  (16.0) → 8 floats: cmd, rx, ry, phi, sweep, x, y, cmd
//	Close  (32.0) → 4 floats: cmd, x, y, cmd  (x,y = subpath start)
const (
	cvMove  = 1.0
	cvLine  = 2.0
	cvQuad  = 4.0
	cvCube  = 8.0
	cvArc   = 16.0
	cvClose = 32.0
)

func canvasPathToGeometry(cp *canvas.Path) *geometry.Path {
	out := geometry.NewPath()
	data := cp.Data()
	i := 0
	// Track the current pen position so we can promote quadratic
	// Bézier segments to cubic correctly (Q→C needs the start point).
	var curX, curY float64
	for i < len(data) {
		cmd := data[i]
		switch cmd {
		case cvMove:
			x, y := data[i+1], data[i+2]
			out.MoveTo(x, y)
			curX, curY = x, y
			i += 4
		case cvLine:
			x, y := data[i+1], data[i+2]
			out.LineTo(x, y)
			curX, curY = x, y
			i += 4
		case cvQuad:
			cx, cy := data[i+1], data[i+2]
			ex, ey := data[i+3], data[i+4]
			// Q→C: c1 = start + 2/3*(ctrl-start), c2 = end + 2/3*(ctrl-end).
			c1x := curX + 2.0/3.0*(cx-curX)
			c1y := curY + 2.0/3.0*(cy-curY)
			c2x := ex + 2.0/3.0*(cx-ex)
			c2y := ey + 2.0/3.0*(cy-ey)
			out.CurveTo(c1x, c1y, c2x, c2y, ex, ey)
			curX, curY = ex, ey
			i += 6
		case cvCube:
			c1x, c1y := data[i+1], data[i+2]
			c2x, c2y := data[i+3], data[i+4]
			ex, ey := data[i+5], data[i+6]
			out.CurveTo(c1x, c1y, c2x, c2y, ex, ey)
			curX, curY = ex, ey
			i += 8
		case cvArc:
			// Arcs are rare in math layout; line-to the endpoint as a
			// rough approximation.
			ex, ey := data[i+5], data[i+6]
			out.LineTo(ex, ey)
			curX, curY = ex, ey
			i += 8
		case cvClose:
			out.Close()
			curX, curY = data[i+1], data[i+2]
			i += 4
		default:
			// Unknown command — skip one float to avoid infinite loop.
			i++
		}
	}
	return out
}

// --- caching ----------------------------------------------------------------

var (
	cacheDirOnce sync.Once
	cacheDir     string
	cacheErr     error
)

func getCacheDir() (string, error) {
	cacheDirOnce.Do(func() {
		if dir := os.Getenv("GOANIM_CACHE_DIR"); dir != "" {
			cacheDir = filepath.Join(dir, "latex")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				cacheErr = err
				return
			}
			cacheDir = filepath.Join(home, ".cache", "goanim", "latex")
		}
		cacheErr = os.MkdirAll(cacheDir, 0o755)
	})
	return cacheDir, cacheErr
}

func cacheKey(source string) string {
	h := sha256.Sum256([]byte(source))
	return hex.EncodeToString(h[:])
}

func readCache(key string) (*geometry.Path, bool) {
	dir, err := getCacheDir()
	if err != nil {
		return nil, false
	}
	f, err := os.Open(filepath.Join(dir, key+".gob"))
	if err != nil {
		return nil, false
	}
	defer func() { _ = f.Close() }()
	var p geometry.Path
	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, false
	}
	return &p, true
}

func writeCache(key string, p *geometry.Path) error {
	dir, err := getCacheDir()
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, key+".gob"))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return gob.NewEncoder(f).Encode(p)
}
