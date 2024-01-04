package client

import (
	"io"
	"os"
	"testing"

	"example.com/db"
)

func testAnnounceResponseNotParsable(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Setup
	dbTorrent, err := setupExistingTorrent(client, dependencies, t)
	if err != nil {
		t.Errorf("Error setting up torrent %v", err)
		return
	}

	dbAnnounce := db.TrackerAnnounce{
		TorrentId:   dbTorrent.TorrentId,
		RawResponse: []byte("This is not parsable"),
	}

	err = client.AnnounceRepo.Create(&dbAnnounce)
	if err != nil {
		t.Errorf("Did not expect error here %v", err)
		return
	}

	// Test
	peers, err := client.ProcessTrackerAnnounce(&dbAnnounce)
	if err == nil {
		t.Errorf("Expected error here.")
		return
	}

	if len(peers) != 0 {
		t.Errorf("Did not expect peers here.")
		return
	}

	// TODO
	// Check error check type here

}

func testAnnounceContainsAnError(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Setup
	dbTorrent, err := setupExistingTorrent(client, dependencies, t)
	if err != nil {
		t.Errorf("Error setting up torrent %v", err)
		return
	}

	errMsg := "HashInfo invalid"
	announceExampleFile, err := os.Open(capturedTrackerResponsePath)
	if err != nil {
		t.Errorf("Could not open announce response example %v", err)
		return
	}
	defer announceExampleFile.Close()

	announceExampleBytes, err := io.ReadAll(announceExampleFile)
	if err != nil {
		t.Errorf("Could not read announce response example %v", err)
		return
	}

	dbAnnounce := db.TrackerAnnounce{
		TorrentId:   dbTorrent.TorrentId,
		Error:       &errMsg,
		RawResponse: announceExampleBytes,
	}

	err = client.AnnounceRepo.Create(&dbAnnounce)
	if err != nil {
		t.Errorf("Did not expect error here %v", err)
		return
	}

	// Test
	peers, err := client.ProcessTrackerAnnounce(&dbAnnounce)
	if err == nil {
		t.Errorf("Expected error here.")
		return
	}

	if len(peers) != 0 {
		t.Errorf("Did not expect peers here.")
		return
	}

	// TODO
	// Check error check type here
}

func testAnnounceProccessingOK(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Setup
	dbTorrent, err := setupExistingTorrent(client, dependencies, t)
	if err != nil {
		t.Errorf("Error setting up torrent %v", err)
		return
	}

	announceExampleFile, err := os.Open(capturedTrackerResponsePath)
	if err != nil {
		t.Errorf("Could not open announce response example %v", err)
		return
	}
	defer announceExampleFile.Close()

	announceExampleBytes, err := io.ReadAll(announceExampleFile)
	if err != nil {
		t.Errorf("Could not read announce response example %v", err)
		return
	}

	dbAnnounce := db.TrackerAnnounce{
		TorrentId:   dbTorrent.TorrentId,
		RawResponse: announceExampleBytes,
	}

	err = client.AnnounceRepo.Create(&dbAnnounce)
	if err != nil {
		t.Errorf("Did not expect error here %v", err)
		return
	}

	// Test
	peers, err := client.ProcessTrackerAnnounce(&dbAnnounce)
	if err != nil {
		t.Errorf("Did not expected error here.")
		return
	}

	if len(peers) == 0 {
		t.Errorf("Expected more peers here.")
		return
	}
}

func TestProcessTrackerAnnounce(t *testing.T) {
	testCases := []testCase{
		{
			name:         "Response not parsable",
			dbSchemaPath: schemaPath,
			testFunction: testAnnounceResponseNotParsable,
		},
		{
			name:         "Announce contains error",
			dbSchemaPath: schemaPath,
			testFunction: testAnnounceContainsAnError,
		},
		{
			name:         "Processing OK",
			dbSchemaPath: schemaPath,
			testFunction: testAnnounceProccessingOK,
		},
	}

	for i := range testCases {
		testCase := testCases[i]

		t.Run(testCase.name, func(t *testing.T) {
			runTestCase(&testCase, t)
		})
	}
}
