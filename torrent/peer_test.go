package torrent

import (
	"bytes"
	"reflect"
	"testing"
)

func TestInitiatingHandshake(t *testing.T) {
	writeBuffer := bytes.NewBuffer([]byte{})

	seeder := Seeder{SeederInfo: PeerInfo{PeerId: GenerateRandomProtocolId()}, SeederWriter: writeBuffer, MetaInfo: &MetaInfo{infoHash: GenerateRandomProtocolId()}}

	err := seeder.InitiateHandshake()

	if err != nil {
		t.Errorf("Did not expect error %#v", err)
	}

	expectingSequence := []struct {
		name  string
		bytes []byte
	}{
		{"Protocol", []byte(HandshakeMsg)},
		{"Reserved", make([]byte, 8)},
		{"Infohash", seeder.MetaInfo.GetInfoHash()},
		{"PeerId", seeder.SeederInfo.PeerId},
	}

	for i := range expectingSequence {
		readingBytes := make([]byte, len(expectingSequence[i].bytes))

		n, err := writeBuffer.Read(readingBytes)
		if err != nil {
			t.Errorf("Did not expect err %#v", err)
		}

		if n < len(readingBytes) {
			t.Errorf("Could not read %d bytes", len(readingBytes))
		}

		if err == nil {
			if !reflect.DeepEqual(readingBytes, expectingSequence[i].bytes) {
				t.Errorf("Expected %#v, got %#v", expectingSequence[i].bytes, readingBytes)
			}
		}
	}
}

type peerMsgSendTestCase struct {
	name        string
	msg         PeerMessage
	sentBytes   []byte
	wantedError error
}

func TestSend(t *testing.T) {
	testCases := []peerMsgSendTestCase{
		{"Choke send", ChokeMessage, []byte{Choke}, nil},
		{"Unchoke send", UnchokeMessage, []byte{Unchoke}, nil},
		{"Interested send", InterestedMessage, []byte{Interested}, nil},
		{"Not interested send", NotInterestedMessage, []byte{NotInterested}, nil},
		{"Keepalive send", KeepAliveMessage, []byte{}, nil},
		{"Bitfield send", PeerMessage{Type: Bitfield, Payload: BitfieldPayload{Bitfield: []byte{1, 2, 3, 4}}}, []byte{5, 1, 2, 3, 4}, nil},
		{"Request send", PeerMessage{Type: Request, Payload: RequestPayload{Index: 0, Begin: 1, Length: 5}}, []byte{6, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 5}, nil},
		{"Piece send", PeerMessage{Type: Piece, Payload: PiecePayload{Index: 0, Begin: 0, Piece: []byte{1, 2, 3, 4, 5}}}, []byte{7, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5}, nil},
		{"Cancel send", PeerMessage{Type: Cancel, Payload: CancelPayload{Index: 0, Begin: 1, Length: 5}}, []byte{8, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 5}, nil},
	}

	for i := range testCases {
		buffer := bytes.NewBuffer([]byte{})

		err := Send(buffer, &testCases[i].msg)

		if err == nil {
			if !reflect.DeepEqual(testCases[i].sentBytes, buffer.Bytes()) {
				t.Errorf("%s wanted to send %#v, but sent %#v", testCases[i].name, testCases[i].sentBytes, buffer.Bytes())
			}
		}

		if err != nil {
			if err != testCases[i].wantedError {
				t.Errorf("%s wanted error %#v, but got %#v", testCases[i].name, testCases[i].wantedError, err)
			}
		}
	}
}

type peerMsgReceiveTestCase struct {
	name           string
	bytesToReceive []byte
	peerMsg        PeerMessage
	wantedError    error
}

func TestReceive(t *testing.T) {
	testCases := []peerMsgReceiveTestCase{
		{"Choke recv", []byte{Choke}, ChokeMessage, nil},
		{"Unchoke recv", []byte{Unchoke}, UnchokeMessage, nil},
		{"Interested recv", []byte{Interested}, InterestedMessage, nil},
		{"Not interested recv", []byte{NotInterested}, NotInterestedMessage, nil},
		// {"Keepalive recv", []byte{}, KeepAliveMessage, nil},
		{"Bitfield recv", []byte{5, 1, 2, 3, 4}, PeerMessage{Type: Bitfield, Payload: BitfieldPayload{Bitfield: []byte{1, 2, 3, 4}}}, nil},
		{"Request recv", []byte{6, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 5}, PeerMessage{Type: Request, Payload: RequestPayload{Index: 0, Begin: 1, Length: 5}}, nil},
		{"Piece recv", []byte{7, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 3, 4, 5}, PeerMessage{Type: Piece, Payload: PiecePayload{Index: 0, Begin: 1, Piece: []byte{1, 2, 3, 4, 5}}}, nil},
		{"Cancel recv", []byte{8, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 5}, PeerMessage{Type: Cancel, Payload: CancelPayload{Index: 0, Begin: 1, Length: 5}}, nil},
	}

	for i := range testCases {
		writer := bytes.NewBuffer([]byte{})

		writer.Write(testCases[i].bytesToReceive)

		msg, err := Receive(writer, func(msgType byte) []byte {
			switch msgType {
			case Bitfield:
				return make([]byte, 4)
			case Piece:
				return make([]byte, 5)
			default:
				return nil
			}
		})

		if err == nil {
			if !reflect.DeepEqual(testCases[i].peerMsg, *msg) {
				t.Errorf("%s wanted to recv %#v, but recv %#v", testCases[i].name, testCases[i].peerMsg, msg)
			}
		}

		if err != nil {
			if err != testCases[i].wantedError {
				t.Errorf("%s wanted error %#v, but got %#v", testCases[i].name, testCases[i].wantedError, err)
			}
		}
	}
}
