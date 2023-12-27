package db

type Peer struct {
	PeerId         int
	ProtocolPeerId []byte
	IP             string
	Port           int
	TorrentId      int
	Reachable      bool
}

type PeerRepository interface {
	Create(peer *Peer) error
	Update(peer *Peer) error
	GetByTorrentId(torrentId int) ([]Peer, error)
	GetByTorrentIdAndProtocolPeerId(torrentId int, protocolPeerId []byte) (*Peer, error)
}
