/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/autobrr/go-deluge"
	"github.com/moistari/rls"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	delugeCmd.AddCommand(delugeListCmd)

	delugeListCmd.Flags().StringSliceP("columns", "c", []string{"ratio", "name"}, "Columns to display")
	delugeListCmd.Flags().StringP("filter", "f", "", "Filter torrents by name")
	delugeListCmd.Flags().BoolP("noheader", "n", false, "Don't print the header line")
	viper.BindPFlag("deluge.columns", delugeListCmd.Flags().Lookup("columns"))
	viper.BindPFlag("deluge.filter", delugeListCmd.Flags().Lookup("filter"))
	viper.BindPFlag("deluge.noheader", delugeListCmd.Flags().Lookup("noheader"))
}

var delugeValidColumns = []string{
	"added",
	"audio",
	"channels",
	"completed",
	"download_location",
	"group",
	"hash",
	"name",
	"next_announce",
	"ratio",
	"reannounce",
	"save_path",
	"seed_time",
	"state",
}

var delugeListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List torrents",
	Run: func(cmd *cobra.Command, args []string) {
		// get and check the flags
		verbosity := viper.GetInt("verbose")
		filter := viper.GetString("deluge.filter")
		noheader := viper.GetBool("deluge.noheader")
		columns := viper.GetStringSlice("deluge.columns")
		for _, column := range columns {
			if !slices.Contains(delugeValidColumns, column) {
				fmt.Printf("Unknown column: %s\n", column)
				fmt.Printf("Valid values for --column: %s\n", strings.Join(delugeValidColumns, ", "))
				os.Exit(1)
			}
		}

		// create a deluge client
		if verbosity > 0 {
			fmt.Printf("server: %s\n", viper.GetString("deluge.server"))
			fmt.Printf("port: %d\n", viper.GetUint("deluge.port"))
			fmt.Printf("username: %s\n", viper.GetString("deluge.username"))
			fmt.Printf("password: %s\n", viper.GetString("deluge.password"))
		}
		client := deluge.NewV2(deluge.Settings{
			Hostname:             viper.GetString("deluge.server"),
			Port:                 viper.GetUint("deluge.port"),
			Login:                viper.GetString("deluge.username"),
			Password:             viper.GetString("deluge.password"),
			DebugServerResponses: true,
		})

		// connect
		err := client.Connect(context.Background())
		if err != nil {
			fmt.Printf("Error connecting to deluge: %s\n", err)
			os.Exit(1)
		}
		if verbosity > 0 {
			fmt.Printf("Connected to deluge\n")
		}

		// get torrents
		torrentsStatus, err := client.TorrentsStatus(context.Background(), deluge.StateUnspecified, nil)
		if err != nil {
			fmt.Printf("Error getting torrents status: %s\n", err)
			os.Exit(1)
		}
		if verbosity > 0 {
			fmt.Printf("Found %d torrents\n", len(torrentsStatus))
		}

		// sort torrentsStatus by name
		keys := make([]string, 0, len(torrentsStatus))
		for key, _ := range torrentsStatus {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			return torrentsStatus[keys[i]].Name < torrentsStatus[keys[j]].Name
		})

		// print as CSV
		if !noheader {
			header := strings.Join(columns, ",")
			fmt.Printf("%s\n", header)
		}
		for _, key := range keys {
			ts := torrentsStatus[key]

			// skip if the name doesn't match the filter
			if filter != "" && !strings.Contains(strings.ToLower(ts.Name), strings.ToLower(filter)) {
				continue
			}

			// format columns and print as CSV
			var line []string
			r := rls.ParseString(ts.Name)
			for _, column := range columns {
				line = append(line, delugeFormatColumn(column, ts, r))
			}
			fmt.Printf("%s\n", strings.Join(line, ","))
		}
	},
}

// format the given column
func delugeFormatColumn(column string, ts *deluge.TorrentStatus, r rls.Release) string {
	switch column {
	case "added":
		return formatTimestamp(int64(ts.TimeAdded))
	case "audio":
		return strings.Join(r.Audio, " ")
	case "channels":
		return r.Channels
	case "completed":
		return formatTimestamp(ts.CompletedTime)
	case "download_location":
		return ts.DownloadLocation
	case "group":
		return r.Group
	case "hash":
		return ts.Hash
	case "name":
		return ts.Name
	case "next_announce", "reannounce":
		return fmt.Sprintf("%d", ts.NextAnnounce)
	case "ratio":
		return fmt.Sprintf("%.1f", ts.Ratio)
	case "save_path":
		return ts.SavePath // same as ts.DownloadLocation but easier to type
	case "seed_time":
		return (time.Duration(ts.SeedingTime) * time.Second).String()
	case "state":
		return ts.State
	default:
		return fmt.Sprintf("Unknown column: %s", column)
	}
}
