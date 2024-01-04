package sqlite

import (
	"fmt"

	"example.com/db"
)

type ConnectionRepositorySQLite struct {
	SQLiteDB
}

func (repo *ConnectionRepositorySQLite) Upsert(connection *db.Connection) error {
	query := `
		INSERT OR REPLACE INTO connections (
			torrent_id, remote_peer_id, im_choked, remote_is_choked,
			im_interested, remote_is_interested, download_rate, upload_rate,
			last_activity
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := repo.db.Exec(query,
		connection.TorrentId,
		connection.RemotePeerId,
		connection.ImChoked,
		connection.RemoteIsChoked,
		connection.ImInterested,
		connection.RemoteIsInterested,
		connection.DownloadRate,
		connection.UploadRate,
		connection.LastActivity,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert connection: %v", err)
	}

	return nil
}

func (r *ConnectionRepositorySQLite) GetByFilterAndOrder(filter *db.ConnectionFilter, orderBy string, descending bool, limit int) ([]db.Connection, error) {
	query := "SELECT * FROM connections WHERE 1=1"

	if filter != nil {
		if filter.TorrentId != nil {
			query += fmt.Sprintf(" AND torrent_id=%d", *filter.TorrentId)
		}
		if filter.RemotePeerId != nil {
			query += fmt.Sprintf(" AND remote_peer_id=%d", *filter.RemotePeerId)
		}
		if filter.ImChoked != nil {
			query += fmt.Sprintf(" AND im_choked=%d", boolToInt(*filter.ImChoked))
		}
		if filter.RemoteIsChoked != nil {
			query += fmt.Sprintf(" AND remote_is_choked=%d", boolToInt(*filter.RemoteIsChoked))
		}
		if filter.ImInterested != nil {
			query += fmt.Sprintf(" AND im_interested=%d", boolToInt(*filter.ImInterested))
		}
		if filter.RemoteIsInterested != nil {
			query += fmt.Sprintf(" AND remote_is_interested=%d", boolToInt(*filter.RemoteIsInterested))
		}
		// Add other filter conditions as needed
	}

	if orderBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", orderBy)
		if descending {
			query += " DESC"
		}
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []db.Connection
	for rows.Next() {
		connection := db.Connection{}
		err := rows.Scan(
			&connection.TorrentId, &connection.RemotePeerId, &connection.ImChoked, &connection.RemoteIsChoked,
			&connection.ImInterested, &connection.RemoteIsInterested, &connection.DownloadRate, &connection.UploadRate, &connection.LastActivity,
		)

		if err != nil {
			return nil, err
		}

		connections = append(connections, connection)
	}

	return connections, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
