package client

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"example.com/db"
	"example.com/torrent"
)

type PeerConnectionBuilder interface {
	BuildPeerConnection(peer *db.Peer) (*torrent.PeerConnection, error)
}

type Client struct {
	PeerConnectionBuilder
	Client db.Client
	Port   uint16

	ClientRepo     db.ClientRepository
	TorrentRepo    db.TorrentRepository
	AnnounceRepo   db.TrackerAnnounceRepository
	PieceRepo      db.PieceRepository
	PeerRepo       db.PeerRepository
	ConnectionRepo db.ConnectionRepository

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

	fullLength, err := metaInfo.GetFullLength()
	if err != nil {
		slog.Error("Could not calculate full length of " + metaInfo.Info.Name)
		return nil, err
	}

	dbTorrent = &db.Torrent{
		Name:        metaInfo.Info.Name,
		Announce:    metaInfo.Announce,
		Size:        fullLength,
		HashInfo:    metaInfo.GetInfoHash(),
		CreatedTime: time.Now(),
		Paused:      false,
		Location:    downloadPath,
		Progress:    0,
		RawMetaInfo: metaInfo.RawBytes,
	}

	infoMsg := fmt.Sprintf("Creating new torrent record %s:%s...", dbTorrent.Name, hex.EncodeToString(dbTorrent.HashInfo))
	slog.Info(infoMsg)

	err = c.TorrentRepo.Create(dbTorrent)
	if err != nil {
		slog.Error("Error when creating torrent record.")
		return nil, err
	}

	slog.Info("Created new torrent record.")

	return dbTorrent, nil
}

func (c *Client) announceExistenceForTorrent(dbTorrent *db.Torrent) db.TrackerAnnounce {
	announceTime := time.Now()
	scheduledTime := announceTime.Add(time.Minute)

	// Fill what we already now.
	dbAnnounce := db.TrackerAnnounce{
		TorrentId:     dbTorrent.TorrentId,
		AnnounceTime:  announceTime,
		ScheduledTime: &scheduledTime,
		Done:          false,
	}

	// Form request for Tracker web server
	announceRequest := torrent.AnnounceRequest{
		AnnounceURL: dbTorrent.Announce,
		PeerId:      c.Client.ProtocolId,
		InfoHash:    dbTorrent.HashInfo,
		Port:        int(c.Port),
		Left:        dbTorrent.Size,
		// TODO
		// Fill this later.
		Uploaded:   0,
		Downloaded: 0,
	}

	announceResponse, err := torrent.Announce(&announceRequest)
	if err != nil {
		// Record an error
		errMsg := err.Error()
		dbAnnounce.Error = &errMsg
	} else {
		// Try again when tracker server said
		scheduledTime = dbAnnounce.AnnounceTime.Add(time.Second * time.Duration(announceResponse.Interval))

		dbAnnounce.RawResponse = announceResponse.RawResponse
	}

	return dbAnnounce
}

func (c *Client) Announce(dbTorrent *db.Torrent) (*db.TrackerAnnounce, error) {
	if !c.initialized {
		return nil, errors.New("Client not initialized.")
	}

	dbAnnounce := c.announceExistenceForTorrent(dbTorrent)

	err := c.AnnounceRepo.Create(&dbAnnounce)
	if err != nil {
		slog.Error("Could not save tracker announce to database.")
		return nil, err
	}

	return &dbAnnounce, nil
}

func (c *Client) ProcessTrackerAnnounce(trackerAnnounce *db.TrackerAnnounce) ([]db.Peer, error) {
	var dbPeers []db.Peer

	if trackerAnnounce.Error != nil {
		return dbPeers, errors.New("Tracker announce contains an error.")
	}

	buffer := bytes.NewBuffer(trackerAnnounce.RawResponse)

	announceResponse, err := torrent.ParseAnnounceResponse(buffer)
	if err != nil {
		errMsg := fmt.Sprintf("Could not parse announce response on %d", trackerAnnounce.TrackerAnnounceId)
		slog.Error(errMsg)
		return dbPeers, err
	}

	for i := range announceResponse.Peers {
		peer := announceResponse.Peers[i]

		dbPeer, err := c.PeerRepo.GetByTorrentIdAndProtocolPeerId(trackerAnnounce.TorrentId, peer.PeerId)
		if err != nil {
			slog.Error("Error on peer database query.")
			return dbPeers, err
		}

		if dbPeer != nil {
			debugMsg := fmt.Sprintf("Protocol Peer Id %s for TorrentId %d already exists, skipping...", hex.EncodeToString(dbPeer.ProtocolPeerId), trackerAnnounce.TrackerAnnounceId)
			slog.Debug(debugMsg)
			// TODO
			// Update port and IP maybe
			continue
		}

		infoMsg := fmt.Sprintf("Creating new peer %s...", hex.EncodeToString(peer.PeerId))
		slog.Info(infoMsg)

		newDbPeer := db.Peer{
			TorrentId:      trackerAnnounce.TorrentId,
			ProtocolPeerId: peer.PeerId,
			IP:             peer.IP.String(),
			Port:           peer.Port,
		}

		err = c.PeerRepo.Create(&newDbPeer)
		if err != nil {
			slog.Error("Could not save peer record to database")
			return dbPeers, err
		}

		infoMsg = fmt.Sprintf("Created new peer %s.", hex.EncodeToString(peer.PeerId))
		slog.Info(infoMsg)

		dbPeers = append(dbPeers, newDbPeer)
	}

	return dbPeers, nil
}

func (c *Client) ConnectToPeer(dbPeer *db.Peer) (*db.Connection, error) {
	peerInfo := fmt.Sprintf("%d:%s", dbPeer.PeerId, hex.EncodeToString(dbPeer.ProtocolPeerId))

	peerConnection, err := c.BuildPeerConnection(dbPeer)
	if err != nil {
		errMsg := fmt.Sprintf("Could not build peer connection with peer %s", peerInfo)
		slog.Error(errMsg)
		return nil, err
	}

	err = peerConnection.InitiateHandshake()
	if err != nil {
		errMsg := fmt.Sprintf("Handshake failed with peer %s", peerInfo)
		slog.Error(errMsg)

		return nil, err
	}

	infoMsg := fmt.Sprintf("Established handshake with %s", peerInfo)
	slog.Info(infoMsg)

	dbConnection := db.Connection{
		TorrentId:          dbPeer.TorrentId,
		RemotePeerId:       dbPeer.PeerId,
		ImChoked:           true,
		RemoteIsChoked:     true,
		ImInterested:       false,
		RemoteIsInterested: false,
		DownloadRate:       0,
		UploadRate:         0,
		LastActivity:       time.Now(),
	}

	err = c.ConnectionRepo.Upsert(&dbConnection)

	return &dbConnection, err
}

func (c *Client) Listen() error {
	address := ":" + strconv.Itoa(int(c.Port))

	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			errMsg := fmt.Sprintf("Error on accepting connection %v", err)
			slog.Error(errMsg)
			continue
		}

		go func(netConnection net.Conn) {

			netConnection.Close()
		}(conn)
	}

}

func (c *Client) DownloadPiece(torrent *db.Torrent, index int) error {
	return nil
}

func (c *Client) CheckDownloadDone(torrent *db.Torrent) error {
	return nil
}
