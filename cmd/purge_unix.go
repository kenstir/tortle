//go:build !windows

package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"

	"golang.org/x/sys/unix"
)

func isRegularFile(stat *unix.Stat_t) bool {
	return stat.Mode&unix.S_IFMT == unix.S_IFREG
}

func isDir(stat *unix.Stat_t) bool {
	return stat.Mode&unix.S_IFMT == unix.S_IFDIR
}

func purgeCopies(torrentPath string, scanPaths []string, dryRun bool) error {
	// check that torrentPath exists and is a regular file or directory
	var stat unix.Stat_t
	err := unix.Lstat(torrentPath, &stat)
	if err != nil {
		return fmt.Errorf("%s: %v", torrentPath, err)
	}

	// check that there is at least one scan path
	if len(scanPaths) == 0 {
		return fmt.Errorf("no --scan-path specified")
	}

	// remember device and inodes for all regular files that have more than one link
	base := filepath.Base(torrentPath)
	torrentDevice := stat.Dev
	torrentInodes := []uint64{}
	if isRegularFile(&stat) {
		vvLogf("%s: regular file (nlink: %d)\n", torrentPath, stat.Nlink)
		if stat.Nlink > 1 {
			torrentInodes = append(torrentInodes, stat.Ino)
		}
	} else if isDir(&stat) {
		vvLogf("%s: directory\n", torrentPath)
		torrentInodes = findAllFilesWithHardLinks(torrentPath)
	} else {
		return fmt.Errorf("%s: not a regular file or directory", torrentPath)
	}
	vLogf("%s: found %d files with hard links\n", torrentPath, len(torrentInodes))

	// exit early if there are no inodes to look for
	if len(torrentInodes) == 0 {
		return nil
	}

	// scan paths for matching files and remove them
	var lastError error
	for _, scanPath := range scanPaths {
		vLogf("scanning %s\n", scanPath)
		err = unix.Lstat(scanPath, &stat)
		if err != nil {
			return fmt.Errorf("%s: %v", scanPath, err)
		}
		if stat.Dev != torrentDevice {
			return fmt.Errorf("%s: different file system", scanPath)
		}
		dups := findMatchingFiles(scanPath, torrentInodes)
		for _, dup := range dups {
			if dryRun || verbosity > 0 {
				logf("unlink %s\n", dup)
			}
			if !dryRun {
				err = unix.Unlink(dup)
				if err != nil {
					logErrorf("%s: error unlinking: %v\n", dup, err)
					lastError = err
				}
			}
		}
		if dryRun {
			vLogf("%s: found %d linked copies in %s\n", base, len(dups), scanPath)
		} else if len(dups) > 0 {
			logf("%s: removed %d linked copies in %s\n", base, len(dups), scanPath)
		}
	}

	return lastError
}

func findAllFilesWithHardLinks(rootPath string) []uint64 {
	var inodes []uint64

	// walk the directory tree returning inodes for all regular files with more than one link
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		vvLogf("visit %s\n", path)
		base := filepath.Base(path)

		// stop walking directories that can't be accessed
		if err != nil && d.IsDir() {
			vLogf("skipdir %s: %v\n", base, err)
			return filepath.SkipDir
		}

		// ignore files that can't be accessed
		if err != nil {
			vLogf("%v\n", err)
			return nil
		}

		// ignore directories
		if d.IsDir() {
			return nil
		}

		// stat the file
		var stat unix.Stat_t
		err = unix.Lstat(path, &stat)
		if err != nil {
			vLogf("%v\n", err)
			return nil
		}

		// keep a regular file with more than one link
		if isRegularFile(&stat) && stat.Nlink > 1 {
			vvLogf("match %s: ino=%d nlink=%d\n", base, stat.Ino, stat.Nlink)
			inodes = append(inodes, stat.Ino)
		}

		return nil
	})
	if err != nil {
		vLogf("%s: error walking directory: %v\n", rootPath, err)
	}

	return inodes
}

func findMatchingFiles(rootPath string, inodes []uint64) []string {
	var matches []string
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		vvLogf("visit %s\n", path)

		// stop walking directories that can't be accessed
		if err != nil && d.IsDir() {
			vLogf("skipdir %s: %v\n", path, err)
			return filepath.SkipDir
		}

		// ignore files that can't be accessed
		if err != nil {
			return nil
		}

		// ignore directories
		if d.IsDir() {
			return nil
		}

		// stat the file
		var stat unix.Stat_t
		err = unix.Lstat(path, &stat)
		if err != nil {
			vLogf("%v\n", err)
			return nil
		}

		// keep a regular file with a matching inode
		if isRegularFile(&stat) && slices.Contains(inodes, stat.Ino) {
			vvLogf("%s: match ino=%d\n", path, stat.Ino)
			matches = append(matches, path)
		}

		return nil
	})
	if err != nil {
		vLogf("%s: error walking directory: %v\n", rootPath, err)
	}

	return matches
}
