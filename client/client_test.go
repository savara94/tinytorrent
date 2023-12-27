package client

import (
	"os"
	"reflect"
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
	schemaPath := "../sqlite/script.sql"

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
