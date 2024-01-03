package torrent

import (
	"crypto/sha1"
	"errors"
	"io"

	"example.com/bencode"
)

type FileInfo struct {
	Length int      `bencode:"length"`
	Path   []string `bencode:"path"`
}

type GeneralInfo struct {
	Name        string      `bencode:"name"`
	PieceLength int         `bencode:"piece length"`
	Pieces      string      `bencode:"pieces"`
	Length      *int        `bencode:"length"`
	Files       *[]FileInfo `bencode:"files"`
	Private     *int        `bencode:"private"`
}

type MetaInfo struct {
	Announce     string      `bencode:"announce"`
	Comment      string      `bencode:"comment"`
	CreatedBy    string      `bencode:"created by"`
	CreationDate int         `bencode:"creation date"`
	Info         GeneralInfo `bencode:"info"`
	RawBytes     []byte

	infoHash []byte
}

var ErrLengthAndFilesNotSpecified = errors.New("either length or files must be specified")

func (metaInfo *MetaInfo) calculateInfoHash() error {
	var anyMap map[string]any

	err := bencode.Unmarshal(metaInfo.RawBytes, &anyMap)
	if err != nil {
		return err
	}

	// Use any map to marshal all stuff that is not part of BEP specification
	infoBencodedBytes, err := bencode.Marshal(anyMap["info"])
	if err != nil {
		return err
	}

	h := sha1.New()

	n, err := h.Write(infoBencodedBytes)

	if err != nil {
		return err
	}

	if n != len(infoBencodedBytes) {
		// panic here
	}

	metaInfo.infoHash = h.Sum(nil)

	return nil
}

func ParseMetaInfo(reader io.Reader) (*MetaInfo, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var metaInfo MetaInfo
	err = bencode.Unmarshal(bytes, &metaInfo)

	if err != nil {
		return nil, err
	}

	if metaInfo.Info.Length == nil && metaInfo.Info.Files == nil {
		return nil, ErrLengthAndFilesNotSpecified
	}

	metaInfo.RawBytes = bytes

	if err := metaInfo.calculateInfoHash(); err != nil {
		return nil, err
	}

	return &metaInfo, nil
}

func (metaInfo *MetaInfo) GetInfoHash() []byte {
	if metaInfo.infoHash == nil {
		metaInfo.calculateInfoHash()
	}

	return metaInfo.infoHash
}

func (metaInfo *MetaInfo) GetFullLength() (int, error) {
	if metaInfo.Info.Length != nil {
		return *metaInfo.Info.Length, nil
	}

	if metaInfo.Info.Files != nil {
		sum := 0
		for i := range *metaInfo.Info.Files {
			sum += (*metaInfo.Info.Files)[i].Length
		}

		return sum, nil
	}

	return 0, errors.New("Can't calculate full length!")
}
