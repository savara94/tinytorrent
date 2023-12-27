package client

import (
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

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

func TestAnnounce(t *testing.T) {
	testCases := []testCase{
		{
			name:            "5xx error",
			temporaryDbPath: "test6.db",
			dbSchemaPath:    schemaPath,
			testFunction:    func(sqliteDb *sqlite.SQLiteDB, t *testing.T) {},
		},
		{
			name:            "Error giving failure reason",
			temporaryDbPath: "test7.db",
			dbSchemaPath:    schemaPath,
			testFunction:    func(sqliteDb *sqlite.SQLiteDB, t *testing.T) {},
		},
		{
			name:            "Announce OK",
			temporaryDbPath: "test8.db",
			dbSchemaPath:    schemaPath,
			testFunction:    func(sqliteDb *sqlite.SQLiteDB, t *testing.T) {},
		},
	}

	for i := range testCases {
		testCase := testCases[i]

		t.Run(testCase.name, func(t *testing.T) {
			runTestCase(&testCase, t)
		})
	}
}
