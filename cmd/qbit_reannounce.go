/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kenstir/torinfo/internal"
)

type ReannounceOptions struct {
	Attempts      int
	Interval      int
	ExtraAttempts int
	ExtraInterval int
	MaxAge        int
}

func init() {
	qbitCmd.AddCommand(qbitReannounceCmd)

	qbitReannounceCmd.Flags().IntP("attempts", "a", 60, "Number of reannounce attempts")
	qbitReannounceCmd.Flags().IntP("interval", "i", 7, "Interval between reannounce attempts")
	qbitReannounceCmd.Flags().IntP("extra_attempts", "A", 2, "Number of extra reannounce attempts")
	qbitReannounceCmd.Flags().IntP("extra_interval", "I", 30, "Interval between extra reannounce attempts")
	qbitReannounceCmd.Flags().IntP("max_age", "m", 60*60, "Maximum age of torrent in seconds")
	viper.BindPFlag("qbit.reannounce.attempts", qbitReannounceCmd.Flags().Lookup("attempts"))
	viper.BindPFlag("qbit.reannounce.interval", qbitReannounceCmd.Flags().Lookup("interval"))
	viper.BindPFlag("qbit.reannounce.extra_attempts", qbitReannounceCmd.Flags().Lookup("extra_attempts"))
	viper.BindPFlag("qbit.reannounce.extra_interval", qbitReannounceCmd.Flags().Lookup("extra_interval"))
	viper.BindPFlag("qbit.reannounce.max_age", qbitReannounceCmd.Flags().Lookup("max_age"))
}

var qbitReannounceCmd = &cobra.Command{
	Use:     "reannounce [hash]",
	Aliases: []string{"re", "fast_start", "faststart", "start"},
	Short:   "Reannounce torrent",
	Args:    cobra.ExactArgs(1),
	Run:     qbitReannounceCmdRun,
}

var stdoutLogger = log.New(os.Stdout, "", log.LstdFlags)

func qbitReannounceCmdRun(cmd *cobra.Command, args []string) {
	// get the flags
	verbosity := viper.GetInt("verbose")
	hash := args[0]
	attempts := viper.GetInt("qbit.reannounce.attempts")
	interval := viper.GetInt("qbit.reannounce.interval")
	extraAttempts := viper.GetInt("qbit.reannounce.extra_attempts")
	extraInterval := viper.GetInt("qbit.reannounce.extra_interval")
	maxAge := viper.GetInt("qbit.reannounce.max_age")

	// create a qbit client
	if verbosity > 0 {
		stdoutLogger.Printf("Connecting to %s as user %s\n", viper.GetString("qbit.server"), viper.GetString("qbit.username"))
	}
	client := internal.NewQbitClient(qbittorrent.Config{
		Host:     viper.GetString("qbit.server"),
		Username: viper.GetString("qbit.username"),
		Password: viper.GetString("qbit.password"),
	})

	// reannounce
	options := ReannounceOptions{
		Attempts:      attempts,
		Interval:      interval,
		ExtraAttempts: extraAttempts,
		ExtraInterval: extraInterval,
		MaxAge:        maxAge,
	}
	err := qbitReannounce(context.Background(), client, hash, options)
	if err != nil {
		stdoutLogger.Fatal(err)
	}
}

func qbitReannounce(ctx context.Context, client internal.QbitClientInterface, hash string, opts ReannounceOptions) error {

	// connect
	err := client.LoginCtx(ctx)
	if err != nil {
		return err
	}
	if verbosity > 0 {
		stdoutLogger.Printf("Connected to qBittorrent\n")
	}

	// get torrent
	torrents, err := client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{
		Hashes: []string{hash},
	})
	if err != nil {
		return err
	}
	if len(torrents) != 1 {
		return fmt.Errorf("%s: torrent not found", hash)
	}
	torrent := torrents[0]

	// perform startup checks
	age := time.Now().Unix() - torrent.AddedOn
	stdoutLogger.Printf("%s: found torrent age=%d\n", hash, age)
	if age > int64(opts.MaxAge) {
		return fmt.Errorf("%s: torrent is %ds old, max_age is %ds", hash, age, opts.MaxAge)
	}
	// if torrent.CompletionOn > 0 {
	// 	stdoutLogger.Printf("%s: torrent is finished\n", hash)
	// 	return
	// }

	// reannounce
	err = reannounceUntilSeeded(ctx, client, hash, opts)
	if err != nil {
		return err
	}
	err = reannounceForGoodMeasure(ctx, client, hash, opts)
	if err != nil {
		return err
	}

	return nil
}

