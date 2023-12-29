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

type testCase struct {
	name            string
	temporaryDbPath string
	dbSchemaPath    string
	testFunction    func(sqliteDb *sqlite.SQLiteDB, t *testing.T)
}

var schemaPath = "../sqlite/script.sql"

func runTestCase(testCase *testCase, t *testing.T) {
	// Common test case setup
	sqliteDb, err := sqlite.NewSQLiteDB(testCase.temporaryDbPath, testCase.dbSchemaPath)
	defer os.Remove(testCase.temporaryDbPath)

	if err != nil {
		t.Errorf("Could not create SQLiteDB %v", err)
		return
	}

	testCase.testFunction(sqliteDb, t)
}

func testFirstInitialize(sqliteDb *sqlite.SQLiteDB, t *testing.T) {
	// Setup
	clientRepo := sqlite.ClientRepositorySQLite{SQLiteDB: *sqliteDb}

	dbClient, err := clientRepo.GetLast()
	if err != nil {
		t.Errorf("Could not retrieve last client %v", err)
		return
	}

	if dbClient != nil {
		t.Errorf("Expected client record not to exist!")
	}

	torrentClient := Client{ClientRepo: &clientRepo}

	// Test
	if err := torrentClient.Initialize(); err != nil {
		t.Errorf("Could not initialize torrent client %v", err)
	}

	dbClient, err = clientRepo.GetLast()
	if err != nil {
		t.Errorf("Could not retrieve last client %v", err)
		return
	}

	if dbClient == nil {
		t.Errorf("Expected client record to exist!")
	}

	if !reflect.DeepEqual(dbClient.ProtocolId, torrentClient.Client.ProtocolId) {
		t.Errorf("Not assigned DBClient to TorrentClient, %#v != %#v", dbClient.ProtocolId, torrentClient.Client.ProtocolId)
	}
}

func testInitializeWhenRecordExists(sqliteDb *sqlite.SQLiteDB, t *testing.T) {
	// Setup
	clientRepo := sqlite.ClientRepositorySQLite{SQLiteDB: *sqliteDb}
	torrentClient := Client{ClientRepo: &clientRepo}

	dbClient := db.Client{ProtocolId: torrent.GenerateRandomProtocolId()}
	err := clientRepo.Create(&dbClient)

	if err != nil {
		t.Errorf("Could not create pre-made client %v", err)
		return
	}

	// Test
	err = torrentClient.Initialize()
	if err != nil {
		t.Errorf("Could not initialize torrent client %v", err)
		return
	}

	if !reflect.DeepEqual(dbClient.ProtocolId, torrentClient.Client.ProtocolId) {
		t.Errorf("Did not pick existing client %#v != %#v", dbClient, torrentClient.Client)
	}
}

func TestInitialize(t *testing.T) {
	testCases := []testCase{
		{
			name:            "First time initialize",
			temporaryDbPath: "test1.db",
			dbSchemaPath:    schemaPath,
			testFunction:    testFirstInitialize,
		},
		{
			name:            "Initialization record exists",
			temporaryDbPath: "test2.db",
			dbSchemaPath:    schemaPath,
			testFunction:    testInitializeWhenRecordExists,
		},
	}

	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			runTestCase(&testCases[i], t)
		})
	}
}

