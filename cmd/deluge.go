/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(delugeCmd)

	delugeCmd.PersistentFlags().StringP("server", "s", "localhost", "server address")
	delugeCmd.PersistentFlags().IntP("port", "p", 9091, "server port")
	delugeCmd.PersistentFlags().StringP("username", "U", "admin", "server username")
	delugeCmd.PersistentFlags().StringP("password", "P", "password", "server password")
	viper.BindPFlag("deluge.server", delugeCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("deluge.port", delugeCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("deluge.username", delugeCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("deluge.password", delugeCmd.PersistentFlags().Lookup("password"))
}

var delugeCmd = &cobra.Command{
	Use:   "deluge",
	Short: "Manage a deluge server",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
