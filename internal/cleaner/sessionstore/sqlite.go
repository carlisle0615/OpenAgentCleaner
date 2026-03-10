package sessionstore

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func OpenSQLiteDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys for %s: %w", path, err)
	}
	return db, nil
}
