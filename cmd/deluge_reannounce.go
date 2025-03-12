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

	"github.com/autobrr/go-deluge"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	delugeCmd.AddCommand(delugeReannounceCmd)

	delugeReannounceCmd.Flags().IntP("attempts", "a", 60, "Number of reannounce attempts")
	delugeReannounceCmd.Flags().IntP("interval", "i", 7, "Interval between reannounce attempts")
	delugeReannounceCmd.Flags().IntP("extra_attempts", "A", 2, "Number of extra reannounce attempts")
	delugeReannounceCmd.Flags().IntP("extra_interval", "I", 30, "Interval between extra reannounce attempts")
	delugeReannounceCmd.Flags().IntP("max_age", "m", 60*60, "Maximum age of torrent in seconds")
	viper.BindPFlag("deluge.reannounce.attempts", delugeReannounceCmd.Flags().Lookup("attempts"))
	viper.BindPFlag("deluge.reannounce.interval", delugeReannounceCmd.Flags().Lookup("interval"))
	viper.BindPFlag("deluge.reannounce.extra_attempts", delugeReannounceCmd.Flags().Lookup("extra_attempts"))
	viper.BindPFlag("deluge.reannounce.extra_interval", delugeReannounceCmd.Flags().Lookup("extra_interval"))
	viper.BindPFlag("deluge.reannounce.max_age", delugeReannounceCmd.Flags().Lookup("max_age"))
}

var delugeReannounceCmd = &cobra.Command{
	Use:     "reannounce hash",
	Aliases: []string{"re", "fast_start", "faststart", "start"},
	Short:   "Reannounce torrent",
	Args:    cobra.ExactArgs(1),
	Run:     delugeReannounceCmdRun,
}

