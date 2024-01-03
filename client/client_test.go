package client

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"example.com/sqlite"
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
