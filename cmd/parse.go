/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/moistari/rls"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(parseCmd)
}

var parseCmd = &cobra.Command{
	Use:     "parse",
	Aliases: []string{"p"},
	Short:   "Parse a release name",
	Run:     parseCmdRun,
}

func parseCmdRun(cmd *cobra.Command, args []string) {
	for _, name := range args {
		r := rls.ParseString(name)
		jsonOutput, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			fmt.Printf("Error converting to JSON: %v\n", err)
			continue
		}
		fmt.Println(string(jsonOutput))
	}
}
