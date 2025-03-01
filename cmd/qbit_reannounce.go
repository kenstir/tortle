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

func init() {
	qbitCmd.AddCommand(qbitReannounceCmd)

}

var qbitReannounceCmd = &cobra.Command{
	Use:     "reannounce [hash]",
	Aliases: []string{"re", "fast_start", "faststart", "start"},
	Short:   "Reannounce torrent",
	Args:    cobra.ExactArgs(1),
	Run:     reannounceCommandRun,
}

func reannounceCommandRun(cmd *cobra.Command, args []string) {
	// get and check the flags
	verbosity := viper.GetInt("verbose")
	hash := args[0]
	// TODO: get these from config
	attempts := 30
	interval := 7
	extraAttempts := 2
	extraInterval := 30
	maxAge := 999999 //60 * 60 // 1 hour

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
		fmt.Printf("Error connecting to qBittorrent: %s\n", err)
		os.Exit(1)
	}
	if verbosity > 0 {
		fmt.Printf("Connected to qBittorrent\n")
	}

	// get torrent
	torrents, err := client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{
		Hashes: []string{hash},
	})
	if err != nil {
		fmt.Printf("Error getting torrents: %s\n", err.Error())
		os.Exit(1)
	}
	if len(torrents) != 1 {
		fmt.Printf("%s: torrent not found\n", hash)
		os.Exit(1)
	}
	torrent := torrents[0]
	if verbosity > 0 {
		fmt.Printf("%s: added=%d\n", hash, torrent.AddedOn)
	}
	now := time.Now().Unix()
	if now-torrent.AddedOn > int64(options.MaxAge) {
		fmt.Printf("%s: torrent is older than %d seconds\n", hash, options.MaxAge)
		os.Exit(1)
	}

	// reannounce
	reannounceUntilSeeds(ctx, client, torrents[0], options)
	reannounceForGoodMeasure(ctx, client, torrents[0], options)
}

func reannounceUntilSeeds(ctx context.Context, client *qbittorrent.Client, t qbittorrent.Torrent, options ReannounceOptions) bool {
	for i := 0; i < options.Attempts; i++ {
		// delay before every attempt
		if verbosity > 0 {
			fmt.Printf("try %d: Sleep %d\n", i, options.Interval)
		}
		time.Sleep(time.Duration(options.Interval) * time.Second)

		// get trackers
		trackers, err := client.GetTorrentTrackersCtx(ctx, t.Hash)
		if err != nil {
			fmt.Printf("%s: Error getting trackers: %s\n", t.Hash, err)
			os.Exit(1)
		}

		// find current tracker
		// TODO: is this wrong for public torrents with DHT and PEX?
		var tracker qbittorrent.TorrentTracker
		found := false
		for _, tr := range trackers {
			fmt.Printf("%s: u=\"%s\" status=%d peer=%d seed=%d leech=%d dl=%d msg=\"%s\"\n", t.Hash, tr.Url, tr.Status, tr.NumPeers, tr.NumSeeds, tr.NumLeechers, tr.NumDownloaded, tr.Message)
			//fmt.Printf("%s: %v\n", t.Hash, tr)
			if tr.Url == t.Tracker {
				tracker = tr
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("%s: Tracker with URL %s not found\n", t.Hash, t.Tracker)
			os.Exit(1)
		}
		fmt.Printf("%s: status %s msg %s\n", t.Hash, trackerStatus(tracker.Status), tracker.Message, tracker.Url)

		// if status not ok then reannounce
		ok, _ := isTrackerStatusOK(trackers)
		if !ok {
			forceReannounce(ctx, client, t)
			continue
		}

		// get seed information
		fmt.Printf("%s: num_seeds=%d\n", t.Hash, tracker.NumSeeds)
		if tracker.NumSeeds > 0 {
			return true
		}
	}

	fmt.Printf("%s: Reannounce attempts exhausted\n", t.Hash)
	os.Exit(1)
	return false
}

func reannounceForGoodMeasure(ctx context.Context, client *qbittorrent.Client, t qbittorrent.Torrent, options ReannounceOptions) {
	for i := 0; i < options.ExtraAttempts; i++ {
		// delay before every attempt
		if verbosity > 0 {
			fmt.Printf("try %d: Sleep %d\n", i, options.ExtraInterval)
		}
		time.Sleep(time.Duration(options.ExtraInterval) * time.Second)

		// force reannounce
		forceReannounce(ctx, client, t)
	}
}

func forceReannounce(ctx context.Context, client *qbittorrent.Client, t qbittorrent.Torrent) {
	fmt.Printf("%s: Forcing reannounce\n", t.Hash)
	// if err := client.ReAnnounceTorrentsCtx(ctx, []string{t.Hash}); err != nil {
	// 	fmt.Printf("%s: Error reannouncing: %s\n", t.Hash, err)
	// 	os.Exit(1)
	// }
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
