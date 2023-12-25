package client

import (
	"io"

	"example.com/db"
	"example.com/torrent"
)

type SeederBuilder interface {
	BuildSeeder(torrent *db.Torrent, pieceIndex int) (torrent.Seeder, error)
}

type Client struct {
	Client db.Client
	Port   uint16

	ClientRepo   db.ClientRepository
	TorrentRepo  db.TorrentRepository
	AnnounceRepo db.TrackerAnnounceRepository
	PieceRepo    db.PieceRepository
	PeerRepo     db.PeerRepository

	SeederBuilder SeederBuilder
}

func (c *Client) Initialize() error {
	return nil
}

func (c *Client) OpenTorrent(reader io.Reader) (*db.Torrent, error) {
	return nil, nil
}

func (c *Client) Announce(trackerWriter io.WriteCloser, trackerReader io.ReadCloser) (*db.Torrent, error) {
	return nil, nil
}

func (c *Client) DownloadPiece(torrent *db.Torrent, index int) error {
	return nil
}

func (c *Client) CheckDownloadDone(torrent *db.Torrent) error {
	return nil
}
