/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(purgeCmd)

	purgeCmd.Flags().BoolP("dry-run", "n", false, "Run without removing anything")
	purgeCmd.Flags().StringSliceP("scan-path", "p", []string{}, "Scan path to look for torrent files")
	viper.BindPFlag("purge.dry-run", purgeCmd.Flags().Lookup("dry-run"))
	viper.BindPFlag("purge.scan-path", purgeCmd.Flags().Lookup("scan-path"))
}

var purgeCmd = &cobra.Command{
	Use:   "purge torrent_path",
	Short: "purge hard-linked copies of all torrent files in the given scan paths",
	Args:  cobra.MinimumNArgs(1),
	Run:   purgeCmdRun,
}

func purgeCmdRun(cmd *cobra.Command, args []string) {
	// get args
	torrentPath := args[0]

	// get the flags and go
	dryRun := viper.GetBool("purge.dry-run")
	scanPaths := viper.GetStringSlice("purge.scan-path")
	err := purgeCopies(torrentPath, scanPaths, dryRun)
	if err != nil {
		fatalError(err)
	}
}
