package torrent

import (
	"os"
	"testing"
)

type metaInfoWant struct {
	announce string
	name     string
}

type testCaseMetaInfo struct {
	filePath  string
	want      *metaInfoWant
	wantedErr error
}

func TestParseMetaInfo(t *testing.T) {
	testCases := []testCaseMetaInfo{
		{"examples/hello_world.torrent", &metaInfoWant{"http://localhost:6969/announce", "hello_world"}, nil},
		{"examples/lovecraft.torrent", &metaInfoWant{"http://localhost:6969/announce", "sevenhplovecraftstories_pc_librivox"}, nil},
		{"examples/missing_length_and_files.torrent", nil, ErrLengthAndFilesNotSpecified},
	}

	for i := range testCases {
		testCase := testCases[i]
		filePath := testCase.filePath
		wantedErr := testCase.wantedErr
		want := testCase.want

		reader, err := os.Open(filePath)

		if err != nil {
			t.Errorf("%d Failed opening %s", i, filePath)
		}

		gotInfo, gottenErr := ParseMetaInfo(reader)

		if wantedErr == nil && gottenErr != nil {
			t.Errorf("Got error %v, wasn't expecting one", gottenErr)
		}

		if wantedErr != nil && gottenErr == nil {
			t.Errorf("wanted %v got nil", wantedErr)
		}

		if wantedErr != nil && gottenErr != nil {
			if wantedErr != gottenErr {
				t.Errorf("wanted %v got %v", wantedErr, gottenErr)
			}
		}

		if want != nil && gottenErr == nil {
			if want.name != gotInfo.Info.Name {
				t.Errorf("wanted %s got %s", want.name, gotInfo.Info.Name)
			}
		}
	}
}
