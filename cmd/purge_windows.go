//go:build windows

package cmd

import (
	"fmt"
)

func purge(torrentPath string, scanPaths []string, dryRun bool) error {
	return fmt.Errorf("purge not supported on Windows")
}
