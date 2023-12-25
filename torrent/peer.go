package torrent

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
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

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateRandomProtocolId() []byte {
	b := make([]byte, 20)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return b
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

func Send(writer io.Writer, message *PeerMessage) error {
	if message.Type == KeepAlive {
		_, err := writer.Write([]byte{})
		return err
	}

	typeByte := []byte{message.Type}
	buffer := bytes.NewBuffer(typeByte)

	// Todo
	// Make this a little bit better :)

	switch payload := message.Payload.(type) {
	case RequestPayload:
		binary.Write(buffer, binary.BigEndian, payload)
	case HavePayload:
		binary.Write(buffer, binary.BigEndian, payload)
	case PiecePayload:
		binary.Write(buffer, binary.BigEndian, payload.Index)
		binary.Write(buffer, binary.BigEndian, payload.Begin)
		binary.Write(buffer, binary.BigEndian, payload.Piece)
	case BitfieldPayload:
		binary.Write(buffer, binary.BigEndian, payload.Bitfield)
	case CancelPayload:
		binary.Write(buffer, binary.BigEndian, payload)
	}

	bytesToSend := buffer.Bytes()

	n, err := writer.Write(bytesToSend)
	if err != nil {
		return err
	}

	if n < len(bytesToSend) {
		errMsg := fmt.Sprintf("Couldn't send %#v", message)
		return errors.New(errMsg)
	}

	return nil
}

func Receive(reader io.Reader, allocator func(msgType byte) []byte) (*PeerMessage, error) {
	typeByte := make([]byte, 1)

	n, err := reader.Read(typeByte)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return &KeepAliveMessage, nil
	}

	var payload any

	// TODO
	// Write this better

	switch typeByte[0] {
	case Choke:
		return &ChokeMessage, nil
	case Unchoke:
		return &UnchokeMessage, nil
	case Interested:
		return &InterestedMessage, nil
	case NotInterested:
		return &NotInterestedMessage, nil
	case Have:
		havePayload := HavePayload{}
		binary.Read(reader, binary.BigEndian, &havePayload)
		payload = havePayload
		break
	case Request:
		rqPayload := RequestPayload{}
		binary.Read(reader, binary.BigEndian, &rqPayload)
		payload = rqPayload
		break
	case Cancel:
		cancPayload := CancelPayload{}
		binary.Read(reader, binary.BigEndian, &cancPayload)
		payload = cancPayload
		break
	case Bitfield:
		bfPayload := BitfieldPayload{Bitfield: allocator(typeByte[0])}
		binary.Read(reader, binary.BigEndian, &bfPayload.Bitfield)
		payload = bfPayload
		break
	case Piece:
		piecePayload := PiecePayload{Piece: allocator(typeByte[0])}
		binary.Read(reader, binary.BigEndian, &piecePayload.Index)
		binary.Read(reader, binary.BigEndian, &piecePayload.Begin)
		binary.Read(reader, binary.BigEndian, &piecePayload.Piece)

		payload = piecePayload
		break
	}

	peerMsg := PeerMessage{Type: typeByte[0], Payload: payload}

	return &peerMsg, nil
}
