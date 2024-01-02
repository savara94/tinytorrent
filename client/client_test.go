package client

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"example.com/bencode"
	"example.com/db"
	"example.com/sqlite"
	"example.com/torrent"
)

type testCaseDependencies struct {
	db            *sqlite.SQLiteDB
	trackerServer *httptest.Server
}

type testCase struct {
	name         string
	dbSchemaPath string
	testFunction func(client *Client, dependencies *testCaseDependencies, t *testing.T)
}

var schemaPath = "../sqlite/script.sql"

func runTestCase(testCase *testCase, t *testing.T) {
	temporaryDbDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("Could not create temporary test DB directory %v", err)
		return
	}

	temporaryDbPath := path.Join(temporaryDbDir, "test.db")

	// Common test case setup
	sqliteDb, err := sqlite.NewSQLiteDB(temporaryDbPath, testCase.dbSchemaPath)
	defer os.Remove(temporaryDbPath)

	if err != nil {
		t.Errorf("Could not create SQLiteDB %v", err)
		return
	}

	trackerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer trackerServer.Close()

	dependencies := testCaseDependencies{
		db:            sqliteDb,
		trackerServer: trackerServer,
	}

	client := Client{
		ClientRepo:   &sqlite.ClientRepositorySQLite{SQLiteDB: *sqliteDb},
		TorrentRepo:  &sqlite.TorrentRepositorySQLite{SQLiteDB: *sqliteDb},
		AnnounceRepo: &sqlite.TrackerAnnounceRepositorySQLite{SQLiteDB: *sqliteDb},
		PieceRepo:    &sqlite.PieceRepositorySQLite{SQLiteDB: *sqliteDb},
		PeerRepo:     &sqlite.PeerRepositorySQLite{SQLiteDB: *sqliteDb},
	}

	testCase.testFunction(&client, &dependencies, t)
}

func testFirstInitialize(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Test
	dbClient, err := client.ClientRepo.GetLast()
	if err != nil {
		t.Errorf("Could not retrieve last client %v", err)
		return
	}

	if dbClient != nil {
		t.Errorf("Expected client record not to exist!")
		return
	}

	if err := client.Initialize(); err != nil {
		t.Errorf("Could not initialize client %v", err)
		return
	}

	dbClient, err = client.ClientRepo.GetLast()
	if err != nil {
		t.Errorf("Could not retrieve last client %v", err)
		return
	}

	if dbClient == nil {
		t.Errorf("Expected client record to exist!")
		return
	}

	if !reflect.DeepEqual(dbClient.ProtocolId, client.Client.ProtocolId) {
		t.Errorf("Not assigned DBClient to TorrentClient, %#v != %#v", dbClient.ProtocolId, client.Client.ProtocolId)
	}
}

func testInitializeWhenRecordExists(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Setup
	dbClient := db.Client{ProtocolId: torrent.GenerateRandomProtocolId()}
	err := client.ClientRepo.Create(&dbClient)

	if err != nil {
		t.Errorf("Could not create pre-made client %v", err)
		return
	}

	// Test
	err = client.Initialize()
	if err != nil {
		t.Errorf("Could not initialize torrent client %v", err)
		return
	}

	if !reflect.DeepEqual(dbClient.ProtocolId, client.Client.ProtocolId) {
		t.Errorf("Did not pick existing client %#v != %#v", dbClient, client.Client)
	}
}

func TestInitialize(t *testing.T) {
	testCases := []testCase{
		{
			name:         "First time initialize",
			dbSchemaPath: schemaPath,
			testFunction: testFirstInitialize,
		},
		{
			name:         "Initialization record exists",
			dbSchemaPath: schemaPath,
			testFunction: testInitializeWhenRecordExists,
		},
	}

	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			runTestCase(&testCases[i], t)
		})
	}
}

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
	fileReader, err := os.Open("../torrent/examples/hello_world.torrent")
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
	fileReader, err := os.Open("../torrent/examples/hello_world.torrent")
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

func testAnnounce5xxError(client *Client, dependencies *testCaseDependencies, t *testing.T) {
	// Setup
	dependencies.trackerServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

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
		return
	}

	err = client.Initialize()
	if err != nil {
		t.Errorf("Could not initialize client %v", err)
		return
	}

	dbTorrent, err := client.OpenTorrent(metaInfoBuffer, tmpDir)
	if err != nil {
		t.Errorf("Could not open torrent %v", err)
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
	}{"You asked for gibberish"}

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
	}

	err = client.Initialize()
	if err != nil {
		t.Errorf("Could not initialize client %v", err)
	}

	dbTorrent, err := client.OpenTorrent(metaInfoBuffer, tmpDir)
	if err != nil {
		t.Errorf("Could not open torrent %v", err)
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
	announceFile, err := os.Open("../torrent/examples/ubuntu-22.04.3-desktop-amd64.iso.torrent.compact0.announce")
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
	}

	err = client.Initialize()
	if err != nil {
		t.Errorf("Could not initialize client %v", err)
	}

	dbTorrent, err := client.OpenTorrent(metaInfoBuffer, tmpDir)
	if err != nil {
		t.Errorf("Could not open torrent %v", err)
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
