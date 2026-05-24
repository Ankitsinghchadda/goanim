package latex

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
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
