/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print sample config file to stdout",
	Long: `Print sample config file to stdout

like so:
    tt config > tt.toml
`,
	Run: configCmdRun,
}

func configCmdRun(cmd *cobra.Command, args []string) {
	// TODO: make this work for default localhost installations
	fmt.Printf(`[deluge]
server = "192.168.1.222"
port = 58846
username = "admin"
password = "password"

[qbit]
server = "http://192.168.1.222:8080"
username = "admin"
password = "password"
#columns = ["ratio","hash","name","save_path"]
`)
}