func reannounceUntilSeeded(ctx context.Context, client internal.QbitClientInterface, hash string, options ReannounceOptions) error {
	for i := 1; i <= options.Attempts; i++ {
		// delay before every attempt
		if verbosity > 0 {
			stdoutLogger.Printf("%s: try %d: Sleep %d\n", hash, i, options.Interval)
		}
		time.Sleep(time.Duration(options.Interval) * time.Second)

		// get trackers
		trackers, err := client.GetTorrentTrackersCtx(ctx, hash)
		if err != nil {
			return err
		}
		if trackers == nil {
			stdoutLogger.Printf("%s: try %d: no trackers\n", hash, i)
			continue
		}

		// if status not ok then reannounce
		ok, seeds := findOKTrackerWithSeeds(trackers, hash)
		if !ok {
			stdoutLogger.Printf("%s: try %d: reannounce\n", hash, i)
			forceReannounce(ctx, client, hash)
			continue
		}

		stdoutLogger.Printf("%s: try %d: found %d seeds\n", hash, i, seeds)
		return nil
	}

	return fmt.Errorf("%s: Reannounce attempts exhausted", hash)
}

func reannounceForGoodMeasure(ctx context.Context, client internal.QbitClientInterface, hash string, options ReannounceOptions) error {
	for i := 1; i <= options.ExtraAttempts; i++ {
		// delay before every attempt
		if verbosity > 0 {
			stdoutLogger.Printf("%s: extra %d: Sleep %d\n", hash, i, options.ExtraInterval)
		}
		time.Sleep(time.Duration(options.ExtraInterval) * time.Second)

		// force reannounce
		stdoutLogger.Printf("%s: extra reannounce %d of %d\n", hash, i, options.ExtraAttempts)
		forceReannounce(ctx, client, hash)
	}

	return nil
}

func forceReannounce(ctx context.Context, client internal.QbitClientInterface, hash string) {
	if err := client.ReAnnounceTorrentsCtx(ctx, []string{hash}); err != nil {
		stdoutLogger.Printf("%s: Error reannouncing: %s\n", hash, err)
	}
}

// Return true if a tracker is OK and has seeds
//
// Adapted from isTrackerStatusOK from https://github.com/autobrr/go-qbittorrent/
// and modified to fit my needs
//
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)#get-torrent-trackers
//
//	0 Tracker is disabled (used for DHT, PeX, and LSD)
//	1 Tracker has not been contacted yet
//	2 Tracker has been contacted and is working
//	3 Tracker is updating
//	4 Tracker has been contacted, but it is not working (or doesn't send proper replies)
func findOKTrackerWithSeeds(trackers []qbittorrent.TorrentTracker, hash string) (bool, int) {
	// until I am confident in the logic below, print the status of every enabled tracker
	for i, tr := range trackers {
		if tr.Status == qbittorrent.TrackerStatusDisabled {
			continue
		}
		hostname := strings.Split(tr.Url, "/")[2]
		stdoutLogger.Printf("%s:        tr[%d] status=%s seed=%d peer=%d msg=\"%s\" u=%s\n", hash, i, trackerStatus(tr.Status), tr.NumSeeds, tr.NumPeers, tr.Message, hostname)
	}

	// find the first tracker with an OK status and seeds
	for _, tr := range trackers {
		if tr.Status == qbittorrent.TrackerStatusOK && tr.NumSeeds > 0 {
			return true, tr.NumSeeds
		}
	}

	return false, -1
}

func isUnregistered(msg string) bool {
	words := []string{"unregistered", "not registered", "not found", "not exist"}

	msg = strings.ToLower(msg)

	for _, v := range words {
		if strings.Contains(msg, v) {
			return true
		}
	}

	return false
}

func trackerStatus(s qbittorrent.TrackerStatus) string {
	switch s {
	case qbittorrent.TrackerStatusDisabled:
		return "Disabled"
	case qbittorrent.TrackerStatusNotContacted:
		return "NotContacted"
	case qbittorrent.TrackerStatusOK:
		return "OK"
	case qbittorrent.TrackerStatusUpdating:
		return "Updating"
	case qbittorrent.TrackerStatusNotWorking:
		return "NotWorking"
	default:
		return "unknown"
	}
}
