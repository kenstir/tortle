/*
Copyright © 2025 Kenneth H. Cox
*/
package main

import "github.com/kenstir/tortle/cmd"

// These variables are set at release time by goreleaser using -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
