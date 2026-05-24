package rough

// rng is the Lehmer (Park-Miller MINSTD) PRNG used by rough.js. We
// reimplement it directly so output matches the JS engine bit-for-bit
// when given the same seed — handy for cross-checking against rough.js
// during development and useful in its own right because the algorithm
// is fast, deterministic, and free of platform-dependent state.
//
// Each call to next() advances the state by:
//
//	seed = (48271 * seed) mod 2^32   (32-bit signed multiply with overflow)
//
// and returns (seed & 0x7FFFFFFF) / 2^31, a float64 in [0, 1).
//
// A seed of zero is invalid — callers should pass a non-zero seed.
// (rough.js falls back to Math.random when seed == 0; we don't.)
type rng struct {
	state int32
}

// newRNG constructs a PRNG. If seed is zero it is bumped to 1 to avoid
// degenerate output.
func newRNG(seed int64) *rng {
	s := int32(seed)
	if s == 0 {
		s = 1
	}
	return &rng{state: s}
}

// next returns a pseudo-random float64 in [0, 1) and advances the state.
func (r *rng) next() float64 {
	r.state = r.state * 48271
	masked := uint32(r.state) & 0x7FFFFFFF
	return float64(masked) / float64(1<<31)
}