func testOpenTorrentWrongFile(sqliteDb *sqlite.SQLiteDB, t *testing.T) {
	// Setup
	torrentRepo := sqlite.TorrentRepositorySQLite{SQLiteDB: *sqliteDb}

	torrentClient := Client{TorrentRepo: &torrentRepo}

	testReader := strings.NewReader("this is not .torrent")

	// Test
	dbTorrent, err := torrentClient.OpenTorrent(testReader, "./")
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

func testDirectoryAlreadyTaken(sqliteDb *sqlite.SQLiteDB, t *testing.T) {
	// Setup
	torrentRepo := sqlite.TorrentRepositorySQLite{SQLiteDB: *sqliteDb}
	torrentClient := Client{TorrentRepo: &torrentRepo}

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
	dbTorrent, err := torrentClient.OpenTorrent(fileReader, downloadPath)
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

func testOpenFewTimes(sqliteDb *sqlite.SQLiteDB, t *testing.T) {
	// Setup
	torrentRepo := sqlite.TorrentRepositorySQLite{SQLiteDB: *sqliteDb}
	torrentClient := Client{TorrentRepo: &torrentRepo}

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

	dbTorrent, err := torrentClient.OpenTorrent(fileReader, tmpDir)
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

	t.Run("Open one more time", func(t *testing.T) {
		_, err := fileReader.Seek(0, io.SeekStart)
		if err != nil {
			t.Errorf("Could not rewind file %v", err)
		}

		newTmpDir, _ := os.MkdirTemp("", "")
		if err != nil {
			t.Errorf("Could not create temp dir %v", err)
			return
		}

		existingTorrent, err := torrentClient.OpenTorrent(fileReader, newTmpDir)
		if err == nil {
			t.Errorf("Expected error here!")
		}

		if err != nil {
			// TODO
			// Check for error type here
		}

		if existingTorrent == nil {
			t.Errorf("Did not expect nil here")
		}

		if existingTorrent.Location != tmpDir {
			t.Errorf("Expected old location %#v", existingTorrent)
		}
	})
}

func TestOpenTorrent(t *testing.T) {
	testCases := []testCase{
		{
			name:            "Wrong file",
			temporaryDbPath: "test3.db",
			dbSchemaPath:    schemaPath,
			testFunction:    testOpenTorrentWrongFile,
		},
		{
			name:            "Directory already taken",
			temporaryDbPath: "test4.db",
			dbSchemaPath:    schemaPath,
			testFunction:    testDirectoryAlreadyTaken,
		},
		{
			name:            "Open valid few times",
			temporaryDbPath: "test5.db",
			dbSchemaPath:    schemaPath,
			testFunction:    testOpenFewTimes,
		},
	}

	for i := range testCases {
		testCase := testCases[i]

		t.Run(testCase.name, func(t *testing.T) {
			runTestCase(&testCase, t)
		})
	}
}

func testAnnounce5xxError(sqliteDb *sqlite.SQLiteDB, t *testing.T) {
	// Setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	announceRepo := sqlite.TrackerAnnounceRepositorySQLite{SQLiteDB: *sqliteDb}
	torrentRepo := sqlite.TorrentRepositorySQLite{SQLiteDB: *sqliteDb}
	clientRepo := sqlite.ClientRepositorySQLite{SQLiteDB: *sqliteDb}

	metaInfo := torrent.MetaInfo{
		Announce: server.URL + "/announce",
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

	client := Client{AnnounceRepo: &announceRepo, TorrentRepo: &torrentRepo, ClientRepo: &clientRepo}
	err = client.Initialize()
	if err != nil {
		t.Errorf("Could not initialize client %v", err)
	}

	dbTorrent, err := client.OpenTorrent(metaInfoBuffer, tmpDir)
	if err != nil {
		t.Errorf("Could not open torrent %v", err)
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

func testAnnounceGivingFailureReason(sqliteDb *sqlite.SQLiteDB, t *testing.T) {
	// Setup
	bodyResponseStruct := struct {
		FailureReason string `bencode:"failure reason"`
	}{"You asked for gibberish"}

	bodyBytes, err := bencode.Marshal(bodyResponseStruct)
	if err != nil {
		t.Errorf("Could not marshal response.")
		return
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		n, err := w.Write(bodyBytes)
		if err != nil {
			t.Errorf("Could not send back response.")
			return
		}

		if n < len(bodyBytes) {
			t.Errorf("Not all bytes are sent.")
		}
	}))
	defer server.Close()

	announceRepo := sqlite.TrackerAnnounceRepositorySQLite{SQLiteDB: *sqliteDb}
	torrentRepo := sqlite.TorrentRepositorySQLite{SQLiteDB: *sqliteDb}
	clientRepo := sqlite.ClientRepositorySQLite{SQLiteDB: *sqliteDb}

	metaInfo := torrent.MetaInfo{
		Announce: server.URL + "/announce",
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

	client := Client{AnnounceRepo: &announceRepo, TorrentRepo: &torrentRepo, ClientRepo: &clientRepo}
	err = client.Initialize()
	if err != nil {
		t.Errorf("Could not initialize client %v", err)
	}

	dbTorrent, err := client.OpenTorrent(metaInfoBuffer, tmpDir)
	if err != nil {
		t.Errorf("Could not open torrent %v", err)
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

func testAnnounceOK(sqliteDb *sqlite.SQLiteDB, t *testing.T) {
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	}))
	defer server.Close()

	announceRepo := sqlite.TrackerAnnounceRepositorySQLite{SQLiteDB: *sqliteDb}
	torrentRepo := sqlite.TorrentRepositorySQLite{SQLiteDB: *sqliteDb}
	peerRepo := sqlite.PeerRepositorySQLite{SQLiteDB: *sqliteDb}
	clientRepo := sqlite.ClientRepositorySQLite{SQLiteDB: *sqliteDb}

	metaInfo := torrent.MetaInfo{
		Announce: server.URL + "/announce",
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

	client := Client{AnnounceRepo: &announceRepo, TorrentRepo: &torrentRepo, ClientRepo: &clientRepo, PeerRepo: &peerRepo}
	err = client.Initialize()
	if err != nil {
		t.Errorf("Could not initialize client %v", err)
	}

	dbTorrent, err := client.OpenTorrent(metaInfoBuffer, tmpDir)
	if err != nil {
		t.Errorf("Could not open torrent %v", err)
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
			name:            "5xx error",
			temporaryDbPath: "test6.db",
			dbSchemaPath:    schemaPath,
			testFunction:    testAnnounce5xxError,
		},
		{
			name:            "Error giving failure reason",
			temporaryDbPath: "test7.db",
			dbSchemaPath:    schemaPath,
			testFunction:    testAnnounceGivingFailureReason,
		},
		{
			name:            "Announce OK",
			temporaryDbPath: "test8.db",
			dbSchemaPath:    schemaPath,
			testFunction:    testAnnounceOK,
		},
	}

	for i := range testCases {
		testCase := testCases[i]

		t.Run(testCase.name, func(t *testing.T) {
			runTestCase(&testCase, t)
		})
	}
}
