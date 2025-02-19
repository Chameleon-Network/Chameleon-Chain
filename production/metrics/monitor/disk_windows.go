//go:build windows
// +build windows

package monitor

import (
	"fmt"
	"syscall"
	"unsafe"
)

func getDiskUsage(path string) string {
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.MustFindProc("GetDiskFreeSpaceExW")
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes int64
	_, _, _ = getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)
	used := totalNumberOfBytes - totalNumberOfFreeBytes
	return fmt.Sprintf("%.2f", float64(used*100)/float64(totalNumberOfBytes))
}
