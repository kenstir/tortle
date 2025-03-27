package cmd

import (
	"context"
	"testing"

	"github.com/autobrr/go-qbittorrent"
	"github.com/stretchr/testify/assert"

	"github.com/kenstir/tortle/mocks"
)

func TestReannounce_TorrentNotFound(t *testing.T) {
	mockClient := mocks.NewQbitMockClient()
	ctx := context.Background()
	hash := "404"
	opts := ReannounceOptions{
		Attempts: 1,
	}

	mockClient.On("LoginCtx", ctx).Return(nil)
	mockClient.On("GetTorrentsCtx", ctx, qbittorrent.TorrentFilterOptions{Hashes: []string{hash}}).Return([]qbittorrent.Torrent{}, nil)

	err := qbitReannounce(ctx, mockClient, hash, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "torrent not found")

	mockClient.AssertExpectations(t)
}

/*
func TestReannounce(t *testing.T) {
	mockClient := mocks.NewQbitMockClient()
	ctx := context.Background()
	logger := log.New(os.Stdout, "", log.LstdFlags)
	hash := "testhash"
	attempts := 2
	interval := 1
	extraAttempts := 1
	extraInterval := 1
	maxAge := 60 * 60

	torrent := qbittorrent.Torrent{Hash: hash}
	trackers := []qbittorrent.TorrentTracker{
		{Status: qbittorrent.TrackerStatusNotWorking, Url: "http://tracker1.example.org"},
		{Status: qbittorrent.TrackerStatusOK, Url: "http://tracker2.example.org"},
	}

	mockClient.On("LoginCtx", ctx).Return(nil)
	mockClient.On("GetTorrents", qbittorrent.TorrentFilterOptions{Hashes: []string{hash}}).Return([]qbittorrent.Torrent{torrent}, nil)
	mockClient.On("GetTorrentTrackersCtx", ctx, hash).Return(trackers, nil)
	mockClient.On("ReAnnounceTorrentsCtx", ctx, []string{hash}).Return(nil)

	options := ReannounceOptions{
		Attempts:      attempts,
		Interval:      interval,
		ExtraAttempts: extraAttempts,
		ExtraInterval: extraInterval,
		MaxAge:        maxAge,
	}
	qbitReannounce(ctx, logger, mockClient, hash, options)

	mockClient.AssertExpectations(t)
}
*/

/*
func TestReannounce_ErrorGettingTorrents(t *testing.T) {
	mockClient := mocks.NewQbitMockClient()
	ctx := context.Background()
	hash := "testhash"
	attempts := 3
	interval := 1
	verbosity := 1

	mockClient.On("LoginCtx", ctx).Return(nil)
	mockClient.On("GetTorrents", qbittorrent.TorrentFilterOptions{Hashes: []string{hash}}).Return(nil, assert.AnError)

	assert.Panics(t, func() {
		qbitReannounce(ctx, mockClient, hash, attempts, interval, verbosity)
	})

	mockClient.AssertExpectations(t)
}
*/
