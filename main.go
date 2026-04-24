/*
Copyright © 2025 Kenneth H. Cox
*/
package main

import (
	"runtime/debug"

	"github.com/kenstir/tortle/cmd"
)

// If built by goreleaser, these variables are set by goreleaser using -ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = ""
	builtBy = ""
)

func main() {
	readBuildInfo()
	cmd.Execute(version, commit, date)
}

func safeSubstr(str string, length int) string {
	if len(str) >= length {
		return str[:length]
	} else {
		return str
	}
}

func readBuildInfo() {
	if builtBy == "goreleaser" {
		commit = safeSubstr(commit, 8)
		date = safeSubstr(date, 10)
		return
	} else {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			commit = "unknown"
			return
		}

		vcsModified := false
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				commit = safeSubstr(setting.Value, 8)
			case "vcs.time":
				date = safeSubstr(setting.Value, 10)
			case "vcs.modified":
				vcsModified = setting.Value == "true"
			}
		}
		if vcsModified {
			commit = "locally_modified"
		}
	}
}
