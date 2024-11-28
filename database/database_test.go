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

	testDB, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatalf("failed to open in-memory database: %v", err)
	}

	err = SetupTables(testDB)
	if err != nil {
		log.Fatalf("failed to setup tables: %v", err)
	}

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func SetupTables(db *sql.DB) error {
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

func PrintTableContents(db *sql.DB, tableName string) {
	query := "SELECT * FROM " + tableName
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("error querying table %s: %v", tableName, err)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Printf("error fetching columns for table %s: %v", tableName, err)
		return
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	log.Printf("contents of table %s:", tableName)
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			log.Printf("error scanning row: %v", err)
			return
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			rowData[col] = values[i]
		}
		log.Println(rowData)
	}
}

func TestAddUser(t *testing.T) {
	username := "testuser"
	email := "test@mail.com"
	password := "password"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	var storedUsername string
	err = testDB.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&storedUsername)
	if err != nil {
		t.Fatalf("failed to fetch user: %v", err)
	}

	if storedUsername != username {
		t.Errorf("expected username %s, got %s", username, storedUsername)
	}
}

func TestRemoveUser(t *testing.T) {
	username := "removalUser"
	email := "removal@mail.com"
	password := "removal"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	var storedUsername string
	err = testDB.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&storedUsername)
	if err != nil {
		t.Fatalf("failed to fetch user: %v", err)
	}

	if storedUsername != username {
		t.Errorf("expected username %s, got %s", username, storedUsername)
	}

	err = RemoveUser(testDB, username)
	if err != nil {
		t.Fatalf("RemoveUser failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	var checkUsername string
	err = testDB.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&checkUsername)
	if err == nil {
		t.Errorf("expected user %s to be removed, but it still exists", username)
	} else if err != sql.ErrNoRows {
		t.Fatalf("unexpected error while checking for removed user: %v", err)
	}

	err = RemoveUser(testDB, "nonexistentUser")
	if err == nil {
		t.Error("Expected RemoveUser to fail for nonexistent user, but it succeeded")
	}
}

func TestCheckPassword(t *testing.T) {
	username := "passwordUser"
	email := "password@mail.com"
	password := "securepassword"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	valid, err := CheckPassword(testDB, username, password)
	if err != nil {
		t.Fatalf("CheckPassword failed: %v", err)
	}

	if !valid {
		t.Errorf("expected password to be valid for user %s", username)
	}

	valid, err = CheckPassword(testDB, username, "wrongpassword")
	if err != nil {
		t.Fatalf("CheckPassword failed: %v", err)
	}

	if valid {
		t.Errorf("expected password to be invalid for user %s", username)
	}
}
