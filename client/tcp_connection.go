package client

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"example.com/db"
	"example.com/torrent"
)

type TcpConnector struct {
	ListenPort       int
	ClientProtocolId []byte

	listener net.Listener
}

func NewTcpConnector(clientProtocolId []byte, port int) (*TcpConnector, error) {
	address := ":" + strconv.Itoa(int(port))

	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &TcpConnector{
		ListenPort:       port,
		ClientProtocolId: clientProtocolId,
		listener:         l,
	}, nil
}

func (c *TcpConnector) Accept() (*torrent.PeerConnection, error) {
	conn, err := c.listener.Accept()
	if err != nil {
		return nil, err
	}

	return &torrent.PeerConnection{
		PeerId:     c.ClientProtocolId,
		PeerWriter: conn,
		PeerReader: conn,
		IsConnectionSevered: func() (bool, error) {
			// TODO
			// Check this as well.
			_, err := conn.Write([]byte{})

			return false, err
		},
		SevereConnection: conn.Close,
	}, nil
}

func (c *TcpConnector) Connect(peer *db.Peer) (*torrent.PeerConnection, error) {
	address := fmt.Sprintf("%s:%d", peer.IP, peer.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		slog.Error("Error connecting to " + peer.IP + ":" + strconv.Itoa(peer.Port))
		return nil, err
	}

	return &torrent.PeerConnection{
		PeerId:       c.ClientProtocolId,
		RemotePeerId: peer.ProtocolPeerId,
		PeerWriter:   conn,
		PeerReader:   conn,
		IsConnectionSevered: func() (bool, error) {
			// TODO
			// Check this as well.
			_, err := conn.Write([]byte{})

			return false, err
		},
		SevereConnection: conn.Close,
	}, nil
}

func (c *TcpConnector) Close() {
	c.listener.Close()
}
