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

func TestChangePassword(t *testing.T) {
	username := "changePasswordUser"
	email := "changepassword@mail.com"
	password := "oldpassword"
	newPassword := "newpassword"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	err = ChangePassword(testDB, username, password, newPassword)
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	valid, err := CheckPassword(testDB, username, newPassword)
	if err != nil {
		t.Fatalf("CheckPassword failed: %v", err)
	}

	if !valid {
		t.Errorf("expected new password to be valid for user %s", username)
	}

	valid, err = CheckPassword(testDB, username, password)
	if err != nil {
		t.Fatalf("CheckPassword failed: %v", err)
	}

	if valid {
		t.Errorf("expected old password to be invalid for user %s", username)
	}
}

func TestChangeEmail(t *testing.T) {
	username := "changeEmailUser"
	email := "oldemail@mail.com"
	newEmail := "newemail@mail.com"
	password := "password"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	err = ChangeEmail(testDB, username, newEmail)
	if err != nil {
		t.Fatalf("ChangeEmail failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	var updatedEmail string
	err = testDB.QueryRow("SELECT email FROM users WHERE username = ?", username).Scan(&updatedEmail)
	if err != nil {
		t.Fatalf("failed to fetch updated email: %v", err)
	}

	if updatedEmail != newEmail {
		t.Errorf("expected email %s, got %s", newEmail, updatedEmail)
	}

	err = ChangeEmail(testDB, username, newEmail)
	if err == nil {
		t.Error("expected ChangeEmail to fail for duplicate email, but it succeeded")
	}
}

func TestChangeUsername(t *testing.T) {
	username := "oldUsername"
	newUsername := "newUsername"
	email := "username@mail.com"
	password := "password"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	err = ChangeUsername(testDB, username, newUsername)
	if err != nil {
		t.Fatalf("ChangeUsername failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	var updatedUsername string
	err = testDB.QueryRow("SELECT username FROM users WHERE username = ?", newUsername).Scan(&updatedUsername)
	if err != nil {
		t.Fatalf("failed to fetch updated username: %v", err)
	}

	if updatedUsername != newUsername {
		t.Errorf("expected username %s, got %s", newUsername, updatedUsername)
	}

	err = ChangeUsername(testDB, newUsername, newUsername)
	if err == nil {
		t.Error("expected ChangeUsername to fail for duplicate username, but it succeeded")
	}

	err = ChangeUsername(testDB, newUsername, "in valid!")
	if err == nil {
		t.Error("expected ChangeUsername to fail for invalid username, but it succeeded")
	}
}

func TestGeneratePasswordResetCode(t *testing.T) {
	username := "resetUser"
	email := "resetuser@test.com"
	password := "testpassword"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	resetCode, err := GeneratePasswordResetCode(testDB, email)
	if err != nil {
		t.Fatalf("GeneratePasswordResetCode failed: %v", err)
	}

	if resetCode == "" {
		t.Error("Expected a reset code to be generated, got an empty string")
	}

	PrintTableContents(testDB, "users")

	var storedResetCode string
	err = testDB.QueryRow("SELECT reset_code FROM users WHERE email = ?", email).Scan(&storedResetCode)
	if err != nil {
		t.Fatalf("Failed to fetch reset code: %v", err)
	}

	if storedResetCode != resetCode {
		t.Errorf("Expected reset code %s, got %s", resetCode, storedResetCode)
	}
}

func TestResetPassword(t *testing.T) {
	username := "resetPasswordUser"
	email := "resetpassword@test.com"
	password := "oldpassword"
	newPassword := "newpassword"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	resetCode, err := GeneratePasswordResetCode(testDB, email)
	if err != nil {
		t.Fatalf("GeneratePasswordResetCode failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	err = ResetPassword(testDB, email, resetCode, newPassword)
	if err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	valid, err := CheckPassword(testDB, username, newPassword)
	if err != nil {
		t.Fatalf("CheckPassword failed: %v", err)
	}

	PrintTableContents(testDB, "users")

	if !valid {
		t.Errorf("Expected new password to be valid for user %s", username)
	}

	var storedResetCode sql.NullString
	err = testDB.QueryRow("SELECT reset_code FROM users WHERE email = ?", email).Scan(&storedResetCode)
	if err != nil {
		t.Fatalf("Failed to fetch reset_code: %v", err)
	}

	if storedResetCode.Valid {
		t.Errorf("Expected reset code to be cleared, but found: %s", storedResetCode.String)
	}

	PrintTableContents(testDB, "users")
}
