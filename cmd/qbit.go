/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"net/url"

	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kenstir/tortle/internal"
)

func init() {
	rootCmd.AddCommand(qbitCmd)

	qbitCmd.PersistentFlags().StringP("server", "S", "http://localhost:8080", "server url")
	qbitCmd.PersistentFlags().StringP("username", "U", "admin", "server username")
	qbitCmd.PersistentFlags().StringP("password", "P", "password", "server password")
}

var qbitCmd = &cobra.Command{
	Use:              "qbit",
	Aliases:          []string{"q", "qb", "qbittorrent"},
	Short:            "Manage a qBittorrent server",
	PersistentPreRun: qbitCmdInit,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func qbitCmdInit(cmd *cobra.Command, args []string) {
	viper.BindPFlag(qbitKey("server"), cmd.InheritedFlags().Lookup("server"))
	viper.BindPFlag(qbitKey("username"), cmd.InheritedFlags().Lookup("username"))
	viper.BindPFlag(qbitKey("password"), cmd.InheritedFlags().Lookup("password"))
}

func qbitCreateClient() *internal.QbitClient {
	if viper.GetInt("verbose") > 0 {
		stdoutLogger.Printf("Connecting to %s as user %s\n", viper.GetString(qbitKey("server")), viper.GetString(qbitKey("username")))
	}
	return internal.NewQbitClient(qbittorrent.Config{
		Host:     viper.GetString(qbitKey("server")),
		Username: viper.GetString(qbitKey("username")),
		Password: viper.GetString(qbitKey("password")),
	})
}

func qbitGetHostPort() (string, string, error) {
	server := viper.GetString(qbitKey("server"))
	u, err := url.Parse(server)
	if err != nil {
		return "", "", err
	}

	return u.Hostname(), u.Port(), nil
}