func delugeReannounceCmdRun(cmd *cobra.Command, args []string) {
	hash := args[0]

	// get the flags
	attempts := viper.GetInt("deluge.reannounce.attempts")
	interval := viper.GetInt("deluge.reannounce.interval")
	extraAttempts := viper.GetInt("deluge.reannounce.extra_attempts")
	extraInterval := viper.GetInt("deluge.reannounce.extra_interval")
	maxAge := viper.GetInt("deluge.reannounce.max_age")

	// create a deluge client
	client := delugeCreateV2Client()

	// reannounce
	options := ReannounceOptions{
		Attempts:      attempts,
		Interval:      interval,
		ExtraAttempts: extraAttempts,
		ExtraInterval: extraInterval,
		MaxAge:        maxAge,
	}
	err := delugeReannounce(context.Background(), client, hash, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func delugeReannounce(ctx context.Context, client deluge.DelugeClient, hash string, opts ReannounceOptions) error {

	// connect
	err := client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	if verbosity > 0 {
		stdoutLogger.Printf("Connected to deluge\n")
	}

	// get torrent status
	ts, err := delugeGetTorrentStatus(ctx, client, hash)
	if err != nil {
		return err
	}

	// perform startup checks
	age := time.Now().Unix() - int64(ts.TimeAdded)
	stdoutLogger.Printf("%s: found torrent age=%d\n", hash, age)
	if age > int64(opts.MaxAge) {
		return fmt.Errorf("%s: torrent is %ds old, max_age is %ds", hash, age, opts.MaxAge)
	}
	// if torrent.CompletionOn > 0 {
	// 	stdoutLogger.Printf("%s: torrent is finished\n", hash)
	// 	return
	// }

	// reannounce
	err = delugeReannounceUntilSeeded(ctx, client, hash, opts)
	if err != nil {
		return err
	}
	err = delugeReannounceForGoodMeasure(ctx, client, hash, opts)
	if err != nil {
		return err
	}

	// log final status
	ts, err = delugeGetTorrentStatus(ctx, client, hash)
	if err != nil {
		return err
	}
	delugeLogTorrentStatus(ctx, ts, "final")

	return nil
}

func delugeReannounceUntilSeeded(ctx context.Context, client deluge.DelugeClient, hash string, options ReannounceOptions) error {
	for i := 1; i <= options.Attempts; i++ {
		prefix := fmt.Sprintf("try %d", i)

		// delay before every attempt
		if verbosity > 0 {
			stdoutLogger.Printf("%s: %s: Sleep %d\n", hash, prefix, options.Interval)
		}
		time.Sleep(time.Duration(options.Interval) * time.Second)

		// get torrent status
		ts, err := delugeGetTorrentStatus(ctx, client, hash)
		if err != nil {
			return err
		}
		delugeLogTorrentStatus(ctx, ts, prefix)

		// maybe skip or reannounce based on status
		skipReannounce, ok := delugeCheckStatus(ts)
		if skipReannounce {
			stdoutLogger.Printf("%s: %s: skipping reannounce\n", hash, prefix)
		} else if ok && (ts.NumSeeds > 0 || ts.TotalSeeds > 0) {
			stdoutLogger.Printf("%s: %s: torrent is OK with seeds=%d total_seeds=%d\n", hash, prefix, ts.NumSeeds, ts.TotalSeeds)
			return nil
		} else {
			delugeForceReannounce(ctx, client, hash, prefix)
		}
	}

	return fmt.Errorf("%s: Reannounce attempts exhausted", hash)
}

func delugeReannounceForGoodMeasure(ctx context.Context, client deluge.DelugeClient, hash string, options ReannounceOptions) error {
	for i := 1; i <= options.ExtraAttempts; i++ {
		prefix := fmt.Sprintf("extra %d", i)

		// delay before every attempt
		if verbosity > 0 {
			stdoutLogger.Printf("%s: %s: sleep %d\n", hash, prefix, options.ExtraInterval)
		}
		time.Sleep(time.Duration(options.ExtraInterval) * time.Second)

		// log torrent status
		ts, err := delugeGetTorrentStatus(ctx, client, hash)
		if err != nil {
			return err
		}
		delugeLogTorrentStatus(ctx, ts, prefix)

		// then reannounce
		delugeForceReannounce(ctx, client, hash, prefix)
	}

	return nil
}

func delugeGetTorrentStatus(ctx context.Context, client deluge.DelugeClient, hash string) (*deluge.TorrentStatus, error) {
	torrentsStatus, err := client.TorrentsStatus(ctx, deluge.StateUnspecified, []string{hash})
	if err != nil {
		return nil, err
	}
	if len(torrentsStatus) != 1 {
		return nil, fmt.Errorf("%s: torrent not found", hash)
	}
	return torrentsStatus[hash], nil
}

// / delugeCheckStatus returns (skipReannounce, isOK)
func delugeCheckStatus(ts *deluge.TorrentStatus) (bool, bool) {
	msg := strings.ToLower(ts.TrackerStatus)

	skipWords := []string{"announce set", "too many requests"}
	for _, v := range skipWords {
		if strings.Contains(msg, v) {
			return true, false
		}
	}

	notOKWords := []string{"unregistered", "end of file", "bad gateway", "error"}
	for _, v := range notOKWords {
		if strings.Contains(msg, v) {
			return false, false
		}
	}

	return false, true
}

func delugeForceReannounce(ctx context.Context, client deluge.DelugeClient, hash string, prefix string) {
	if err := client.ForceReannounce(ctx, []string{hash}); err != nil {
		stdoutLogger.Printf("%s: Error reannouncing: %s\n", hash, err)
	} else {
		stdoutLogger.Printf("%s: %s: reannounce sent\n", hash, prefix)
	}
}

func delugeLogTorrentStatus(ctx context.Context, ts *deluge.TorrentStatus, prefix string) {
	duration := time.Duration(ts.NextAnnounce) * time.Second
	progress := int(ts.Progress + 0.5)
	stdoutLogger.Printf("%s: %s: torrent: status=\"%s\" seeds=%d total_seeds=%d peer=%d progress=%d%% reannounce=%d(%s)\n", ts.Hash, prefix, ts.TrackerStatus, ts.NumSeeds, ts.TotalSeeds, ts.TotalPeers, progress, ts.NextAnnounce, duration.String())
}
