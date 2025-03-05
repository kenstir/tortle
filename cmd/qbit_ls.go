/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent"
	"github.com/moistari/rls"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	qbitCmd.AddCommand(qbitListCmd)

	qbitListCmd.Flags().StringSliceP("columns", "c", []string{"ratio", "name"}, "Columns to display")
	qbitListCmd.Flags().StringP("filter", "f", "", "Filter torrents by name")
	qbitListCmd.Flags().BoolP("noheader", "n", false, "Don't print the header line")
	viper.BindPFlag("qbit.columns", qbitListCmd.Flags().Lookup("columns"))
	viper.BindPFlag("qbit.filter", qbitListCmd.Flags().Lookup("filter"))
	viper.BindPFlag("qbit.noheader", qbitListCmd.Flags().Lookup("noheader"))
}

var qbitValidColumns = []string{
	"added",
	"audio",
	"channels",
	"completed",
	"download_path",
	"group",
	"hash",
	"name",
	"ratio",
	"save_path",
	"seed_time",
	"state",
}

var qbitListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List torrents",
	Run: func(cmd *cobra.Command, args []string) {
		// get and check the flags
		verbosity := viper.GetInt("verbose")
		filter := viper.GetString("qbit.filter")
		noheader := viper.GetBool("qbit.noheader")
		columns := viper.GetStringSlice("qbit.columns")
		for _, column := range columns {
			if !slices.Contains(qbitValidColumns, column) {
				fmt.Printf("Unknown column: %s\n", column)
				fmt.Printf("Valid values for --column: %s\n", strings.Join(qbitValidColumns, ", "))
				os.Exit(1)
			}
		}

		// create a qbit client
		if verbosity > 0 {
			fmt.Printf("server: %s\n", viper.GetString("qbit.server"))
			fmt.Printf("username: %s\n", viper.GetString("qbit.username"))
			fmt.Printf("password: %s\n", viper.GetString("qbit.password"))
		}
		client := qbittorrent.NewClient(qbittorrent.Config{
			Host:     viper.GetString("qbit.server"),
			Username: viper.GetString("qbit.username"),
			Password: viper.GetString("qbit.password"),
		})

		// connect
		err := client.LoginCtx(context.Background())
		if err != nil {
			fmt.Printf("Error connecting to qBittorrent: %s\n", err)
			os.Exit(1)
		}
		if verbosity > 0 {
			fmt.Printf("Connected to qBittorrent\n")
		}

		// get torrents
		torrents, err := client.GetTorrents(qbittorrent.TorrentFilterOptions{
			Sort: "name",
		})
		if err != nil {
			fmt.Printf("Error getting torrents: %s\n", err.Error())
			os.Exit(1)
		}
		if verbosity > 0 {
			fmt.Fprintf(os.Stderr, "Found %d torrents\n", len(torrents))
		}

		// print as CSV
		if !noheader {
			header := strings.Join(columns, ",")
			fmt.Printf("%s\n", header)
		}
		for _, t := range torrents {
			// skip if the name doesn't match the filter
			if filter != "" && !strings.Contains(strings.ToLower(t.Name), strings.ToLower(filter)) {
				continue
			}

			// format columns and print as CSV
			var line []string
			r := rls.ParseString(t.Name)
			for _, column := range columns {
				line = append(line, formatColumn(column, t, r))
			}
			fmt.Printf("%s\n", strings.Join(line, ","))
		}
	},
}

// format the given column
func formatColumn(column string, t qbittorrent.Torrent, r rls.Release) string {
	switch column {
	case "added":
		return formatTimestamp(int64(t.AddedOn))
	case "audio":
		return strings.Join(r.Audio, " ")
	case "channels":
		return r.Channels
	case "completed":
		return formatTimestamp(t.CompletionOn)
	case "download_path":
		return t.DownloadPath
	case "group":
		return r.Group
	case "hash":
		return t.Hash
	case "name":
		return t.Name
	case "ratio":
		return fmt.Sprintf("%.1f", t.Ratio)
	case "save_path":
		return t.SavePath
	case "seed_time":
		return (time.Duration(t.SeedingTime) * time.Second).String()
	case "state":
		return string(t.State)
	default:
		return fmt.Sprintf("Unknown column: %s", column)
	}
}
