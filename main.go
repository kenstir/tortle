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
	date    = "unknown"
)

func main() {
	info, ok := debug.ReadBuildInfo()
	if ok {
		vcsRevision := ""
		vcsModified := false
		for _, setting := range info.Settings {
			//fmt.Printf("Key: %s, Value: %s\n", setting.Key, setting.Value)
			if setting.Key == "vcs.revision" {
				vcsRevision = setting.Value
			} else if setting.Key == "vcs.modified" {
				vcsModified = setting.Value == "true"
			}
		}
		if vcsRevision != commit {
			commit = vcsRevision
			if vcsModified {
				commit = "locally modified"
			}
		}
	}
	cmd.Execute(version, commit, date)
}
