//go:build windows

package cmd

import (
	"fmt"
)

func purgeCopies(_ string, _ []string, _ bool) error {
	return fmt.Errorf("purge not supported on Windows")
}
