package db

type Peer struct {
	PeerId         int
	ProtocolPeerId []byte
	TorrentId      int
	IP             string
	Port           int
	Pieces         []Piece
}

type PeerRepository interface {
	Create(peer *Peer) error
	Update(peer *Peer) error
	GetByTorrentId(torrentId int) ([]Peer, error)
	GetByTorrentIdAndProtocolPeerId(torrentId int, protocolPeerId []byte) (*Peer, error)
	GetPeersForTorrentWithPieceIndex(torrentId int, pieceIndex int) ([]Peer, error)
}
