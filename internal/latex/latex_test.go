package latex

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ankitsinghchadda/goanim/core/geometry"
)

// TestCompileProducesPath — basic smoke test: parsing E = mc^2
// produces a non-empty path.
func TestCompileProducesPath(t *testing.T) {
	// Use a fresh temp cache so we know we're hitting the parser.
	tmp := t.TempDir()
	os.Setenv("GOANIM_CACHE_DIR", tmp)
	defer os.Unsetenv("GOANIM_CACHE_DIR")
	// Force re-init of the cache dir.
	resetCacheForTest()

	p, err := Compile("E = mc^2", 80)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if len(p.Cmds) == 0 {
		t.Fatal("expected non-empty path")
	}
}

// TestCachingSpeedsUp — a second compile of the same source should
// be much faster (cache hit) than the first.
func TestCachingSpeedsUp(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("GOANIM_CACHE_DIR", tmp)
	defer os.Unsetenv("GOANIM_CACHE_DIR")
	resetCacheForTest()

	src := `\frac{a+b}{c-d}`

	start := time.Now()
	_, err := Compile(src, 80)
	if err != nil {
		t.Fatalf("first compile: %v", err)
	}
	firstDur := time.Since(start)

	// Verify the cache file was written.
	files, _ := filepath.Glob(filepath.Join(tmp, "latex", "*.gob"))
	if len(files) == 0 {
		t.Fatal("expected cache file to be written")
	}

	// Second call should hit the cache.
	start = time.Now()
	_, err = Compile(src, 80)
	if err != nil {
		t.Fatalf("second compile: %v", err)
	}
	cachedDur := time.Since(start)

	if cachedDur > firstDur/2 {
		t.Logf("warning: cache may not have helped (first %v, cached %v)", firstDur, cachedDur)
	}
}

// TestSubPaths — splitting a composite path returns at least one
// subpath per glyph (one MoveTo each).
func TestSubPaths(t *testing.T) {
	p, err := Compile("xy", 60)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	subs := SubPaths(p)
	if len(subs) < 2 {
		t.Errorf("expected >= 2 subpaths for 'xy', got %d", len(subs))
	}
}

// resetCacheForTest re-initializes the cache singleton so test envs
// take effect. Test-only helper.
func resetCacheForTest() {
	cacheDirOnce = sync.Once{}
	cacheDir = ""
	cacheErr = nil
}

// boxPath appends a closed rectangle subpath to p.
func boxPath(p *geometry.Path, x0, y0, x1, y1 float64) {
	p.MoveTo(x0, y0)
	p.LineTo(x1, y0)
	p.LineTo(x1, y1)
	p.LineTo(x0, y1)
	p.Close()
}

// TestClusterGlyphsContainment — an outer outline with an inner
// counter (like the closed shape of "O" or "a") should merge into one
// cluster.
func TestClusterGlyphsContainment(t *testing.T) {
	p := geometry.NewPath()
	boxPath(p, 0, 0, 10, 10) // outer
	boxPath(p, 3, 3, 7, 7)   // counter, fully contained
	got := ClusterGlyphs(p)
	if len(got) != 1 {
		t.Errorf("containment: expected 1 cluster, got %d", len(got))
	}
}

// TestClusterGlyphsDisjoint — two side-by-side letters with no overlap
// should stay separate.
func TestClusterGlyphsDisjoint(t *testing.T) {
	p := geometry.NewPath()
	boxPath(p, 0, 0, 10, 10)  // letter 1
	boxPath(p, 15, 0, 25, 10) // letter 2
	got := ClusterGlyphs(p)
	if len(got) != 2 {
		t.Errorf("disjoint: expected 2 clusters, got %d", len(got))
	}
}

// TestClusterGlyphsAccent — a narrow short dot tightly above a taller
// stem (the "i" pattern) should merge.
func TestClusterGlyphsAccent(t *testing.T) {
	p := geometry.NewPath()
	boxPath(p, 4, 0, 6, 6)     // stem
	boxPath(p, 4, 8, 6, 9.5)   // dot, small gap above stem
	got := ClusterGlyphs(p)
	if len(got) != 1 {
		t.Errorf("accent: expected 1 cluster, got %d", len(got))
	}
}

// TestClusterGlyphsFractionStaysSplit — a wide thin bar over a
// narrower numerator (\frac pattern) must NOT merge — the bar is
// wider than the body, which is the signal that it isn't an accent.
func TestClusterGlyphsFractionStaysSplit(t *testing.T) {
	p := geometry.NewPath()
	boxPath(p, 5, 15, 20, 45) // numerator (narrow, tall)
	boxPath(p, 0, 12, 25, 13) // fraction bar (wide, thin, just below)
	got := ClusterGlyphs(p)
	if len(got) != 2 {
		t.Errorf("fraction: expected 2 clusters (bar must not absorb numerator), got %d", len(got))
	}
}

// TestClusterGlyphsEqualsStaysSplit — same-size stacked pieces (the
// two bars of "=") shouldn't fuse, since neither is small relative to
// the other. This is a known limitation — acceptable for animation.
func TestClusterGlyphsEqualsStaysSplit(t *testing.T) {
	p := geometry.NewPath()
	boxPath(p, 0, 4, 10, 5) // top bar
	boxPath(p, 0, 1, 10, 2) // bottom bar
	got := ClusterGlyphs(p)
	if len(got) != 2 {
		t.Errorf("equals: expected 2 clusters (same-size stacks stay split), got %d", len(got))
	}
}

// TestClusterGlyphsReducesForLetterWithCounter — running clustering on
// a real LaTeX-rendered glyph with a counter ("O") should reduce the
// path count vs raw SubPaths.
func TestClusterGlyphsReducesForLetterWithCounter(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("GOANIM_CACHE_DIR", tmp)
	defer os.Unsetenv("GOANIM_CACHE_DIR")
	resetCacheForTest()

	p, err := Compile("O", 80)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	raw := SubPaths(p)
	clustered := ClusterGlyphs(p)
	if len(raw) < 2 {
		t.Skipf("canvas produced only %d subpaths for 'O' — nothing to cluster", len(raw))
	}
	if len(clustered) >= len(raw) {
		t.Errorf("expected clustering to reduce 'O' subpaths; raw=%d clustered=%d", len(raw), len(clustered))
	}
}
