/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"github.com/autobrr/go-deluge"
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
	Use:     "deluge",
	Aliases: []string{"d"},
	Short:   "Manage a deluge server",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func delugeCreateV2Client() *deluge.ClientV2 {
	if viper.GetInt("verbose") > 0 {
		stdoutLogger.Printf("Connecting to %s:%d as user %s\n", viper.GetString("deluge.server"), viper.GetInt("deluge.port"), viper.GetString("deluge.username"))
	}
	return deluge.NewV2(deluge.Settings{
		Hostname: viper.GetString("deluge.server"),
		Port:     uint(viper.GetInt("deluge.port")),
		Login:    viper.GetString("deluge.username"),
		Password: viper.GetString("deluge.password"),
	})
}
