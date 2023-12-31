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
		INSERT INTO peer (protocol_peer_id, ip, port, torrent_id, reachable)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(peer.ProtocolPeerId, peer.IP, peer.Port, peer.TorrentId, peer.Reachable)
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
		SET protocol_peer_id=?, ip=?, port=?, torrent_id=?, reachable=?
		WHERE peer_id=?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(peer.ProtocolPeerId, peer.IP, peer.Port, peer.TorrentId, peer.Reachable, peer.PeerId)
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
		err := rows.Scan(&peer.PeerId, &peer.ProtocolPeerId, &peer.IP, &peer.Port, &peer.TorrentId, &peer.Reachable)
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
		SELECT peer_id, protocol_peer_id, ip, port, torrent_id, reachable
		FROM peer
		WHERE torrent_id = ? AND protocol_peer_id = ?;
	`, torrentId, protocolPeerId)

	err := row.Scan(&peer.PeerId, &peer.ProtocolPeerId, &peer.IP, &peer.Port, &peer.TorrentId, &peer.Reachable)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &peer, nil
}
