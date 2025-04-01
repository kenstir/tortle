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

func qbitCreateClient() *internal.QbitClient {
	if viper.GetInt("verbose") > 0 {
		stdoutLogger.Printf("Connecting to %s as user %s\n", viper.GetString("qbit.server"), viper.GetString("qbit.username"))
	}
	return internal.NewQbitClient(qbittorrent.Config{
		Host:     viper.GetString("qbit.server"),
		Username: viper.GetString("qbit.username"),
		Password: viper.GetString("qbit.password"),
	})
}

func qbitGetHostPort() (string, string, error) {
	server := viper.GetString("qbit.server")
	u, err := url.Parse(server)
	if err != nil {
		return "", "", err
	}

	return u.Hostname(), u.Port(), nil
}
