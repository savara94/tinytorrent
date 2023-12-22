package torrent

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
