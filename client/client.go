package client

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
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
	SeederBuilder
	Client db.Client
	Port   uint16

	ClientRepo   db.ClientRepository
	TorrentRepo  db.TorrentRepository
	AnnounceRepo db.TrackerAnnounceRepository
	PieceRepo    db.PieceRepository
	PeerRepo     db.PeerRepository

	initialized bool
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
	c.initialized = true

	return nil
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
		slog.Error("Error creating directory " + directoryPath)
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

	dbTorrent = &db.Torrent{
		Name:        metaInfo.Info.Name,
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

func (c *Client) Announce(dbTorrent *db.Torrent) (*db.TrackerAnnounce, error) {
	if !c.initialized {
		return nil, errors.New("Client not initialized.")
	}

	rawMetaInfoBuffer := bytes.NewBuffer(dbTorrent.RawMetaInfo)
	metaInfo, err := torrent.ParseMetaInfo(rawMetaInfoBuffer)

	if err != nil {
		errMsg := fmt.Sprintf("Invalid RawMetaInfo in database, record #%d", dbTorrent.TorrentId)
		slog.Error(errMsg)

		return nil, err
	}

	// Make this more readable
	// -----------------------
	dbAnnounce := db.TrackerAnnounce{
		TorrentId:    dbTorrent.TorrentId,
		AnnounceTime: time.Now(),
		Done:         false,
	}

	var nextAnnounceTime time.Time

	announceResponse, err := torrent.Announce(c.Client.ProtocolId, int(c.Port), metaInfo)
	if err != nil {
		// Record an error
		errMsg := err.Error()
		dbAnnounce.Error = &errMsg

		// Try again in a minute
		nextAnnounceTime = dbAnnounce.AnnounceTime.Add(time.Minute)
	} else {
		// Try again when tracker server said
		nextAnnounceTime = dbAnnounce.AnnounceTime.Add(time.Second * time.Duration(announceResponse.Interval))

		dbAnnounce.RawResponse = announceResponse.RawResponse
	}

	dbAnnounce.ScheduledTime = &nextAnnounceTime
	// -----------------------

	err = c.AnnounceRepo.Create(&dbAnnounce)
	if err != nil {
		slog.Error("Could not save tracker announce to database.")
		return nil, err
	}

	if announceResponse == nil {
		return &dbAnnounce, nil
	}

	// TODO
	// Will move this to be done in background later on

	for i := range announceResponse.Peers {
		peer := announceResponse.Peers[i]

		dbPeer, err := c.PeerRepo.GetByTorrentIdAndProtocolPeerId(dbAnnounce.TorrentId, peer.PeerId)
		if err != nil {
			slog.Error("Error on peer database query.")
			return &dbAnnounce, err
		}

		if dbPeer != nil {
			debugMsg := fmt.Sprintf("Protocol Peer Id %s for TorrentId %d already exists, skipping...", hex.EncodeToString(dbPeer.ProtocolPeerId), dbAnnounce.TorrentId)
			slog.Debug(debugMsg)

			continue
		}

		infoMsg := fmt.Sprintf("Creating new peer %s...", hex.EncodeToString(peer.PeerId))
		slog.Info(infoMsg)

		dbPeer = &db.Peer{
			TorrentId:      dbAnnounce.TorrentId,
			ProtocolPeerId: peer.PeerId,
			IP:             peer.IP.String(),
			Port:           peer.Port,
			// Assume is reachable for now
			Reachable: true,
		}

		err = c.PeerRepo.Create(dbPeer)
		if err != nil {
			slog.Error("Could not save peer record to database")
			return &dbAnnounce, err
		}

		infoMsg = fmt.Sprintf("Created new peer %s.", hex.EncodeToString(peer.PeerId))
		slog.Info(infoMsg)
	}

	return &dbAnnounce, nil
}

func (c *Client) DownloadPiece(torrent *db.Torrent, index int) error {
	return nil
}

func (c *Client) CheckDownloadDone(torrent *db.Torrent) error {
	return nil
}
