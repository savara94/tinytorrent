package client

import (
	"encoding/hex"
	"errors"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
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

func (c *Client) OpenTorrent(reader io.Reader, downloadPath string) (*db.Torrent, error) {
	metaInfo, err := torrent.ParseMetaInfo(reader)
	if err != nil {
		slog.Error("This can't be parsed as a torrent file.")
		return nil, err
	}

	directoryPath := path.Join(downloadPath, metaInfo.Info.Name)

	err = os.Mkdir(directoryPath, os.ModeDir)
	if err != nil {
		slog.Info("Error creating directory " + directoryPath)
		return nil, err
	}

	infoHash := metaInfo.GetInfoHash()
	dbTorrent, err := c.TorrentRepo.GetByHashInfo(infoHash)
	if err != nil {
		slog.Error("Could not retrieve by infohash " + hex.EncodeToString(infoHash))
		return nil, err
	}

	if dbTorrent != nil {
		return dbTorrent, errors.New("Torrent already exists.")
	}

	slog.Info("Creating new torrent record " + metaInfo.Info.Name + "...")

	dbTorrent = &db.Torrent{
		HashInfo:    metaInfo.GetInfoHash(),
		CreatedTime: time.Now(),
		Paused:      false,
		Location:    downloadPath,
		Progress:    0,
		RawMetaInfo: metaInfo.RawBytes,
	}

	err = c.TorrentRepo.Create(dbTorrent)
	if err != nil {
		slog.Error("Error when creating torrent record.")
		return nil, err
	}

	slog.Info("Created new torrent record.")

	return dbTorrent, nil
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
