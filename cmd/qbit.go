/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(qbitCmd)

	qbitCmd.PersistentFlags().StringP("server", "s", "http://localhost:8080", "server url")
	qbitCmd.PersistentFlags().StringP("username", "U", "admin", "server username")
	qbitCmd.PersistentFlags().StringP("password", "P", "password", "server password")
	viper.BindPFlag("qbit.server", qbitCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("qbit.username", qbitCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("qbit.password", qbitCmd.PersistentFlags().Lookup("password"))
}

var qbitCmd = &cobra.Command{
	Use:     "qbit",
	Aliases: []string{"q", "qb"},
	Short:   "Manage a qBittorrent server",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
