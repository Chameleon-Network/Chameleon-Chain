//go:build !windows
// +build !windows

package monitor

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func getDiskUsage(path string) string {
	fs := unix.Statfs_t{}
	err := unix.Statfs(path, &fs)
	if err == nil {
		all := fs.Blocks * uint64(fs.Bsize)
		free := fs.Bfree * uint64(fs.Bsize)
		used := all - free
		return fmt.Sprintf("%.2f", float64(used*100)/float64(all))
	}
	return "N/A"
}
