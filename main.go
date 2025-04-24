/*
Copyright © 2025 Kenneth H. Cox
*/
package main

import (
	"runtime/debug"

	"github.com/kenstir/tortle/cmd"
)

// These variables are set at release time by goreleaser using -ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = ""
	builtBy = ""
)

func main() {
	overrideVersionInfo()
	cmd.Execute(version, commit, date)
}

func overrideVersionInfo() {
	if builtBy == "goreleaser" {
		// everything is set by goreleaser, skip
		return
	}
	info, ok := debug.ReadBuildInfo()
	if ok {
		vcsRevision := "unknown"
		vcsModified := false
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				vcsRevision = setting.Value
			} else if setting.Key == "vcs.modified" {
				vcsModified = setting.Value == "true"
			}
		}
		if vcsModified {
			commit = "locally modified"
		} else {
			commit = vcsRevision
		}
	}
}
