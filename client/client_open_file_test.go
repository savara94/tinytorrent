package client

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"
)

var helloWorldTorrentPath = "../torrent/examples/hello_world.torrent"

func testOpenTorrentWrongFile(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Test
	testReader := strings.NewReader("this is not .torrent")
	dbTorrent, err := client.OpenTorrent(testReader, "./")
	if err == nil {
		// TODO
		// Check type as well
		t.Errorf("Expected error here!")
		return
	}

	if dbTorrent != nil {
		t.Errorf("Expected nil here!")
		return
	}
}

func testDirectoryAlreadyTaken(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Setup
	fileReader, err := os.Open(helloWorldTorrentPath)
	if err != nil {
		t.Errorf("Could not open test file %v", err)
		return
	}
	defer fileReader.Close()

	// This is contained inside file
	torrentName := "hello_world"
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("Could not create temp dir %v", err)
		return
	}

	downloadPath := path.Join(tmpDir, torrentName)

	// Test
	dbTorrent, err := client.OpenTorrent(fileReader, downloadPath)
	if err == nil {
		// TODO
		// Check for type here.
		t.Errorf("Expected error here")
		return
	}

	if dbTorrent != nil {
		t.Errorf("Expected nil here")
		return
	}
}

func testOpenFewTimes(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Setup
	fileReader, err := os.Open(helloWorldTorrentPath)
	if err != nil {
		t.Errorf("Could not open test file %v", err)
		return
	}
	defer fileReader.Close()

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("Could not create temp dir %v", err)
		return
	}

	// Test
	dbTorrent, err := client.OpenTorrent(fileReader, tmpDir)
	if err != nil {
		// TODO
		// Check for type here.
		t.Errorf("Did not expect error here here %v", err)
		return
	}

	if dbTorrent == nil {
		t.Errorf("Did not expect nil here")
		return
	}

	if dbTorrent.TorrentId == 0 {
		t.Errorf("TorrentId not updated %#v", dbTorrent)
		return
	}

	if dbTorrent.Name != "hello_world" {
		t.Errorf("Name not set")
		return
	}

	if dbTorrent.Size == 0 {
		t.Errorf("Size not set")
		return
	}

	t.Run("Open one more time", func(t *testing.T) {
		_, err := fileReader.Seek(0, io.SeekStart)
		if err != nil {
			t.Errorf("Could not rewind file %v", err)
			return
		}

		newTmpDir, _ := os.MkdirTemp("", "")
		if err != nil {
			t.Errorf("Could not create temp dir %v", err)
			return
		}

		existingTorrent, err := client.OpenTorrent(fileReader, newTmpDir)
		if err == nil {
			t.Errorf("Expected error here!")
			return
		}

		if err != nil {
			// TODO
			// Check for error type here
		}

		if existingTorrent == nil {
			t.Errorf("Did not expect nil here")
			return
		}

		if existingTorrent.Location != tmpDir {
			t.Errorf("Expected old location %#v", existingTorrent)
			return
		}
	})
}

func TestOpenTorrent(t *testing.T) {
	testCases := []testCase{
		{
			name:         "Wrong file",
			dbSchemaPath: schemaPath,
			testFunction: testOpenTorrentWrongFile,
		},
		{
			name:         "Directory already taken",
			dbSchemaPath: schemaPath,
			testFunction: testDirectoryAlreadyTaken,
		},
		{
			name:         "Open valid few times",
			dbSchemaPath: schemaPath,
			testFunction: testOpenFewTimes,
		},
	}

	for i := range testCases {
		testCase := testCases[i]

		t.Run(testCase.name, func(t *testing.T) {
			runTestCase(&testCase, t)
		})
	}
}
