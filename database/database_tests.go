package database

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	testDB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("failed to open in-memory database: %v", err)
	}

	err = setupTables(testDB)
	if err != nil {
		log.Fatalf("failed to setup tables: %v", err)
	}

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func setupTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		username TEXT UNIQUE NOT NULL,
    		password TEXT NOT NULL,
			email TEXT NOT NULL,
			reset_code TEXT DEFAULT NULL,
    		topics_opened INTEGER DEFAULT 0,
    		messages_sent INTEGER DEFAULT 0,
    		creation_date DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS topics (
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		title TEXT UNIQUE NOT NULL,
    		messages INTEGER DEFAULT 0,
			upvotes INTEGER DEFAULT 0,
    		creation_date DATETIME DEFAULT CURRENT_TIMESTAMP,
    		creator_id INTEGER,
    		FOREIGN KEY (creator_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		message TEXT,
    		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			likes INTEGER DEFAULT 0,
    		user_id INTEGER,
    		parent_id INTEGER DEFAULT NULL,
    		topic_id INTEGER NOT NULL,
    		FOREIGN KEY (user_id) REFERENCES users(id),
    		FOREIGN KEY (parent_id) REFERENCES messages(id),
    		FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
		);`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("error executing query: %s\nError: %v", query, err)
			return err
		}
	}
	return nil
}
