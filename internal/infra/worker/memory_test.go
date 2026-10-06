package worker

import (
	"math"
	"os"
	"testing"
)

func TestParseProcStatmRSS(t *testing.T) {
	rss, err := parseProcStatmRSS([]byte("100 25 10 1 0 5 0\n"), 4096)
	if err != nil {
		t.Fatalf("parse statm: %v", err)
	}
	if want := int64(25 * 4096); rss != want {
		t.Fatalf("RSS = %d, want %d", rss, want)
	}
	for _, payload := range [][]byte{nil, []byte("100"), []byte("100 nope")} {
		if _, err := parseProcStatmRSS(payload, 4096); err == nil {
			t.Fatalf("parse statm %q succeeded, want error", payload)
		}
	}
	if _, err := parseProcStatmRSS([]byte("1 2"), math.MaxInt64); err == nil {
		t.Fatal("overflowing statm RSS succeeded, want error")
	}
}

func TestProcessRSSBytesReadsCurrentProcess(t *testing.T) {
	rss, err := processRSSBytes(os.Getpid())
	if err != nil {
		t.Fatalf("read current process RSS: %v", err)
	}
	if rss <= 0 {
		t.Fatalf("current process RSS = %d, want positive", rss)
	}
}
