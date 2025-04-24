/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/autobrr/go-qbittorrent"
	"github.com/kenstir/tortle/internal"
	"github.com/spf13/cobra"
)

func init() {
	qbitCmd.AddCommand(qbitStatsCmd)
}

var qbitStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get stats in InfluxDB line protocol format",
	Run:   qbitStatsCmdRun,
}

func qbitStatsCmdRun(cmd *cobra.Command, args []string) {
	// create a qbit client
	client := qbitCreateClient()

	// get and print stats
	err := qbitStats(context.Background(), client)
	if err != nil {
		fatalError(err)
	}
}

func qbitStats(ctx context.Context, client internal.QbitClientInterface) error {
	// connect
	err := client.LoginCtx(ctx)
	if err != nil {
		return err
	}
	if verbosity > 0 {
		stdoutLogger.Printf("Connected to qBittorrent\n")
	}

	// get transfer info
	info, err := client.GetTransferInfoCtx(ctx)
	if err != nil {
		return err
	}

	// organize data into tags and fields
	// See also https://docs.influxdata.com/influxdb/v1/write_protocols/line_protocol_tutorial/
	host, port, err := qbitGetHostPort()
	if err != nil {
		return err
	}
	tags := []string{
		"client_type=qbittorrent",
		fmt.Sprintf("client_host=%s", host),
		fmt.Sprintf("client_port=%s", port),
	}
	fields := []string{
		fmt.Sprintf("download_rate=%d", info.DlInfoSpeed),
		fmt.Sprintf("upload_rate=%d", info.UpInfoSpeed),
		fmt.Sprintf("total_download=%du", info.DlInfoData),
		fmt.Sprintf("total_upload=%du", info.UpInfoData),
	}

	// add calculated fields
	torrents, err := client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{})
	if err != nil {
		return err
	}
	fields = append(fields, qbitStatsAddComputedFields(torrents)...)

	printMeasurement("tt_stats", tags, fields)
	return nil
}

func qbitStatsAddComputedFields(torrents []qbittorrent.Torrent) []string {
	numActive := 0
	numSeeding := 0
	numDownloading := 0
	numError := 0
	for _, t := range torrents {
		if t.UpSpeed > 0 || t.DlSpeed > 0 {
			numActive++
		}
		if t.State == qbittorrent.TorrentStateError {
			numError++
		}
		if t.State == qbittorrent.TorrentStateUploading || t.State == qbittorrent.TorrentStateStalledUp || t.State == qbittorrent.TorrentStateForcedUp {
			numSeeding++
		} else if t.State == qbittorrent.TorrentStateDownloading {
			numDownloading++
		}
	}

	fields := []string{
		fmt.Sprintf("num_torrents=%du", len(torrents)),
		fmt.Sprintf("num_active=%du", numActive),
		fmt.Sprintf("num_seeding=%du", numSeeding),
		fmt.Sprintf("num_downloading=%du", numDownloading),
		fmt.Sprintf("num_error=%du", numError),
	}
	return fields
}
