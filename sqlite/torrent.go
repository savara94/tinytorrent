package sqlite

import "example.com/db"

type TorrentRepositorySQLite struct {
	SQLiteDB
}

func (r *TorrentRepositorySQLite) Create(torrent *db.Torrent) error {
	stmt, err := r.db.Prepare(`
		INSERT INTO torrent (hash_info, created_time, paused, location, progress, raw_meta_info)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(torrent.HashInfo, torrent.CreatedTime, torrent.Paused, torrent.Location, torrent.Progress, torrent.RawMetaInfo)
	if err != nil {
		return err
	}

	// Retrieve the last inserted ID and update the Torrent struct
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	torrent.TorrentId = int(lastInsertID)

	return nil
}

func (r *TorrentRepositorySQLite) Update(torrent *db.Torrent) error {
	stmt, err := r.db.Prepare(`
		UPDATE torrent
		SET hash_info=?, created_time=?, paused=?, location=?, progress=?, raw_meta_info=?
		WHERE torrent_id=?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(torrent.HashInfo, torrent.CreatedTime, torrent.Paused, torrent.Location, torrent.Progress, torrent.RawMetaInfo, torrent.TorrentId)
	return err
}

func (r *TorrentRepositorySQLite) Delete(torrent *db.Torrent) error {
	_, err := r.db.Exec("DELETE FROM torrent WHERE torrent_id=?", torrent.TorrentId)
	return err
}

func (r *TorrentRepositorySQLite) GetAll() ([]db.Torrent, error) {
	rows, err := r.db.Query("SELECT * FROM torrent")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var torrents []db.Torrent
	for rows.Next() {
		var torrent db.Torrent
		err := rows.Scan(&torrent.TorrentId, &torrent.HashInfo, &torrent.CreatedTime, &torrent.Paused, &torrent.Location, &torrent.Progress, &torrent.RawMetaInfo)
		if err != nil {
			return nil, err
		}
		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

func (r *TorrentRepositorySQLite) GetByHashInfo(hashInfo []byte) (*db.Torrent, error) {
	var torrent db.Torrent
	err := r.db.QueryRow("SELECT * FROM torrent WHERE hash_info=?", hashInfo).Scan(
		&torrent.TorrentId, &torrent.HashInfo, &torrent.CreatedTime, &torrent.Paused, &torrent.Location, &torrent.Progress, &torrent.RawMetaInfo,
	)
	if err != nil {
		return nil, err
	}
	return &torrent, nil
}
