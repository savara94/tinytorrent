package sqlite

import (
	"database/sql"

	"example.com/db"
)

type PeerRepositorySQLite struct {
	SQLiteDB
}

func (r *PeerRepositorySQLite) Create(peer *db.Peer) error {
	stmt, err := r.db.Prepare(`
		INSERT INTO peer (protocol_peer_id, ip, port, torrent_id)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(peer.ProtocolPeerId, peer.IP, peer.Port, peer.TorrentId)
	if err != nil {
		return err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	peer.PeerId = int(lastInsertID)

	return nil
}

func (r *PeerRepositorySQLite) Update(peer *db.Peer) error {
	stmt, err := r.db.Prepare(`
		UPDATE peer
		SET protocol_peer_id=?, ip=?, port=?, torrent_id=?
		WHERE peer_id=?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(peer.ProtocolPeerId, peer.IP, peer.Port, peer.TorrentId, peer.PeerId)
	return err
}

func (r *PeerRepositorySQLite) GetByTorrentId(torrentId int) ([]db.Peer, error) {
	rows, err := r.db.Query("SELECT * FROM peer WHERE torrent_id=?", torrentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []db.Peer
	for rows.Next() {
		var peer db.Peer
		err := rows.Scan(&peer.PeerId, &peer.ProtocolPeerId, &peer.IP, &peer.Port, &peer.TorrentId)
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}

	return peers, nil
}

func (r *PeerRepositorySQLite) GetByTorrentIdAndProtocolPeerId(torrentId int, protocolPeerId []byte) (*db.Peer, error) {
	var peer db.Peer

	row := r.db.QueryRow(`
		SELECT peer_id, protocol_peer_id, ip, port, torrent_id
		FROM peer
		WHERE torrent_id = ? AND protocol_peer_id = ?;
	`, torrentId, protocolPeerId)

	err := row.Scan(&peer.PeerId, &peer.ProtocolPeerId, &peer.IP, &peer.Port, &peer.TorrentId)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &peer, nil
}

func (r *PeerRepositorySQLite) GetPeersForTorrentWithPieceIndex(torrentId int, pieceIndex int) ([]db.Peer, error) {
	query := `
		SELECT p.*
		FROM peers p
		JOIN pieces pc ON p.peer_id = pc.located_at_peer_id
		WHERE pc.torrent_id = ? AND pc.index = ?
	`

	rows, err := r.db.Query(query, torrentId, pieceIndex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []db.Peer
	for rows.Next() {
		peer := db.Peer{}
		err := rows.Scan(&peer.PeerId, &peer.ProtocolPeerId, &peer.TorrentId, &peer.IP, &peer.Port)
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}

	return peers, nil
}
