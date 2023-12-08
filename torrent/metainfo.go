package torrent

import (
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
}

var ErrLengthAndFilesNotSpecified = errors.New("either length or files must be specified")

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

	return &metaInfo, nil
}
