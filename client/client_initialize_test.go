package client

import (
	"reflect"
	"testing"

	"example.com/db"
	"example.com/torrent"
)

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
