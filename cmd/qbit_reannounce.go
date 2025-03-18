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
	Use:     "reannounce hash",
	Aliases: []string{"re", "reann", "faststart", "start"},
	Short:   "Reannounce torrent until healthy",
	Args:    cobra.ExactArgs(1),
	Run:     qbitReannounceCmdRun,
}

var stdoutLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)
var stderrLogger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds)

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
	err = qbitReannounceUntilOK(ctx, client, hash, opts)
	if err != nil {
		return err
	}
	err = qbitReannounceForGoodMeasure(ctx, client, hash, opts)
	if err != nil {
		return err
	}

	// log final state
	qbitLogTorrentProperties(ctx, client, hash, "final")

	return nil
}

func qbitReannounceUntilOK(ctx context.Context, client internal.QbitClientInterface, hash string, options ReannounceOptions) error {
	for i := 1; i <= options.Attempts; i++ {
		prefix := fmt.Sprintf("try %d", i)

		// delay before every attempt
		if verbosity > 0 {
			stdoutLogger.Printf("%s: %s: sleep %d\n", hash, prefix, options.Interval)
		}
		time.Sleep(time.Duration(options.Interval) * time.Second)

		// get trackers
		trackers, err := client.GetTorrentTrackersCtx(ctx, hash)
		if err != nil {
			return err
		}
		if trackers == nil {
			return fmt.Errorf("%s: no trackers?", hash)
		}

		// if status not ok then reannounce
		ok, seeds := findOKTracker(trackers, hash, prefix)
		if !ok {
			qbitLogTorrentProperties(ctx, client, hash, prefix)
			qbitForceReannounce(ctx, client, hash, prefix)
			continue
		}

		stdoutLogger.Printf("%s: %s: torrent is OK with %d seeds\n", hash, prefix, seeds)
		return nil
	}

	return fmt.Errorf("%s: Reannounce attempts exhausted", hash)
}

func qbitReannounceForGoodMeasure(ctx context.Context, client internal.QbitClientInterface, hash string, options ReannounceOptions) error {
	for i := 1; i <= options.ExtraAttempts; i++ {
		prefix := fmt.Sprintf("extra %d", i)

		// delay before every attempt
		if verbosity > 0 {
			stdoutLogger.Printf("%s: %s: sleep %d\n", hash, prefix, options.ExtraInterval)
		}
		time.Sleep(time.Duration(options.ExtraInterval) * time.Second)

		// log state then reannounce
		qbitLogTorrentProperties(ctx, client, hash, prefix)
		qbitForceReannounce(ctx, client, hash, prefix)
	}

	return nil
}

func qbitForceReannounce(ctx context.Context, client internal.QbitClientInterface, hash string, prefix string) {
	if err := client.ReAnnounceTorrentsCtx(ctx, []string{hash}); err != nil {
		stdoutLogger.Printf("%s: Error reannouncing: %s\n", hash, err)
	} else {
		stdoutLogger.Printf("%s: %s: reannounce requested\n", hash, prefix)
	}
}

func qbitLogTorrentProperties(ctx context.Context, client internal.QbitClientInterface, hash string, prefix string) {
	props, err := client.GetTorrentPropertiesCtx(ctx, hash)
	if err != nil {
		stdoutLogger.Printf("%s: Error getting properties: %s\n", hash, err)
	}
	duration := time.Duration(props.Reannounce) * time.Second
	stdoutLogger.Printf("%s: %s: torrent: seed=%d peer=%d pieces=%d/%d(%d%%) reannounce=%d(%s)\n", hash, prefix, props.SeedsTotal, props.PeersTotal, props.PiecesHave, props.PiecesNum, int(100*props.PiecesHave/props.PiecesNum), props.Reannounce, duration.String())
}

// Return true if a tracker is OK
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
func findOKTracker(trackers []qbittorrent.TorrentTracker, hash string, prefix string) (bool, int) {
	// until I am confident in the logic below, print the status of every enabled tracker
	for i, tr := range trackers {
		if tr.Status == qbittorrent.TrackerStatusDisabled {
			continue
		}
		hostname := strings.Split(tr.Url, "/")[2]
		stdoutLogger.Printf("%s: %s: trackers[%d]: status=%s seed=%d peer=%d msg=\"%s\" u=%s\n", hash, prefix, i, trackerStatus(tr.Status), tr.NumSeeds, tr.NumPeers, tr.Message, hostname)
	}

	// find the first tracker with an OK status and seeds
	for _, tr := range trackers {
		if tr.Status == qbittorrent.TrackerStatusOK {
			return true, tr.NumSeeds
		}
	}

	return false, -1
}

// trackerStatus returns a string representation of the tracker status
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
