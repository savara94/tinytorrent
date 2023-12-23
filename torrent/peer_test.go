package torrent

import (
	"bytes"
	"reflect"
	"testing"
)

func TestInitiatingHandshake(t *testing.T) {

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

func TestReceive(t *testing.T) {

}
