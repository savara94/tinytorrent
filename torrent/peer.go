package torrent

import (
	"errors"
	"io"
)

type Peer struct {
	PeerId       []byte
	Port         int
	MetaInfoList []MetaInfo
}

const HandshakeMsg = "\x19" + "BitTorrent protocol"

const (
	Choke byte = iota
	Unchoke
	Interested
	NotInterested
	Have
	Bitfield
	Request
	Piece
	Cancel
	KeepAlive
)

type PeerMessage struct {
	Type    byte
	Payload any
}

var ChokeMessage = PeerMessage{Type: Choke, Payload: nil}
var UnchokeMessage = PeerMessage{Type: Unchoke, Payload: nil}
var InterestedMessage = PeerMessage{Type: Interested, Payload: nil}
var NotInterestedMessage = PeerMessage{Type: NotInterested, Payload: nil}
var KeepAliveMessage = PeerMessage{Type: KeepAlive, Payload: nil}

type BitfieldPayload struct {
	Bitfield []byte
}

type HavePayload struct {
	Index int32
}

type RequestPayload struct {
	Index  int32
	Begin  int32
	Length int32
}

type CancelPayload struct {
	Index  int32
	Begin  int32
	Length int32
}

type PiecePayload struct {
	Index int32
	Begin int32
	Piece []byte
}

type Seeder struct {
	SeederInfo   PeerInfo
	SeederWriter io.Writer
	SeederReader io.Reader
	MetaInfo     *MetaInfo
}

func (seeder *Seeder) InitiateHandshake() error {
	n, err := seeder.SeederWriter.Write([]byte(HandshakeMsg))
	if err != nil {
		return err
	}

	if n < len(HandshakeMsg) {
		return errors.New("Couldn't send handshake bytes")
	}

	// 8 empty bytes are next
	reservedBytes := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	n, err = seeder.SeederWriter.Write(reservedBytes)
	if err != nil {
		return err
	}

	if n < len(reservedBytes) {
		return errors.New("Couldn't send reserved bytes.")
	}

	infoHash := seeder.MetaInfo.GetInfoHash()
	n, err = seeder.SeederWriter.Write(infoHash)
	if err != nil {
		return err
	}

	if n < len(infoHash) {
		return errors.New("Couldn't send infohash bytes.")
	}

	n, err = seeder.SeederWriter.Write(seeder.SeederInfo.PeerId)
	if err != nil {
		return err
	}

	if n < len(seeder.SeederInfo.PeerId) {
		return errors.New("Couldn't send peer id bytes.")
	}

	// If seeder agrees, it won't severe the connection.
	return nil
}
