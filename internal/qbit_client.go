/*
Copyright © 2025 Kenneth H. Cox
*/

package internal

import (
	"context"

	"github.com/autobrr/go-qbittorrent"
)

type QbitClientInterface interface {
	LoginCtx(context.Context) error
	GetTorrentsCtx(context.Context, qbittorrent.TorrentFilterOptions) ([]qbittorrent.Torrent, error)
	GetTorrentTrackersCtx(context.Context, string) ([]qbittorrent.TorrentTracker, error)
	ReAnnounceTorrentsCtx(context.Context, []string) error
}

type QbitClient struct {
	client *qbittorrent.Client
}

func NewQbitClient(cfg qbittorrent.Config) *QbitClient {
	client := qbittorrent.NewClient(cfg)
	return &QbitClient{
		client: client,
	}
}

func (qc *QbitClient) LoginCtx(ctx context.Context) error {
	return qc.client.LoginCtx(ctx)
}

func (qc *QbitClient) GetTorrentsCtx(ctx context.Context, filter qbittorrent.TorrentFilterOptions) ([]qbittorrent.Torrent, error) {
	return qc.client.GetTorrentsCtx(ctx, filter)
}

func (qc *QbitClient) GetTorrentTrackersCtx(ctx context.Context, hash string) ([]qbittorrent.TorrentTracker, error) {
	return qc.client.GetTorrentTrackersCtx(ctx, hash)
}

func (qc *QbitClient) ReAnnounceTorrentsCtx(ctx context.Context, hashes []string) error {
	return qc.client.ReAnnounceTorrentsCtx(ctx, hashes)
}
