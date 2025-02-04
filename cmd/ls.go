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
	"github.com/moistari/rls"
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

var validColumns = []string{
	"added",
	"audio",
	"channels",
	"completed",
	"download_location",
	"group",
	"name",
	"ratio",
	"save_path",
	"seed_time",
	"state",
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List torrents",
	Run: func(cmd *cobra.Command, args []string) {
		verbosity := viper.GetInt("verbose")
		columns := viper.GetStringSlice("columns")
		for _, column := range columns {
			if !checkColumn(column) {
				fmt.Printf("Unknown column: %s\n", column)
				fmt.Printf("Valid values for --column: %s\n", strings.Join(validColumns, ", "))
				os.Exit(1)
			}
		}

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
		if !viper.GetBool("noheader") {
			header := strings.Join(columns, ",")
			fmt.Printf("%s\n", header)
		}
		for _, ts := range torrentsStatus {
			var line []string
			r := rls.ParseString(ts.Name)
			for _, column := range columns {
				line = append(line, formatColumn(column, ts, r))
			}
			fmt.Printf("%s\n", strings.Join(line, ","))
		}
	},
}

// return true if the column is in the slice validColumns
func checkColumn(column string) bool {
	for _, validColumn := range validColumns {
		if column == validColumn {
			return true
		}
	}
	return false
}

// format the given column
func formatColumn(column string, ts *deluge.TorrentStatus, r rls.Release) string {
	switch column {
	case "added":
		return dateString(int64(ts.TimeAdded))
	case "audio":
		return strings.Join(r.Audio, " ")
	case "channels":
		return r.Channels
	case "completed":
		return dateString(ts.CompletedTime)
	case "download_location":
		return ts.DownloadLocation
	case "name":
		return ts.Name
	case "ratio":
		return fmt.Sprintf("%.1f", ts.Ratio)
	case "save_path":
		return ts.SavePath // same as ts.DownloadLocation but easier to type
	case "seed_time":
		return (time.Duration(ts.SeedingTime) * time.Second).String()
	case "state":
		return ts.State
	case "group":
		return r.Group
	default:
		return fmt.Sprintf("Unknown column: %s", column)
	}
}

// convert a unix timestamp to a string
func dateString(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	//return t.Format(time.RFC3339) //2022-04-11T15:33:20-04:00
	return t.Format("2006-01-02 15:04:05")
}
