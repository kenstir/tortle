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

	"github.com/kenstir/tortle/internal"
)

func init() {
	qbitCmd.AddCommand(qbitListCmd)

	qbitListCmd.Flags().StringSliceP("columns", "c", []string{"ratio", "name"}, "Columns to display")
	qbitListCmd.Flags().StringP("filter", "f", "", "Filter torrents by name")
	qbitListCmd.Flags().BoolP("humanize", "h", true, "Humanize sizes, e.g. \"2.1 GiB\"")
	qbitListCmd.Flags().BoolP("noheader", "n", false, "Don't print the header line")
	viper.BindPFlag("qbit.columns", qbitListCmd.Flags().Lookup("columns"))
	viper.BindPFlag("qbit.filter", qbitListCmd.Flags().Lookup("filter"))
	viper.BindPFlag("qbit.humanize", qbitListCmd.Flags().Lookup("humanize"))
	viper.BindPFlag("qbit.noheader", qbitListCmd.Flags().Lookup("noheader"))
}

var qbitValidColumns = []string{
	"added",
	"audio",
	"channels",
	"completed",
	"download_path",
	"downloaded",
	"group",
	"hash",
	"name",
	// "next_announce", // for this we need to call GetTorrentPropertiesCtx
	"ratio",
	// "reannounce", // for this we need to call GetTorrentPropertiesCtx
	"save_path",
	"seed_time",
	"state",
	"uploaded",
}

var qbitListCmd = &cobra.Command{
	Use:   "ls [hash]...",
	Short: "List torrents",
	Run:   qbitListCmdRun,
}

func qbitListCmdRun(cmd *cobra.Command, args []string) {
	// get args
	var hashes []string
	hashes = append(hashes, args...)

	// check flags
	columns := viper.GetStringSlice("qbit.columns")
	for _, column := range columns {
		if !slices.Contains(qbitValidColumns, column) {
			fmt.Printf("Unknown column: %s\n", column)
			fmt.Printf("Valid values for --column: %s\n", strings.Join(qbitValidColumns, ", "))
			os.Exit(1)
		}
	}

	// create a qbit client
	client := qbitCreateClient()

	// collect options and go
	opts := ListOptions{
		Columns:  columns,
		Filter:   viper.GetString("qbit.filter"),
		NoHeader: viper.GetBool("qbit.noheader"),
		Humanize: viper.GetBool("qbit.humanize"),
	}
	err := qbitList(context.Background(), client, hashes, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func qbitList(ctx context.Context, client internal.QbitClientInterface, hashes []string, opts ListOptions) error {
	// connect
	err := client.LoginCtx(context.Background())
	if err != nil {
		return err
	}

	// get torrents
	torrents, err := client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{
		Sort:   "name",
		Hashes: hashes,
	})
	if err != nil {
		return err
	}

	// check that all specified torrents were found
	if len(hashes) > 0 && len(hashes) != len(torrents) {
		for _, hash := range hashes {
			found := false
			for _, t := range torrents {
				if t.Hash == hash {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("%s: torrent not found", hash)
			}
		}
	}
	if verbosity > 0 {
		stderrLogger.Printf("Found %d torrents\n", len(torrents))
	}

	// print as CSV
	if !opts.NoHeader {
		fmt.Printf("%s\n", strings.Join(opts.Columns, ","))
	}
	for _, t := range torrents {
		// skip if the name doesn't match the filter
		if opts.Filter != "" && !strings.Contains(strings.ToLower(t.Name), strings.ToLower(opts.Filter)) {
			continue
		}

		// format columns and print as CSV
		var line []string
		r := rls.ParseString(t.Name)
		for _, column := range opts.Columns {
			line = append(line, qbitFormatColumn(column, t, r, opts.Humanize))
		}
		fmt.Printf("%s\n", strings.Join(line, ","))
	}

	return nil
}

// format the given column
func qbitFormatColumn(column string, t qbittorrent.Torrent, r rls.Release, humanize bool) string {
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
	case "downloaded":
		if humanize {
			return humanizeBytes(t.Downloaded)
		}
		return fmt.Sprintf("%d", t.Downloaded)
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
	case "uploaded":
		if humanize {
			return humanizeBytes(t.Uploaded)
		}
		return fmt.Sprintf("%d", t.Uploaded)
	default:
		return fmt.Sprintf("Unknown column: %s", column)
	}
}
