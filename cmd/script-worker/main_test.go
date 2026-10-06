package main

import (
	"runtime/debug"
	"testing"
)

func TestSetMemoryLimitLeavesProcessHeadroom(t *testing.T) {
	previous := debug.SetMemoryLimit(-1)
	t.Cleanup(func() { debug.SetMemoryLimit(previous) })

	const maxResidentBytes = int64(256 << 20)
	setMemoryLimit(maxResidentBytes)

	want := maxResidentBytes - maxResidentBytes/10
	if got := debug.SetMemoryLimit(-1); got != want {
		t.Fatalf("Go runtime memory limit = %d, want %d", got, want)
	}
}

func TestRuntimeMemoryLimitRetainsSmallPositiveBudget(t *testing.T) {
	if got := runtimeMemoryLimit(1); got != 1 {
		t.Fatalf("runtime memory limit = %d, want 1", got)
	}
}
