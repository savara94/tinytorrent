package torrent

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

type handshakeTestCase struct {
	name                 string
	peerId               []byte
	remotePeerId         []byte
	infoHash             []byte
	readBuffer           io.Reader
	connectionChecker    ConnectionChecker
	expectedError        bool
	expectedErrorMsgPart string
}

func runInitiatingHandshakeTests(testCase *handshakeTestCase, t *testing.T) {
	// Setup
	buffer := bytes.NewBuffer([]byte{})

	peerConnection := PeerConnection{
		PeerId:              testCase.peerId,
		RemotePeerId:        testCase.remotePeerId,
		InfoHash:            testCase.infoHash,
		PeerWriter:          buffer,
		PeerReader:          testCase.readBuffer,
		IsConnectionSevered: testCase.connectionChecker,
	}

	// Test
	err := peerConnection.InitiateHandshake()

	// Test errors
	if testCase.expectedError && err == nil {
		t.Errorf("Expected error here.")
		return
	}

	if testCase.expectedError && err != nil {
		if !strings.Contains(err.Error(), testCase.expectedErrorMsgPart) {
			t.Errorf("Expected part of message %s in %v", testCase.expectedErrorMsgPart, err)
			return
		}
	}

	if !testCase.expectedError && err != nil {
		t.Errorf("Did not expect error here. %v", err)
		return
	}

	// Test what's been wrote to buffer
	gotHandshake := buffer.Bytes()
	expectedHandshake := createHandshakeBytes(peerConnection.PeerId, testCase.infoHash)

	if !reflect.DeepEqual(expectedHandshake, gotHandshake) {
		t.Errorf("Did not send what was expected %v | %v", expectedHandshake, gotHandshake)
		return
	}
}

func createHandshakeBytes(peerId []byte, infoHash []byte) []byte {
	handshakeBytes := append([]byte(HandshakeMsg), []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	handshakeBytes = append(handshakeBytes, infoHash...)
	handshakeBytes = append(handshakeBytes, peerId...)

	return handshakeBytes
}

func TestInitiatingHandshake(t *testing.T) {
	peerId := GenerateRandomProtocolId()
	remotePeerId := GenerateRandomProtocolId()
	infoHash := GenerateRandomProtocolId()

	testCases := []handshakeTestCase{
		{
			name:                 "Connection severed by remote",
			peerId:               peerId,
			remotePeerId:         remotePeerId,
			infoHash:             infoHash,
			readBuffer:           nil,
			connectionChecker:    func() (bool, error) { return true, nil },
			expectedError:        true,
			expectedErrorMsgPart: "severed",
		},
		{
			name:                 "Connection check error",
			peerId:               peerId,
			remotePeerId:         remotePeerId,
			infoHash:             infoHash,
			readBuffer:           nil,
			connectionChecker:    func() (bool, error) { return true, errors.New("random message") },
			expectedError:        true,
			expectedErrorMsgPart: "random message",
		},
		{
			name:                 "Wrong handshake from remote side",
			peerId:               peerId,
			remotePeerId:         remotePeerId,
			infoHash:             infoHash,
			readBuffer:           bytes.NewBuffer([]byte("This is not OK.")),
			connectionChecker:    func() (bool, error) { return false, nil },
			expectedError:        true,
			expectedErrorMsgPart: "handshake",
		},
		{
			name:              "OK",
			peerId:            peerId,
			remotePeerId:      remotePeerId,
			infoHash:          infoHash,
			readBuffer:        bytes.NewBuffer(createHandshakeBytes(remotePeerId, infoHash)),
			connectionChecker: func() (bool, error) { return false, nil },
			expectedError:     false,
		},
	}

	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			runInitiatingHandshakeTests(&testCases[i], t)
		})
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
