//go:build !windows

package cmd

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func VerbosePrintf(format string, args ...interface{}) {
	if verbosity > 0 {
		stderrLogger.Printf(format, args...)
	}
}

func purge(torrentPath string, scanPaths []string, dryRun bool) error {
	// find the device and inode of the torrent file
	// fileInfo, err := os.Stat(torrentPath)
	// if err != nil {
	// 	return err
	// }

	var stat unix.Stat_t
	err := unix.Lstat(torrentPath, &stat)
	if err != nil {
		return fmt.Errorf("%s: %v", torrentPath, err)
	}

	switch stat.Mode & unix.S_IFMT {
	case unix.S_IFREG:
		VerbosePrintf("%s: Regular file\n", torrentPath)
	case unix.S_IFDIR:
		VerbosePrintf("%s: Directory\n", torrentPath)
	default:
		return fmt.Errorf("%s: not a regular file or directory", torrentPath)
	}
	VerbosePrintf("Torrent file: %s\n", torrentPath)
	VerbosePrintf("Dev:          %d\n", stat.Dev)
	VerbosePrintf("Inode:        %d\n", stat.Ino)
	VerbosePrintf("Nlink:        %d\n", stat.Nlink)

	// TODO: Implement the actual purging logic
	// This would involve:
	// 1. Walking through scanPaths
	// 2. Finding files with same device/inode
	// 3. Removing hard links if !dryRun

	return nil
}
