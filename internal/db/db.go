package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func dbPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "track-", "track.db")
}

func Open() (*DB, error) {
	dir := filepath.Dir(dbPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", dbPath())
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

func MustOpen() *DB {
	db, err := Open()
	if err != nil {
		panic("failed to open database: " + err.Error())
	}
	return db
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS packages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tracking_number TEXT UNIQUE NOT NULL,
			nickname TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'Unknown',
			status_category TEXT NOT NULL DEFAULT 'unknown',
			origin_city TEXT DEFAULT '',
			origin_state TEXT DEFAULT '',
			dest_city TEXT DEFAULT '',
			dest_state TEXT DEFAULT '',
			expected_delivery TEXT DEFAULT '',
			last_updated TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tracking_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tracking_number TEXT NOT NULL,
			event_date TEXT NOT NULL,
			event_description TEXT NOT NULL,
			city TEXT DEFAULT '',
			state TEXT DEFAULT '',
			zip TEXT DEFAULT '',
			country TEXT DEFAULT '',
			FOREIGN KEY (tracking_number) REFERENCES packages(tracking_number) ON DELETE CASCADE
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_events_unique
			ON tracking_events(tracking_number, event_date, event_description)`,
		`CREATE INDEX IF NOT EXISTS idx_events_tracking
			ON tracking_events(tracking_number, event_date DESC)`,
	}

	for _, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			return err
		}
	}

	// Enable foreign keys
	_, err := db.conn.Exec("PRAGMA foreign_keys = ON")
	return err
}
