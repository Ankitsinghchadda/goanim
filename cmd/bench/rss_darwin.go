//go:build darwin

package main

// isLinux selects KB-vs-bytes interpretation in peakRSSBytes. Darwin
// reports getrusage Maxrss in BYTES already; Linux reports in KB.
const isLinux = false
