package client

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"example.com/bencode"
	"example.com/db"
	"example.com/torrent"
)

var capturedTrackerResponsePath = "../torrent/examples/ubuntu-22.04.3-desktop-amd64.iso.torrent.compact0.announce"

func setupExistingTorrent(client *Client, dependencies *testCaseDependencies, t *testing.T) (*db.Torrent, error) {
	metaInfo := torrent.MetaInfo{
		Announce: dependencies.trackerServer.URL + "/announce",
		Info: torrent.GeneralInfo{
			Name:  "fake",
			Files: &[]torrent.FileInfo{},
		},
	}

	metaInfoBytes, err := bencode.Marshal(metaInfo)
	metaInfoBuffer := bytes.NewBuffer(metaInfoBytes)

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("Could not create test directory %v", tmpDir)
		return nil, err
	}

	err = client.Initialize()
	if err != nil {
		t.Errorf("Could not initialize client %v", err)
		return nil, err
	}

	dbTorrent, err := client.OpenTorrent(metaInfoBuffer, tmpDir)
	if err != nil {
		t.Errorf("Could not open torrent %v", err)
		return nil, err
	}

	return dbTorrent, nil
}

func testAnnounce5xxError(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Setup
	dependencies.trackerServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	dbTorrent, err := setupExistingTorrent(client, dependencies, t)
	if err != nil {
		t.Errorf("Error on test case setup %v", err)
		return
	}

	// Test
	dbAnnounce, err := client.Announce(dbTorrent)
	if err != nil {
		t.Errorf("Not expecting error error here.")
		return
	}

	if dbAnnounce == nil {
		t.Errorf("Expected data to be created. %v", err)
		return
	}

	if dbAnnounce.Error == nil {
		t.Errorf("Expected error here to be stated.")
		return
	}

	if *dbAnnounce.ScheduledTime != dbAnnounce.AnnounceTime.Add(time.Minute) {
		t.Errorf("Scheduled time not set properly. %v", dbAnnounce)
		return
	}
}

func testAnnounceGivingFailureReason(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Setup
	bodyResponseStruct := struct {
		FailureReason string `bencode:"failure reason"`
	}{
		FailureReason: "You asked for gibberish",
	}

	bodyBytes, err := bencode.Marshal(bodyResponseStruct)
	if err != nil {
		t.Errorf("Could not marshal response.")
		return
	}

	dependencies.trackerServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		n, err := w.Write(bodyBytes)
		if err != nil {
			t.Errorf("Could not send back response.")
			return
		}

		if n < len(bodyBytes) {
			t.Errorf("Not all bytes are sent.")
		}
	})

	dbTorrent, err := setupExistingTorrent(client, dependencies, t)
	if err != nil {
		t.Errorf("Error on test case setup %v", err)
		return
	}

	// Test
	dbAnnounce, err := client.Announce(dbTorrent)
	if err != nil {
		t.Errorf("Not expected error here.")
		return
	}

	if dbAnnounce == nil {
		t.Errorf("Expected to exist here.")
		return
	}

	if dbAnnounce.Error == nil {
		t.Errorf("Expected error to be specified.")
		return
	}

	if *dbAnnounce.ScheduledTime != dbAnnounce.AnnounceTime.Add(time.Minute) {
		t.Errorf("Scheduled time not set properly.")
		return
	}
}

func testAnnounceOK(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Setup
	announceFile, err := os.Open(capturedTrackerResponsePath)
	if err != nil {
		t.Errorf("Could not open announce file.")
		return
	}

	announceBytes, err := io.ReadAll(announceFile)
	if err != nil {
		t.Errorf("Could not read announce file")
		return
	}

	dependencies.trackerServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		n, err := w.Write(announceBytes)
		if err != nil {
			t.Errorf("Failed to write announce response")
			return
		}

		if n < len(announceBytes) {
			t.Errorf("Failed to write complete response")
			return
		}
	})

	dbTorrent, err := setupExistingTorrent(client, dependencies, t)
	if err != nil {
		t.Errorf("Error on test case setup %v", err)
		return
	}

	// Test
	dbAnnounce, err := client.Announce(dbTorrent)
	if err != nil {
		t.Errorf("Not expected error here. %v", err)
		return
	}

	if dbAnnounce == nil {
		t.Errorf("Expected to exist here.")
		return
	}

	if dbAnnounce.Error != nil {
		t.Errorf("Expected error not to be specified.")
		return
	}

	if *dbAnnounce.ScheduledTime != dbAnnounce.AnnounceTime.Add(time.Second*1800) {
		t.Errorf("Scheduled time not set properly.")
		return
	}

	if len(dbAnnounce.RawResponse) == 0 {
		t.Errorf("Raw response not set properly")
		return
	}
}

func TestAnnounce(t *testing.T) {
	testCases := []testCase{
		{
			name:         "5xx error",
			dbSchemaPath: schemaPath,
			testFunction: testAnnounce5xxError,
		},
		{
			name:         "Error giving failure reason",
			dbSchemaPath: schemaPath,
			testFunction: testAnnounceGivingFailureReason,
		},
		{
			name:         "Announce OK",
			dbSchemaPath: schemaPath,
			testFunction: testAnnounceOK,
		},
	}

	for i := range testCases {
		testCase := testCases[i]

		t.Run(testCase.name, func(t *testing.T) {
			runTestCase(&testCase, t)
		})
	}
}
