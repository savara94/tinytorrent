package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"example.com/db"
	"example.com/torrent"
)

type PeerConnector interface {
	Connect(peer *db.Peer) (*torrent.PeerConnection, error)
}

type PeerListener interface {
	Accept() (*torrent.PeerConnection, error)
	Close() error
}

type Connection struct {
	PeerConnection *torrent.PeerConnection
	Info           *db.Connection
}

type ConnectionManager struct {
	PeerConnector
	PeerListener
	ProtocolId     []byte
	TorrentRepo    db.TorrentRepository
	PeerRepo       db.PeerRepository
	ConnectionRepo db.ConnectionRepository

	peerConnMap map[string]*Connection
}

func (cm *ConnectionManager) makeKey(peerId []byte, infoHash []byte) string {
	hexPeerProtocolId := hex.EncodeToString(peerId)
	hexInfoHash := hex.EncodeToString(infoHash)

	key := fmt.Sprintf("%s:%s", hexPeerProtocolId, hexInfoHash)

	return key
}

func (cm *ConnectionManager) makeDefaultConnectionInfo(torrentId int, remotePeerId int) *db.Connection {
	return &db.Connection{
		TorrentId:          torrentId,
		RemotePeerId:       remotePeerId,
		ImChoked:           true,
		RemoteIsChoked:     true,
		ImInterested:       false,
		RemoteIsInterested: false,
		DownloadRate:       0,
		UploadRate:         0,
		LastActivity:       time.Now(),
	}
}

func (cm *ConnectionManager) GetOrConnect(peer *db.Peer, infoHash []byte) (*Connection, error) {
	// TODO
	// Make this thread safe

	key := cm.makeKey(peer.ProtocolPeerId, infoHash)

	connection, ok := cm.peerConnMap[key]
	if ok {
		slog.Info("Connection already exists, returning handle...")
		return connection, nil
	}

	slog.Info("Connection does not exist, will try to make one...")

	peerConnection, err := cm.Connect(peer)
	if err != nil {
		slog.Error("Could not make a connection.")
		return nil, err
	}

	peerConnection.InfoHash = infoHash

	err = peerConnection.InitiateHandshake()
	if err != nil {
		slog.Error("Handshake failed with peer " + peer.IP)
		return nil, err
	}

	slog.Info("Established handshake with " + peer.IP)

	connectionInfo := cm.makeDefaultConnectionInfo(peer.TorrentId, peer.PeerId)

	err = cm.ConnectionRepo.Upsert(connectionInfo)
	if err != nil {
		slog.Error("Error on connection upsert for peer" + peer.IP)
		return nil, err
	}

	connection = &Connection{
		PeerConnection: peerConnection,
		Info:           connectionInfo,
	}

	// TODO
	// Add method that will do this in a thread safe manner.

	cm.peerConnMap[key] = connection

	// TODO
	// Run goroutine that waits for messages

	return connection, nil
}

func (cm *ConnectionManager) Listen() (*Connection, error) {
	peerConnection, err := cm.Accept()
	if err != nil {
		return nil, err
	}

	err = peerConnection.AcceptHandshake()
	if err != nil {
		return nil, err
	}

	torrent, err := cm.TorrentRepo.GetByHashInfo(peerConnection.InfoHash)
	if err != nil {
		return nil, err
	}

	if torrent == nil {
		errMsg := fmt.Sprintf("I don't have that torrent %s, sorry.", hex.EncodeToString(peerConnection.InfoHash))

		severeConnectionErr := peerConnection.SevereConnection()
		if severeConnectionErr != nil {
			slog.Error("Error on severe connection.")
		}

		return nil, errors.New(errMsg)
	}

	remotePeer, err := cm.PeerRepo.GetByTorrentIdAndProtocolPeerId(torrent.TorrentId, peerConnection.RemotePeerId)
	if err != nil {
		slog.Error("Could not fetch peer from database")
		return nil, err
	}

	if remotePeer == nil {
		// Peer knows about me, but I don't know about him, let's register him. :)
		remotePeer = &db.Peer{
			TorrentId:      torrent.TorrentId,
			ProtocolPeerId: peerConnection.RemotePeerId,
			// TODO
			// Find out these through connection factory
			// IP: "",
			// Port: 0,
		}

		err = cm.PeerRepo.Create(remotePeer)
		if err != nil {
			slog.Error("Could not register new peer.")
			return nil, err
		}
	}

	connectionInfo := cm.makeDefaultConnectionInfo(torrent.TorrentId, remotePeer.PeerId)
	err = cm.ConnectionRepo.Upsert(connectionInfo)
	if err != nil {
		slog.Error("Could not upsert connection upon accept.")
		return nil, err
	}

	connection := &Connection{
		PeerConnection: peerConnection,
		Info:           connectionInfo,
	}

	// TODO
	// Add method that will do this in a thread safe manner.

	key := cm.makeKey(peerConnection.RemotePeerId, peerConnection.InfoHash)
	cm.peerConnMap[key] = connection

	// TODO
	// Run go-routine that waits for messages.

	return connection, nil
}

func (cm *ConnectionManager) Close() error {
	return cm.PeerListener.Close()
}
