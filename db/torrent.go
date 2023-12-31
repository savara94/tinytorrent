package db

import "time"

type Torrent struct {
	TorrentId   int
	Name        string
	Announce    string
	HashInfo    []byte
	Size        int
	CreatedTime time.Time
	Paused      bool
	Location    string
	Progress    int
	Announces   []TrackerAnnounce
	Pieces      []Piece
	RawMetaInfo []byte
}

type TorrentRepository interface {
	Create(torrent *Torrent) error
	Update(torrent *Torrent) error
	Delete(torrent *Torrent) error
	GetAll() ([]Torrent, error)
	GetByHashInfo(hashInfo []byte) (*Torrent, error)
}
