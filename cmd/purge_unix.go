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
		return fmt.Errorf("no --scan-paths specified")
	}

	// remember device and inodes for all regular files that have more than one link
	torrentDevice := stat.Dev
	torrentInodes := []uint64{}
	if isRegularFile(&stat) {
		vLogf("%s: regular file (nlink: %d)\n", torrentPath, stat.Nlink)
		if stat.Nlink > 1 {
			torrentInodes = append(torrentInodes, stat.Ino)
		}
	} else if isDir(&stat) {
		vLogf("%s: directory\n", torrentPath)
		torrentInodes = findAllFilesWithHardLinks(torrentPath)
	} else {
		return fmt.Errorf("%s: not a regular file or directory", torrentPath)
	}

	// exit early if there are no inodes to look for
	if len(torrentInodes) == 0 {
		vLogf("no regular files with more than one link")
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
				stderrLogger.Printf("unlink %s\n", dup)
			}
			if !dryRun {
				err = unix.Unlink(dup)
				if err != nil {
					stderrLogger.Printf("%s: error unlinking: %v\n", dup, err)
					lastError = err
				}
			}
		}
	}

	return lastError
}

func findAllFilesWithHardLinks(rootPath string) []uint64 {
	var inodes []uint64

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		vvLogf("walk %s\n", path)
		base := filepath.Base(path)

		// skip files that can't be accessed
		if err != nil {
			vLogf("skip %s: %v\n", base, err)
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// skip directories
		if d.IsDir() {
			return nil
		}

		// stat the file
		var stat unix.Stat_t
		err = unix.Lstat(path, &stat)
		if err != nil {
			vLogf("skip %s: %v\n", base, err)
			return nil
		}

		// skip if not a regular file
		if !isRegularFile(&stat) {
			vvLogf("skip %s: not a regular file\n", base)
			return nil
		}

		// keep files with more than one link
		if stat.Nlink > 1 {
			vvLogf("match %s: ino=%d nlink=%d\n", base, stat.Ino, stat.Nlink)
			inodes = append(inodes, stat.Ino)
		}

		return nil
	})
	if err != nil {
		vLogf("%s: error walking directory: %v\n", rootPath, err)
	}
	vvLogf("%s: found %d files with hard links\n", rootPath, len(inodes))
	return inodes
}

func findMatchingFiles(rootPath string, inodes []uint64) []string {
	var matches []string
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		vvLogf("walk %s\n", path)
		if err != nil {
			vLogf("%s: skipping: %v\n", path, err)
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// skip directories
		if d.IsDir() {
			return nil
		}

		// stat the file
		var stat unix.Stat_t
		err = unix.Lstat(path, &stat)
		if err != nil {
			vLogf("%s: skipping: %v\n", path, err)
			return nil
		}

		// check if it's a regular file with a matching inode
		if isRegularFile(&stat) && slices.Contains(inodes, stat.Ino) {
			vvLogf("%s: match (inode: %d)\n", path, stat.Ino)
			matches = append(matches, path)
		}

		return nil
	})
	if err != nil {
		vLogf("%s: error walking directory: %v\n", rootPath, err)
	}
	vvLogf("%s: found %d matches\n", rootPath, len(matches))
	return matches
}
