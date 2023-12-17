package torrent

import (
	"encoding/hex"
	"os"
	"testing"
)

type metaInfoTestCase struct {
	announce string
	name     string
	infoHash string
}

type testCase struct {
	name      string
	filePath  string
	want      metaInfoTestCase
	wantedErr error
}

func assertMetaInfo(testCase *testCase, t *testing.T) error {
	name := testCase.name
	filePath := testCase.filePath

	reader, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer reader.Close()

	info, err := ParseMetaInfo(reader)
	if err != nil {
		return err
	}

	metaInfoGot := metaInfoTestCase{name: info.Info.Name, announce: info.Announce, infoHash: infoHashToHexString(info)}

	if testCase.want != metaInfoGot {
		t.Errorf("%s got %#v wanted %#v", name, metaInfoGot, testCase.want)
	}

	return nil
}

func assertError(err error, testCase *testCase, t *testing.T) {
	name := testCase.name
	wantedErr := testCase.wantedErr

	if wantedErr == nil && err != nil {
		t.Errorf("%s Got error %v, wasn't expecting one", name, err)
	}

	if wantedErr != nil && err == nil {
		t.Errorf("%s wanted %v got nil", name, wantedErr)
	}

	if wantedErr != err {
		t.Errorf("%s wanted %v got %v", name, wantedErr, err)
	}
}

func infoHashToHexString(metaInfo *MetaInfo) string {
	infoHash := metaInfo.GetInfoHash()
	return hex.EncodeToString(infoHash)
}

func TestParseMetaInfo(t *testing.T) {
	testCases := []testCase{
		{"hello world", "examples/hello_world.torrent", metaInfoTestCase{"http://localhost:6969/announce", "hello_world", "2af633c618e64c9ea3972789f5764fbea6f42d40"}, nil},
		{"lovecraft", "examples/lovecraft.torrent", metaInfoTestCase{"http://bt1.archive.org:6969/announce", "sevenhplovecraftstories_pc_librivox", "588fb9c2cf9f7ceb973976c1e0eaaf38c3444999"}, nil},
		{"missing length and files", "examples/missing_length_and_files.torrent", metaInfoTestCase{}, ErrLengthAndFilesNotSpecified},
	}

	for i := range testCases {
		err := assertMetaInfo(&testCases[i], t)
		assertError(err, &testCases[i], t)
	}
}
