package sqlite

import (
	"fmt"

	"example.com/db"
)

type PieceRepositorySQLite struct {
	SQLiteDB
}

func (r *PieceRepositorySQLite) Create(piece *db.Piece) error {
	stmt, err := r.db.Prepare(`
		INSERT INTO piece (torrent_id, located_at_peer_id, came_from_peer_id, start, end, index, length)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		piece.TorrentId, piece.LocatedAtPeerId, piece.CameFromPeerId, piece.Start, piece.End, piece.Index, piece.Length,
	)
	if err != nil {
		return err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	piece.PieceId = int(lastInsertID)

	return nil
}

func (r *PieceRepositorySQLite) Update(piece *db.Piece) error {
	stmt, err := r.db.Prepare(`
		UPDATE piece
		SET torrent_id=?, located_at_peer_id=?, came_from_peer_id=?, start=?, end=?, index=?, length=?
		WHERE piece_id=?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		piece.TorrentId, piece.LocatedAtPeerId, piece.CameFromPeerId, piece.Start, piece.End, piece.Index, piece.Length, piece.PieceId,
	)
	return err
}

func (r *PieceRepositorySQLite) GetByFilterAndOrder(filter *db.PieceFilter, orderBy string, limit int) ([]db.Piece, error) {
	query := "SELECT * FROM pieces WHERE 1=1"

	if filter != nil {
		if filter.PieceId != nil {
			query += fmt.Sprintf(" AND piece_id=%d", *filter.PieceId)
		}
		if filter.TorrentId != nil {
			query += fmt.Sprintf(" AND torrent_id=%d", *filter.TorrentId)
		}
		if filter.LocatedAtPeerId != nil {
			query += fmt.Sprintf(" AND located_at_peer_id=%d", *filter.LocatedAtPeerId)
		}
		if filter.Index != nil {
			query += fmt.Sprintf(" AND index=%d", *filter.Index)
		}
		// Add other filter conditions as needed
	}

	if orderBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", orderBy)
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pieces []db.Piece
	for rows.Next() {
		piece := db.Piece{}
		err := rows.Scan(
			&piece.PieceId, &piece.TorrentId, &piece.LocatedAtPeerId, &piece.CameFromPeerId, &piece.Start, &piece.End, &piece.Index, &piece.Length,
		)

		if err != nil {
			return nil, err
		}

		pieces = append(pieces, piece)
	}

	return pieces, nil
}
