package client

import (
	"os"
	"reflect"
	"testing"

	"example.com/db"
	"example.com/sqlite"
	"example.com/torrent"
)

func TestInitialize(t *testing.T) {
	t.Run("First time", func(t *testing.T) {
		sqliteDb, err := sqlite.NewSQLiteDB("test1.db", "../sqlite/script.sql")
		defer os.Remove("test1.db")

		if err != nil {
			t.Errorf("Could not create SQLiteDB %v", err)
			return
		}

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
	})

	t.Run("Client record exists", func(t *testing.T) {
		sqliteDb, err := sqlite.NewSQLiteDB("test2.db", "../sqlite/script.sql")
		defer os.Remove("test2.db")

		if err != nil {
			t.Errorf("Could not create SQLiteDB %v", err)
			return
		}

		clientRepo := sqlite.ClientRepositorySQLite{SQLiteDB: *sqliteDb}
		torrentClient := Client{ClientRepo: &clientRepo}

		dbClient := db.Client{ProtocolId: torrent.GenerateRandomProtocolId()}
		err = clientRepo.Create(&dbClient)
		if err != nil {
			t.Errorf("Could not create pre-made client %v", err)
			return
		}

		err = torrentClient.Initialize()
		if err != nil {
			t.Errorf("Could not initialize torrent client %v", err)
			return
		}

		if !reflect.DeepEqual(dbClient.ProtocolId, torrentClient.Client.ProtocolId) {
			t.Errorf("Did not pick existing client %#v != %#v", dbClient, torrentClient.Client)
		}
	})

}
