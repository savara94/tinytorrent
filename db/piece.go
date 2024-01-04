package db

import "time"

type Piece struct {
	PieceId         int
	TorrentId       int
	LocatedAtPeerId int

	CameFromPeerId *int
	Start          *time.Time
	End            *time.Time
	Index          int
	Length         int
}

type PieceFilter struct {
	PieceId         *int
	TorrentId       *int
	LocatedAtPeerId *int
	Index           *int
}

func (pf *PieceFilter) WithPieceId(id int) *PieceFilter {
	pf.PieceId = &id

	return pf
}

func (pf *PieceFilter) ThatBelongsToTorrent(id int) *PieceFilter {
	pf.TorrentId = &id

	return pf
}

func (pf *PieceFilter) LocatedAtPeer(id int) *PieceFilter {
	pf.LocatedAtPeerId = &id

	return pf
}

func (pf *PieceFilter) WithIndex(index int) *PieceFilter {
	pf.Index = &index

	return pf
}

type PieceRepository interface {
	Create(piece *Piece) error
	Update(piece *Piece) error
	GetByFilterAndOrder(filter *PieceFilter, orderBy string, limit int) ([]Piece, error)
}
