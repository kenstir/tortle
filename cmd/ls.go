/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	deluge "github.com/autobrr/go-deluge"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.Flags().StringSliceP("columns", "c", []string{"ratio", "name"}, "Columns to display")
	lsCmd.Flags().BoolP("noheader", "n", false, "Don't print the header line")
	viper.BindPFlag("noheader", lsCmd.Flags().Lookup("noheader"))
	viper.BindPFlag("columns", lsCmd.Flags().Lookup("columns"))
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
			Hostname:             viper.GetString("server"),
			Port:                 viper.GetUint("port"),
			Login:                viper.GetString("username"),
			Password:             viper.GetString("password"),
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
		columns := viper.GetStringSlice("columns")
		if !viper.GetBool("noheader") {
			header := strings.Join(columns, ",")
			fmt.Printf("%s\n", header)
		}
		for _, ts := range torrentsStatus {
			// for each column, add the value to a slice
			// then join the slice with commas
			// then print the line
			var line []string
			for _, column := range columns {
				switch column {
				case "added":
					line = append(line, dateString(ts.TimeAdded))
				case "name":
					line = append(line, ts.Name)
				case "ratio":
					line = append(line, fmt.Sprintf("%.1f", ts.Ratio))
				case "state":
					line = append(line, ts.State)
				default:
					fmt.Printf("Unknown column: %s\n", column)
					os.Exit(1)
				}
			}
			fmt.Printf("%s\n", strings.Join(line, ","))
		}
	},
}

// / convert a unix timestamp to a string
func dateString(str float32) string {
	t := time.Unix(int64(str), 0)
	//return t.Format(time.RFC3339)
	return t.Format("2006-01-02 15:04:05")
}
