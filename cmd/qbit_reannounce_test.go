package cmd

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/autobrr/go-qbittorrent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) LoginCtx(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) GetTorrents(filter qbittorrent.TorrentFilterOptions) ([]qbittorrent.Torrent, error) {
	args := m.Called(filter)
	return args.Get(0).([]qbittorrent.Torrent), args.Error(1)
}

func (m *MockClient) GetTorrentTrackersCtx(ctx context.Context, hash string) ([]qbittorrent.TorrentTracker, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).([]qbittorrent.TorrentTracker), args.Error(1)
}

func (m *MockClient) ReAnnounceTorrentsCtx(ctx context.Context, hashes []string) error {
	args := m.Called(ctx, hashes)
	return args.Error(0)
}

func TestReannounce(t *testing.T) {
	mockClient := new(MockClient)
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
	reannounce(ctx, logger, mockClient, hash, options)

	mockClient.AssertExpectations(t)
}

func TestReannounce_TorrentNotFound(t *testing.T) {
	mockClient := new(MockClient)
	ctx := context.Background()
	hash := "testhash"
	attempts := 3
	interval := 1
	verbosity := 1

	mockClient.On("LoginCtx", ctx).Return(nil)
	mockClient.On("GetTorrents", qbittorrent.TorrentFilterOptions{Hashes: []string{hash}}).Return([]qbittorrent.Torrent{}, nil)

	assert.Panics(t, func() {
		reannounce(ctx, mockClient, hash, attempts, interval, verbosity)
	})

	mockClient.AssertExpectations(t)
}

func TestReannounce_ErrorGettingTorrents(t *testing.T) {
	mockClient := new(MockClient)
	ctx := context.Background()
	hash := "testhash"
	attempts := 3
	interval := 1
	verbosity := 1

	mockClient.On("LoginCtx", ctx).Return(nil)
	mockClient.On("GetTorrents", qbittorrent.TorrentFilterOptions{Hashes: []string{hash}}).Return(nil, assert.AnError)

	assert.Panics(t, func() {
		reannounce(ctx, mockClient, hash, attempts, interval, verbosity)
	})

	mockClient.AssertExpectations(t)
}
