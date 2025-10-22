/*
Copyright (c) 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"

	"github.com/kenstir/tortle/internal"
	"github.com/spf13/cobra"
)

func init() {
	qbitCmd.AddCommand(qbitRmCmd)

	qbitRmCmd.Flags().StringP("filter", "f", "", "Find torrents by name")
}

var qbitRmCmd = &cobra.Command{
	Use:     "rm hash [hash]...",
	Aliases: []string{"remove", "del", "delete"},
	Short:   "Remove torrents",
	Long:    "Remove torrents from qBittorrent by their hash.",
	Args:    cobra.MinimumNArgs(1),
	Run:     qbitRmCmdRun,
}

func qbitRmCmdRun(cmd *cobra.Command, args []string) {
	// create a qbit client
	client := qbitCreateClient()

	err := qbitRm(context.Background(), client, args)
	if err != nil {
		fatalError(err)
	}
}

func qbitRm(ctx context.Context, client internal.QbitClientInterface, hashes []string) error {
	// connect
	err := client.LoginCtx(ctx)
	if err != nil {
		return err
	}

	// remove torrents
	err = client.DeleteTorrentsCtx(ctx, hashes, true)
	if err != nil {
		return err
	}

	return nil
}
