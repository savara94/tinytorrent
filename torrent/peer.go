package torrent

type Peer struct {
	PeerId       []byte
	Port         int
	MetaInfoList []MetaInfo
}
