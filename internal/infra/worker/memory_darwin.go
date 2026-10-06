//go:build darwin

package worker

import (
	"fmt"
	"syscall"
	"unsafe"
)

// procTaskInfo mirrors Darwin's public struct proc_taskinfo enough to read
// pti_resident_size without invoking an external ps process. The proc_info
// syscall is the primitive used by libproc's proc_pidinfo wrapper.
type procTaskInfo struct {
	virtualSize      uint64
	residentSize     uint64
	totalUser        uint64
	totalSystem      uint64
	threadsUser      uint64
	threadsSystem    uint64
	policy           int32
	faults           int32
	pageins          int32
	cowFaults        int32
	messagesSent     int32
	messagesReceived int32
	syscallsMach     int32
	syscallsUnix     int32
	csw              int32
	threadnum        int32
	numrunning       int32
	priority         int32
}

func processRSSBytes(processID int) (int64, error) {
	var info procTaskInfo
	result, _, errno := syscall.Syscall6(
		syscall.SYS_PROC_INFO,
		2, // PROC_INFO_CALL_PIDINFO
		uintptr(processID),
		4, // PROC_PIDTASKINFO
		0,
		uintptr(unsafe.Pointer(&info)),
		uintptr(unsafe.Sizeof(info)),
	)
	if errno != 0 {
		return 0, fmt.Errorf("worker: read process RSS: %w", errno)
	}
	if result == 0 || info.residentSize == 0 || info.residentSize > uint64(^uint64(0)>>1) {
		return 0, fmt.Errorf("worker: invalid process RSS result")
	}
	return int64(info.residentSize), nil
}
