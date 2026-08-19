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
	purgeCmd.Flags().StringSliceP("scan-path", "p", []string{}, "Paths to look for hard-linked copies of the files in TORRENT_PATH")
	viper.BindPFlag("purge.dry-run", purgeCmd.Flags().Lookup("dry-run"))
	viper.BindPFlag("purge.scan-path", purgeCmd.Flags().Lookup("scan-path"))
}

var purgeCmd = &cobra.Command{
	Use:   "purge TORRENT_PATH",
	Short: "purge hard-linked copies of files in TORRENT_PATH",
	Long: `Scan every directory in SCAN-PATH and remove any files which are hard-linked to files in TORRENT_PATH.

purge is intended to be used as an external program, removing hard-linked copies of files
created by Sonarr when the original torrent is removed.  Specify the Sonarr root folder
in the tt.toml file, e.g.

  [purge]
  scan-path = ["/mnt/sonarr-root-folder"]

Then run an automation at the time a torrent is removed, e.g.

  tt purge "{save_path}/{name}"
`,
	Args: cobra.MinimumNArgs(1),
	Run:  purgeCmdRun,
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
