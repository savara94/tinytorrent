package sqlite

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	DbPath     string
	SchemaPath string

	db *sql.DB
}

func NewSQLiteDB(dbPath string, schemaPath string) (*SQLiteDB, error) {
	connStr := fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbPath)
	database, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, err
	}

	// Enable foreign key support
	_, err = database.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	// Execute schema SQL
	schemaFile, err := os.Open(schemaPath)
	if err != nil {
		return nil, err
	}

	schemaBytes, err := io.ReadAll(schemaFile)
	if err != nil {
		return nil, err
	}
	schema := string(schemaBytes)

	_, err = database.Exec(schema)
	if err != nil {
		return nil, err
	}

	return &SQLiteDB{dbPath, schemaPath, database}, nil
}
