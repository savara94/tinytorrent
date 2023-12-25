package db

import "time"

type TrackerAnnounce struct {
	TrackerAnnounceId int
	TorrentId         int
	AnnounceTime      time.Time
	Announciation     string
	ScheduledTime     *time.Time
	Error             *string
	Done              bool
}

type TrackerAnnounceRepository interface {
	Create(announce *TrackerAnnounce) error
	Update(announce *TrackerAnnounce) error
	GetByTorrentId(torrentId int) ([]TrackerAnnounce, error)
	GetScheduledAfter(time time.Time) ([]TrackerAnnounce, error)
}
