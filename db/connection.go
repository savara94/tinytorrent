package db

import "time"

type Connection struct {
	TorrentId          int
	RemotePeerId       int
	ImChoked           bool
	RemoteIsChoked     bool
	ImInterested       bool
	RemoteIsInterested bool
	DownloadRate       float32
	UploadRate         float32
	LastActivity       time.Time
}

type ConnectionFilter struct {
	TorrentId          *int
	RemotePeerId       *int
	ImChoked           *bool
	RemoteIsChoked     *bool
	ImInterested       *bool
	RemoteIsInterested *bool
}

func (cf *ConnectionFilter) RelatedToTorrent(id int) *ConnectionFilter {
	cf.TorrentId = &id

	return cf
}

func (cf *ConnectionFilter) RemotePeerHasId(id int) *ConnectionFilter {
	cf.RemotePeerId = &id

	return cf
}

func (cf *ConnectionFilter) WhereImChoked(yes bool) *ConnectionFilter {
	cf.ImChoked = &yes

	return cf
}

func (cf *ConnectionFilter) WhereRemoteChoked(yes bool) *ConnectionFilter {
	cf.RemoteIsChoked = &yes

	return cf
}

func (cf *ConnectionFilter) WhereImInterested(yes bool) *ConnectionFilter {
	cf.ImInterested = &yes

	return cf
}

func (cf *ConnectionFilter) WhereRemoteIsInterested(yes bool) *ConnectionFilter {
	cf.RemoteIsInterested = &yes

	return cf
}

type ConnectionRepository interface {
	Upsert(connection *Connection) error
	GetByFilterAndOrder(filter *ConnectionFilter, orderBy string, descending bool, limit int) ([]Connection, error)
}
