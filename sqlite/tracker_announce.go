package sqlite

import (
	"time"

	"example.com/db"
)

type TrackerAnnounceRepositorySQLite struct {
	SQLiteDB
}

func (r *TrackerAnnounceRepositorySQLite) Create(announce *db.TrackerAnnounce) error {
	stmt, err := r.db.Prepare(`
		INSERT INTO tracker_announce (torrent_id, announce_time, announciation, scheduled_time, err, done, raw_response)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(announce.TorrentId, announce.AnnounceTime, announce.Announciation, announce.ScheduledTime, announce.Error, announce.Done, announce.RawResponse)
	if err != nil {
		return err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	announce.TrackerAnnounceId = int(lastInsertID)

	return nil
}

func (r *TrackerAnnounceRepositorySQLite) Update(announce *db.TrackerAnnounce) error {
	stmt, err := r.db.Prepare(`
		UPDATE tracker_announce
		SET torrent_id=?, announce_time=?, announciation=?, scheduled_time=?, err=?, done=?, raw_response=?
		WHERE tracker_announce_id=?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(announce.TorrentId, announce.AnnounceTime, announce.Announciation, announce.ScheduledTime, announce.Error, announce.Done, announce.TrackerAnnounceId, announce.RawResponse)
	return err
}

func (r *TrackerAnnounceRepositorySQLite) GetByTorrentId(torrentId int) ([]db.TrackerAnnounce, error) {
	rows, err := r.db.Query("SELECT * FROM tracker_announce WHERE torrent_id=?", torrentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announces []db.TrackerAnnounce
	for rows.Next() {
		var announce db.TrackerAnnounce
		err := rows.Scan(&announce.TrackerAnnounceId, &announce.TorrentId, &announce.AnnounceTime, &announce.Announciation, &announce.ScheduledTime, &announce.Error, &announce.Done, &announce.RawResponse)
		if err != nil {
			return nil, err
		}
		announces = append(announces, announce)
	}

	return announces, nil
}

func (r *TrackerAnnounceRepositorySQLite) GetScheduledAfter(scheduledTime time.Time) ([]db.TrackerAnnounce, error) {
	rows, err := r.db.Query("SELECT * FROM tracker_announce WHERE scheduled_time > ?", scheduledTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announces []db.TrackerAnnounce
	for rows.Next() {
		var announce db.TrackerAnnounce
		err := rows.Scan(&announce.TrackerAnnounceId, &announce.TorrentId, &announce.AnnounceTime, &announce.Announciation, &announce.ScheduledTime, &announce.Error, &announce.Done, &announce.RawResponse)
		if err != nil {
			return nil, err
		}
		announces = append(announces, announce)
	}

	return announces, nil
}
