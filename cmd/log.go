/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"io"
	"log"
	"os"
)

var stdoutLogger = newLogger(os.Stdout)
var stderrLogger = newLogger(os.Stderr)

func newLogger(w io.Writer) *log.Logger {
	return log.New(w, "", log.Ldate|log.Ltime|log.Lmicroseconds)
}
