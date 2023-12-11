package torrent

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
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
	Pieces      []byte      `json:"pieces"`
	Length      *int        `json:"length,omitempty"`
	Files       *[]FileInfo `json:"files,omitempty"`
	Private     *int        `json:"private,omitempty"`
}

type MetaInfo struct {
	Announce     string      `json:"announce"`
	Comment      string      `json:"comment"`
	CreatedBy    string      `json:"created by"`
	CreationDate int         `json:"creation date"`
	Encoding     string      `json:"encoding"`
	Info         GeneralInfo `json:"info"`
	infoHash     []byte
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

func (metaInfo *MetaInfo) Encode() ([]byte, error) {
	byteArray, err := json.Marshal(metaInfo)

	if err != nil {
		return nil, err
	}

	metaInfoMap := make(map[string]any)

	err = json.Unmarshal(byteArray, &metaInfoMap)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%v", metaInfoMap)

	bencoded, err := bencode.Encode(metaInfoMap)
	if err != nil {
		return nil, err
	}

	return []byte(bencoded), nil
}

func (metaInfo *MetaInfo) GetInfoHash() []byte {
	return metaInfo.infoHash
}
