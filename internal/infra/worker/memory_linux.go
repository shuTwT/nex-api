//go:build linux

package worker

import (
	"fmt"
	"os"
)

func processRSSBytes(processID int) (int64, error) {
	payload, err := os.ReadFile(fmt.Sprintf("/proc/%d/statm", processID))
	if err != nil {
		return 0, fmt.Errorf("worker: read process RSS: %w", err)
	}
	return parseProcStatmRSS(payload, int64(os.Getpagesize()))
}
