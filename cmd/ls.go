/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	deluge "github.com/autobrr/go-deluge"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List torrents",
	Run: func(cmd *cobra.Command, args []string) {
		// parent flags
		// viper.BindPFlag("server", cmd.Parent().PersistentFlags().Lookup("server"))
		// viper.BindPFlag("port", cmd.Parent().PersistentFlags().Lookup("port"))
		// //viper.BindPFlag("username", cmd.Parent().PersistentFlags().Lookup("username"))
		// viper.BindPFlag("username", cmd.Flags().Lookup("usernamexxx"))
		// viper.BindPFlag("password", cmd.Parent().PersistentFlags().Lookup("password"))

		// debug
		fmt.Printf("server: %s\n", viper.GetString("server"))
		fmt.Printf("port: %d\n", viper.GetUint("port"))
		fmt.Printf("username: %s\n", viper.GetString("username"))
		fmt.Printf("password: %s\n", viper.GetString("password"))
		// if viper.GetUint("port") > 0 {
		// 	return
		// }

		client := deluge.NewV2(deluge.Settings{
			Hostname: viper.GetString("server"),
			Port:     viper.GetUint("port"),
			Login:    viper.GetString("username"),
			Password: viper.GetString("password"),
			DebugServerResponses: true,
		})

		err := client.Connect(context.Background())
		if err != nil {
			fmt.Printf("Error connecting to deluge: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Connected to deluge\n")
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
