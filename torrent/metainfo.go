package torrent

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"io"

	"example.com/bencode"
)

type FileInfo struct {
	Length int      `json:"length"`
	Path   []string `json:"path"`
}

type GeneralInfo struct {
	Name        string      `json:"name"`
	PieceLength int         `json:"piece length"`
	Pieces      string      `json:"pieces"`
	Length      *int        `json:"length"`
	Files       *[]FileInfo `json:"files"`
}

type MetaInfo struct {
	Announce string      `json:"announce"`
	Info     GeneralInfo `json:"info"`
	infoHash []byte
}

var ErrLengthAndFilesNotSpecified = errors.New("either length or files must be specified")

func (metaInfo *MetaInfo) calculateInfoHash() error {
	infoByteArray, err := json.Marshal(metaInfo.Info)
	if err != nil {
		return err
	}

	infoMap := make(map[string]any)
	err = json.Unmarshal(infoByteArray, &infoMap)
	if err != nil {
		return err
	}

	bencoded, err := bencode.Encode(infoMap)
	if err != nil {
		return err
	}

	h := sha1.New()

	n, err := io.WriteString(h, bencoded)

	if err != nil || n != len(bencoded) {
		return err
	}

	metaInfo.infoHash = h.Sum(nil)

	return nil
}

func ParseMetaInfo(reader io.Reader) (*MetaInfo, error) {
	bencode, err := bencode.Decode(reader)

	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(bencode)

	if err != nil {
		return nil, err
	}

	metaInfo := MetaInfo{}
	err = json.Unmarshal(bytes, &metaInfo)

	if err != nil {
		return nil, err
	}

	if metaInfo.Info.Length == nil && metaInfo.Info.Files == nil {
		return nil, ErrLengthAndFilesNotSpecified
	}

	if err := metaInfo.calculateInfoHash(); err != nil {
		return nil, err
	}

	return &metaInfo, nil
}

func (metaInfo *MetaInfo) GetInfoHash() []byte {
	return metaInfo.infoHash
}
