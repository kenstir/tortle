/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ReannounceOptions struct {
	Attempts      int
	Interval      int
	ExtraAttempts int
	ExtraInterval int
	MaxAge        int
}

type ReannounceContext struct {
	Client *qbittorrent.Client
	Log    *log.Logger
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
	Run:     reannounceCommandRun,
}

var stdoutLogger = log.New(os.Stdout, "", log.LstdFlags)

func reannounceCommandRun(cmd *cobra.Command, args []string) {
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
	client := qbittorrent.NewClient(qbittorrent.Config{
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
	reannounce(context.Background(), client, hash, options)
}

func reannounce(ctx context.Context, client *qbittorrent.Client, hash string, options ReannounceOptions) {

	// connect
	err := client.LoginCtx(ctx)
	if err != nil {
		stdoutLogger.Fatalf("Error connecting to qBittorrent: %s\n", err)
	}
	if verbosity > 0 {
		stdoutLogger.Printf("Connected to qBittorrent\n")
	}

	// get torrent
	torrents, err := client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{
		Hashes: []string{hash},
	})
	if err != nil {
		stdoutLogger.Fatalf("Error getting torrents: %s\n", err.Error())
	}
	if len(torrents) != 1 {
		stdoutLogger.Fatalf("%s: torrent not found\n", hash)
	}
	torrent := torrents[0]

	// perform startup checks
	age := time.Now().Unix() - torrent.AddedOn
	stdoutLogger.Printf("%s: found torrent age=%d tracker=%s\n", hash, age, torrent.Tracker)
	if age > int64(options.MaxAge) {
		stdoutLogger.Printf("%s: torrent is %ds old, max_age is %ds\n", hash, age, options.MaxAge)
		return
	}
	if torrent.CompletionOn > 0 {
		stdoutLogger.Printf("%s: torrent is finished\n", hash)
		return
	}

	// reannounce
	reannounceUntilSeeded(ctx, client, torrent, options)
	reannounceForGoodMeasure(ctx, client, torrent, options)
}

func reannounceUntilSeeded(ctx context.Context, client *qbittorrent.Client, t qbittorrent.Torrent, options ReannounceOptions) bool {
	for i := 1; i <= options.Attempts; i++ {
		// delay before every attempt
		if verbosity > 0 {
			stdoutLogger.Printf("%s: try %d: Sleep %d\n", t.Hash, i, options.Interval)
		}
		time.Sleep(time.Duration(options.Interval) * time.Second)

		// get trackers
		trackers, err := client.GetTorrentTrackersCtx(ctx, t.Hash)
		if err != nil {
			stdoutLogger.Fatal(err)
		}

		// find current tracker
		var tracker qbittorrent.TorrentTracker
		found := false
		for _, tr := range trackers {
			if tr.Status != qbittorrent.TrackerStatusDisabled {
				stdoutLogger.Printf("%s: status=%s peer=%d seed=%d dl=%d msg=\"%s\" u=\"%s\"\n", t.Hash, trackerStatus(tr.Status), tr.NumPeers, tr.NumSeeds, tr.NumDownloaded, tr.Message, tr.Url)
			}
			if tr.Url == t.Tracker {
				tracker = tr
				found = true
				break
			}
		}
		if !found {
			stdoutLogger.Fatalf("%s: Tracker with URL %s not found\n", t.Hash, t.Tracker)
		}
		stdoutLogger.Printf("%s: status %s msg %s\n", t.Hash, trackerStatus(tracker.Status), tracker.Message)

		// if status not ok then reannounce
		ok, _ := isTrackerStatusOK(trackers)
		if !ok {
			stdoutLogger.Printf("%s: try %d: reannounce\n", t.Hash, i)
			forceReannounce(ctx, client, t)
			continue
		}

		// if we have no seeds then reannounce
		if tracker.NumSeeds > 0 {
			return true
		} else {
			stdoutLogger.Printf("%s: try %d: reannounce\n", t.Hash, i)
			forceReannounce(ctx, client, t)
			continue
		}
	}

	stdoutLogger.Fatalf("%s: Reannounce attempts exhausted\n", t.Hash)
	return false
}

func reannounceForGoodMeasure(ctx context.Context, client *qbittorrent.Client, t qbittorrent.Torrent, options ReannounceOptions) {
	for i := 1; i <= options.ExtraAttempts; i++ {
		// delay before every attempt
		if verbosity > 0 {
			stdoutLogger.Printf("%s: extra %d: Sleep %d\n", t.Hash, i, options.ExtraInterval)
		}
		time.Sleep(time.Duration(options.ExtraInterval) * time.Second)

		// force reannounce
		stdoutLogger.Printf("%s: extra reannounce %d of %d\n", t.Hash, i, options.ExtraAttempts)
		forceReannounce(ctx, client, t)
	}
}

func forceReannounce(ctx context.Context, client *qbittorrent.Client, t qbittorrent.Torrent) {
	if err := client.ReAnnounceTorrentsCtx(ctx, []string{t.Hash}); err != nil {
		stdoutLogger.Printf("%s: Error reannouncing: %s\n", t.Hash, err)
	}
}

// Check if status not working or something else
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)#get-torrent-trackers
//
//	0 Tracker is disabled (used for DHT, PeX, and LSD)
//	1 Tracker has not been contacted yet
//	2 Tracker has been contacted and is working
//	3 Tracker is updating
//	4 Tracker has been contacted, but it is not working (or doesn't send proper replies)
func isTrackerStatusOK(trackers []qbittorrent.TorrentTracker) (bool, int) {
	for i, tracker := range trackers {
		if tracker.Status == qbittorrent.TrackerStatusDisabled {
			continue
		}

		// check for certain messages before the tracker status to catch ok status with unreg msg
		if isUnregistered(tracker.Message) {
			return false, -1
		}

		if tracker.Status == qbittorrent.TrackerStatusOK {
			return true, i
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
		return "disabled"
	case qbittorrent.TrackerStatusNotContacted:
		return "not contacted"
	case qbittorrent.TrackerStatusOK:
		return "ok"
	case qbittorrent.TrackerStatusUpdating:
		return "updating"
	case qbittorrent.TrackerStatusNotWorking:
		return "not working"
	default:
		return "unknown"
	}
}
