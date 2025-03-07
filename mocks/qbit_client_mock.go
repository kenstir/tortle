package mocks

import (
	"context"

	"github.com/autobrr/go-qbittorrent"
	"github.com/stretchr/testify/mock"
)

// QbitMockClient is a mocked object implementing the QbitClientInterface
type QbitMockClient struct {
	mock.Mock
}

func NewQbitMockClient() *QbitMockClient {
	return &QbitMockClient{}
}

func (_m *QbitMockClient) LoginCtx(ctx context.Context) error {
	args := _m.Called(ctx)
	return args.Error(0)
}

func (_m *QbitMockClient) GetTorrentsCtx(ctx context.Context, filter qbittorrent.TorrentFilterOptions) ([]qbittorrent.Torrent, error) {
	args := _m.Called(ctx, filter)
	return args.Get(0).([]qbittorrent.Torrent), args.Error(1)
}

func (_m *QbitMockClient) GetTorrentTrackersCtx(ctx context.Context, hash string) ([]qbittorrent.TorrentTracker, error) {
	args := _m.Called(ctx, hash)
	return args.Get(0).([]qbittorrent.TorrentTracker), args.Error(1)
}

func (_m *QbitMockClient) GetTorrentPropertiesCtx(ctx context.Context, hash string) (qbittorrent.TorrentProperties, error) {
	args := _m.Called(ctx, hash)
	return args.Get(0).(qbittorrent.TorrentProperties), args.Error(1)
}

func (_m *QbitMockClient) ReAnnounceTorrentsCtx(ctx context.Context, hashes []string) error {
	args := _m.Called(ctx, hashes)
	return args.Error(0)
}
