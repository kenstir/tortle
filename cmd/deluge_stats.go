/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/autobrr/go-deluge"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	delugeCmd.AddCommand(delugeStatsCmd)
}

var delugeStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get stats in InfluxDB line protocol format",
	Run:   delugeStatsCmdRun,
}

func delugeStatsCmdRun(cmd *cobra.Command, args []string) {
	// create a deluge client
	client := delugeCreateV2Client()

	// get and print stats
	err := delugeStats(context.Background(), client)
	if err != nil {
		fatalError(err)
	}
}

func delugeStats(ctx context.Context, client deluge.DelugeClient) error {
	// connect
	err := client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// get session status
	status, err := client.GetSessionStatus(ctx)
	if err != nil {
		return err
	}

	// organize data into tags and fields
	// See also https://docs.influxdata.com/influxdb/v1/write_protocols/line_protocol_tutorial/
	tags := []string{
		"client_type=deluge",
		fmt.Sprintf("client_host=%s", viper.GetString("deluge.server")),
		fmt.Sprintf("client_port=%d", viper.GetInt("deluge.port")),
	}
	fields := []string{
		// Seems nobody wants to see DownloadRate and UploadRate;
		// PayloadDownloadRate and PayloadUploadRate are the ones shown in the Deluge GUI
		// fmt.Sprintf("download_rate=%.1f", status.DownloadRate),
		// fmt.Sprintf("upload_rate=%.1f", status.UploadRate),
		fmt.Sprintf("download_rate=%.1f", status.PayloadDownloadRate),
		fmt.Sprintf("upload_rate=%.1f", status.PayloadUploadRate),
		fmt.Sprintf("total_download=%du", status.TotalDownload),
		fmt.Sprintf("total_upload=%du", status.TotalUpload),
	}

	// add calculated fields
	torrentsStatus, err := client.TorrentsStatus(ctx, deluge.StateUnspecified, nil)
	if err != nil {
		return err
	}
	fields = append(fields, delugeStatsCalculatedFields(torrentsStatus)...)

	printMeasurement("tt_stats", tags, fields)
	return nil
}

func delugeStatsCalculatedFields(torrentsStatus map[string]*deluge.TorrentStatus) []string {
	numActive := 0
	numSeeding := 0
	numDownloading := 0
	numError := 0
	for _, ts := range torrentsStatus {
		if ts.UploadPayloadRate > 0 || ts.DownloadPayloadRate > 0 {
			numActive++
		}
		if ts.State == string(deluge.StateSeeding) {
			numSeeding++
		} else if ts.State == string(deluge.StateDownloading) {
			numDownloading++
		} else if ts.State == string(deluge.StateError) {
			numError++
		}
	}
	fields := []string{
		fmt.Sprintf("num_torrents=%du", len(torrentsStatus)),
		fmt.Sprintf("num_active=%du", numActive),
		fmt.Sprintf("num_seeding=%du", numSeeding),
		fmt.Sprintf("num_downloading=%du", numDownloading),
		fmt.Sprintf("num_error=%du", numError),
	}
	return fields
}
