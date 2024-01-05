package torrent

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"reflect"
)

const HandshakeMsg = "\x13" + "BitTorrent protocol"

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

type ConnectionChecker func() (bool, error)
type ConnectionTerminator func() error

type PeerConnection struct {
	PeerId              []byte
	RemotePeerId        []byte
	PeerWriter          io.Writer
	PeerReader          io.Reader
	MetaInfo            *MetaInfo
	IsConnectionSevered ConnectionChecker
	SevereConnection    ConnectionTerminator
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateRandomProtocolId() []byte {
	b := make([]byte, 20)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return b
}

func (pc *PeerConnection) InitiateHandshake() error {
	var handshakeBytesWithoutPeerId []byte

	handshakeBytesWithoutPeerId = append(handshakeBytesWithoutPeerId, []byte(HandshakeMsg)...)
	handshakeBytesWithoutPeerId = append(handshakeBytesWithoutPeerId, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	handshakeBytesWithoutPeerId = append(handshakeBytesWithoutPeerId, pc.MetaInfo.GetInfoHash()...)

	myHandshakeBytes := append(handshakeBytesWithoutPeerId, pc.PeerId...)
	remoteHandshakeBytes := append(handshakeBytesWithoutPeerId, pc.RemotePeerId...)

	n, err := pc.PeerWriter.Write(myHandshakeBytes)
	if err != nil {
		return err
	}

	if n < len(myHandshakeBytes) {
		return errors.New("Couldn't send all my handshake bytes.")
	}

	severed, err := pc.IsConnectionSevered()
	if err != nil {
		return err
	}

	if severed {
		errMsg := fmt.Sprintf("Connection is severed by remote peer %s", hex.EncodeToString(pc.RemotePeerId))
		return errors.New(errMsg)
	}

	readBytes := make([]byte, len(remoteHandshakeBytes))

	n, err = pc.PeerReader.Read(readBytes)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(readBytes, remoteHandshakeBytes) {
		err = pc.SevereConnection()
		if err != nil {
			errMsg := fmt.Sprintf("Could not severe connection with %s %v", hex.EncodeToString(pc.RemotePeerId), err)
			slog.Error(errMsg)
		}

		return errors.New("Did not receive expected handshake bytes.")
	}

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
