package worker

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	memoryWatchInterval              = 100 * time.Millisecond
	maxConsecutiveMemoryReadFailures = 3
)

func startMemoryWatchdog(processID int, limitBytes int64, onLimit, onFailure func()) func() {
	stop := make(chan struct{})
	done := make(chan struct{})
	var stopOnce sync.Once
	go func() {
		defer close(done)
		ticker := time.NewTicker(memoryWatchInterval)
		defer ticker.Stop()
		consecutiveFailures := 0
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				rss, err := processRSSBytes(processID)
				if err != nil {
					consecutiveFailures++
					if consecutiveFailures >= maxConsecutiveMemoryReadFailures {
						onFailure()
						return
					}
					continue
				}
				consecutiveFailures = 0
				if rss > limitBytes {
					onLimit()
					return
				}
			}
		}
	}()
	return func() {
		stopOnce.Do(func() { close(stop) })
		<-done
	}
}

func parseProcStatmRSS(payload []byte, pageSize int64) (int64, error) {
	fields := strings.Fields(string(payload))
	if len(fields) < 2 {
		return 0, errors.New("worker: malformed process statm")
	}
	residentPages, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || residentPages < 0 || pageSize <= 0 {
		return 0, errors.New("worker: malformed process statm RSS")
	}
	if residentPages > math.MaxInt64/pageSize {
		return 0, errors.New("worker: process RSS overflows int64")
	}
	return residentPages * pageSize, nil
}
