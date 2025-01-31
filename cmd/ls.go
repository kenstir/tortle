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

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolP("noheader", "n", false, "Don't print the header line")
	viper.BindPFlag("noheader", lsCmd.Flags().Lookup("noheader"))
}

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
		if !viper.GetBool("noheader") {
			fmt.Printf("ratio,name\n")
		}
		for _, ts := range torrentsStatus {
			fmt.Printf("%.1f,%s\n", ts.Ratio, ts.Name)
		}
	},
}
