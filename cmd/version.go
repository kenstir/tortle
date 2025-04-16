/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	Run:   versionCmdRun,
}

func versionCmdRun(cmd *cobra.Command, args []string) {
	fmt.Printf("tt version %s (%s) built %s\n", versionInfo.Version, versionInfo.Commit, versionInfo.Date)
}
