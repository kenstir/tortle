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
}

var delugeCmd = &cobra.Command{
	Use:              "deluge",
	Aliases:          []string{"d"},
	Short:            "Manage a deluge server",
	PersistentPreRun: delugeCmdInit,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func delugeCmdInit(cmd *cobra.Command, args []string) {
	viper.BindPFlag(delugeKey("server"), cmd.InheritedFlags().Lookup("server"))
	viper.BindPFlag(delugeKey("port"), cmd.InheritedFlags().Lookup("port"))
	viper.BindPFlag(delugeKey("username"), cmd.InheritedFlags().Lookup("username"))
	viper.BindPFlag(delugeKey("password"), cmd.InheritedFlags().Lookup("password"))
}

func delugeCreateV2Client() *deluge.ClientV2 {
	if viper.GetInt("verbose") > 0 {
		stdoutLogger.Printf("Connecting to %s:%d as user %s\n", viper.GetString(delugeKey("server")), viper.GetInt(delugeKey("port")), viper.GetString(delugeKey("username")))
	}
	return deluge.NewV2(deluge.Settings{
		Hostname: viper.GetString(delugeKey("server")),
		Port:     uint(viper.GetInt(delugeKey("port"))),
		Login:    viper.GetString(delugeKey("username")),
		Password: viper.GetString(delugeKey("password")),
	})
}
