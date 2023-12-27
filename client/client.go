package client

import (
	"io"
	"log"
	"log/slog"
	"time"

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
	clientDb, err := c.ClientRepo.GetLast()
	if err != nil {
		slog.Error("Could not retrieve client record.")
		return err
	}

	if clientDb == nil {
		log.Print("First time running. Creating client record...")

		clientDb = &db.Client{
			ProtocolId: torrent.GenerateRandomProtocolId(),
			Created:    time.Now(),
		}

		err := c.ClientRepo.Create(clientDb)
		if err != nil {
			slog.Error("Could not create client record.")
			return err
		}

		slog.Info("Created client record.")
	}

	c.Client = *clientDb
	return err
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
