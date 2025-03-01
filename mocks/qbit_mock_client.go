package mocks

import (
	"context"

	"github.com/autobrr/go-qbittorrent"
	"github.com/stretchr/testify/mock"
)

// QbitMockClient is a mocked object implementing the qbittorrent.Client interface
type QbitMockClient struct {
	mock.Mock
}

func (_m *QbitMockClient) LoginCtx(ctx context.Context) error {
	args := _m.Called(ctx)
	return args.Error(0)
}

func (_m *QbitMockClient) GetTorrentsCtx(ctx context.Context, o qbittorrent.TorrentFilterOptions) ([]qbittorrent.Torrent, error) {
	args := _m.Called(ctx, o)
	return args.Get(0).([]qbittorrent.Torrent), args.Error(1)
}

func (_m *QbitMockClient) GetTorrentsTrackerCtx(ctx context.Context, hash string) ([]qbittorrent.TorrentTracker, error) {
	args := _m.Called(ctx, hash)
	return args.Get(0).([]qbittorrent.TorrentTracker), args.Error(1)
}

func (_m *QbitMockClient) ReAnnounceTorrentsCtx(ctx context.Context, hashes []string) error {
	args := _m.Called(ctx, hashes)
	return args.Error(0)
}
