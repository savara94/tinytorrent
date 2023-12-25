package db

import "time"

type Piece struct {
	PieceId      int
	TorrentId    int
	PeerId       int
	IsDownloaded bool
	Start        time.Time
	End          *time.Time
	Index        int
	Length       int
	Confirmed    *bool
}

type PieceRepository interface {
	Create(piece *Piece) error
	Update(piece *Piece) error
	GetByTorrentId(torrentId int) ([]Piece, error)
}
