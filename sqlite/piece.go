package sqlite

import "example.com/db"

type PieceRepositorySQLite struct {
	SQLiteDB
}

func (r *PieceRepositorySQLite) Create(piece *db.Piece) error {
	stmt, err := r.db.Prepare(`
		INSERT INTO piece (torrent_id, peer_id, is_downloaded, start, end, index, length, confirmed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(piece.TorrentId, piece.PeerId, piece.IsDownloaded, piece.Start, piece.End, piece.Index, piece.Length, piece.Confirmed)
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
		SET torrent_id=?, peer_id=?, is_downloaded=?, start=?, end=?, index=?, length=?, confirmed=?
		WHERE piece_id=?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(piece.TorrentId, piece.PeerId, piece.IsDownloaded, piece.Start, piece.End, piece.Index, piece.Length, piece.Confirmed, piece.PieceId)
	return err
}

func (r *PieceRepositorySQLite) GetByTorrentId(torrentId int) ([]db.Piece, error) {
	rows, err := r.db.Query("SELECT * FROM piece WHERE torrent_id=?", torrentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pieces []db.Piece
	for rows.Next() {
		var piece db.Piece
		err := rows.Scan(&piece.PieceId, &piece.TorrentId, &piece.PeerId, &piece.IsDownloaded, &piece.Start, &piece.End, &piece.Index, &piece.Length, &piece.Confirmed)
		if err != nil {
			return nil, err
		}
		pieces = append(pieces, piece)
	}

	return pieces, nil
}
