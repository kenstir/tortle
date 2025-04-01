/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"
	"fmt"
	"os"

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
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
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
		// PayloadDownloadRate and PayloadUploadRate are the ones shown in the GUI
		// fmt.Sprintf("download_rate=%.1f", status.DownloadRate),
		// fmt.Sprintf("upload_rate=%.1f", status.UploadRate),
		fmt.Sprintf("download_rate=%.1f", status.PayloadDownloadRate),
		fmt.Sprintf("upload_rate=%.1f", status.PayloadUploadRate),
		fmt.Sprintf("total_download=%du", status.TotalDownload),
		fmt.Sprintf("total_upload=%du", status.TotalUpload),
	}

	printMeasurement("tt_stats", tags, fields)
	return nil
}
