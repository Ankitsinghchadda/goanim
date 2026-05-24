//go:build darwin || linux

package main

import (
	"syscall"
)

// peakRSSBytes returns the peak resident-set-size of this process in
// bytes. Implemented via getrusage(RUSAGE_SELF). On Darwin the kernel
// reports ru_maxrss in BYTES; on Linux it's in KILOBYTES — we
// normalize.
//
// Calling this multiple times during a run is fine; the value is
// monotonic (the kernel tracks the high-water mark).
func peakRSSBytes() uint64 {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0
	}
	v := uint64(ru.Maxrss)
	// Linux reports in KB. Heuristic: any number under (say) 1<<22 KB
	// (~4GB) on Linux scales by 1024; Darwin reports in bytes natively.
	// We can't reliably detect at runtime, so prefer build tags... but
	// inlining the conversion via GOOS is simpler:
	if isLinux {
		return v * 1024
	}
	return v
}
