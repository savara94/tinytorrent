package sqlite

import (
	"database/sql"

	"example.com/db"
)

type ClientRepositorySQLite struct {
	SQLiteDB
}

func (r *ClientRepositorySQLite) Create(client *db.Client) error {
	stmt, err := r.db.Prepare("INSERT INTO clients (protocol_id, created) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(client.ProtocolId, client.Created)
	if err != nil {
		return err
	}

	lastInsertedID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Update the client struct with the generated ID
	client.ClientId = int(lastInsertedID)

	return nil
}

func (r *ClientRepositorySQLite) GetLast() (*db.Client, error) {
	row := r.db.QueryRow("SELECT client_id, protocol_id, created FROM clients ORDER BY client_id DESC LIMIT 1")

	var client db.Client
	err := row.Scan(&client.ClientId, &client.ProtocolId, &client.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No rows found
		}
		return nil, err
	}

	return &client, nil
}
