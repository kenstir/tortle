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
		verbosity := viper.GetInt("verbose")

		// debug
		if verbosity > 0 {
			fmt.Printf("server: %s\n", viper.GetString("server"))
			fmt.Printf("port: %d\n", viper.GetUint("port"))
			fmt.Printf("username: %s\n", viper.GetString("username"))
			fmt.Printf("password: %s\n", viper.GetString("password"))
		}

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
		if verbosity > 0 {
			fmt.Printf("Connected to deluge\n")
		}

		// methods, err := client.MethodsList(context.Background())
		// if err != nil {
		// 	fmt.Printf("Error getting methods: %s\n", err)
		// 	os.Exit(1)
		// }
		// for _, method := range methods {
		// 	fmt.Printf("%s\n", method)
		// }
		// fmt.Printf("Found %d methods\n", len(methods))

		torrentsStatus, err := client.TorrentsStatus(context.Background(), deluge.StateUnspecified, nil)
		if err != nil {
			fmt.Printf("Error getting torrents status: %s\n", err)
			os.Exit(1)
		}
		if verbosity > 0 {
			fmt.Printf("Found %d torrents\n", len(torrentsStatus))
		}
		fmt.Printf("name,ratio\n")
		for _, ts := range torrentsStatus {
			fmt.Printf("%s,%.1f\n", ts.Name, ts.Ratio)
		}
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
