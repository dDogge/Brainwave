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
		t.Error("expected a reset code to be generated, got an empty string")
	}

	PrintTableContents(testDB, "users")

	var storedResetCode string
	err = testDB.QueryRow("SELECT reset_code FROM users WHERE email = ?", email).Scan(&storedResetCode)
	if err != nil {
		t.Fatalf("failed to fetch reset code: %v", err)
	}

	if storedResetCode != resetCode {
		t.Errorf("expected reset code %s, got %s", resetCode, storedResetCode)
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
		t.Errorf("expected new password to be valid for user %s", username)
	}

	var storedResetCode sql.NullString
	err = testDB.QueryRow("SELECT reset_code FROM users WHERE email = ?", email).Scan(&storedResetCode)
	if err != nil {
		t.Fatalf("failed to fetch reset_code: %v", err)
	}

	if storedResetCode.Valid {
		t.Errorf("expected reset code to be cleared, but found: %s", storedResetCode.String)
	}

	PrintTableContents(testDB, "users")
}

func TestGetAllUsers(t *testing.T) {
	users := []struct {
		username string
		email    string
		password string
	}{
		{"user1", "user1@test.com", "password1"},
		{"user2", "user2@test.com", "password2"},
		{"user3", "user3@test.com", "password3"},
	}

	for _, user := range users {
		err := AddUser(testDB, user.username, user.email, user.password)
		if err != nil {
			t.Fatalf("AddUser failed for %s: %v", user.username, err)
		}
	}

	PrintTableContents(testDB, "users")

	allUsers, err := GetAllUsers(testDB)
	if err != nil {
		t.Fatalf("GetAllUsers failed: %v", err)
	}

	if len(allUsers) != len(users) {
		t.Errorf("expected %d users, got %d", len(users), len(allUsers))
	}

	for _, user := range users {
		found := false
		for _, u := range allUsers {
			if u["username"] == user.username {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("user %s not found in result", user.username)
		}
	}
}

func TestAddTopic(t *testing.T) {
	username := "topicUser"
	email := "topicuser@test.com"
	password := "password"
	topicTitle := "Test Topic"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = AddTopic(testDB, topicTitle, username)
	if err != nil {
		t.Fatalf("AddTopic failed: %v", err)
	}

	PrintTableContents(testDB, "topics")

	var storedTitle string
	err = testDB.QueryRow("SELECT title FROM topics WHERE title = ?", topicTitle).Scan(&storedTitle)
	if err != nil {
		t.Fatalf("failed to fetch topic: %v", err)
	}

	if storedTitle != topicTitle {
		t.Errorf("expected topic title %s, got %s", topicTitle, storedTitle)
	}

	var topicsOpened int
	err = testDB.QueryRow("SELECT topics_opened FROM users WHERE username = ?", username).Scan(&topicsOpened)
	if err != nil {
		t.Fatalf("failed to fetch topics_opened for user: %v", err)
	}

	if topicsOpened != 1 {
		t.Errorf("expected topics_opened to be 1, got %d", topicsOpened)
	}

	err = AddTopic(testDB, topicTitle, username)
	if err == nil {
		t.Error("expected AddTopic to fail for duplicate title, but it succeeded")
	}
}

func TestRemoveTopic(t *testing.T) {
	username := "removeTopicUser"
	email := "removetopicuser@test.com"
	password := "password"
	topicTitle := "Removable Topic"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = AddTopic(testDB, topicTitle, username)
	if err != nil {
		t.Fatalf("AddTopic failed: %v", err)
	}

	PrintTableContents(testDB, "topics")

	err = RemoveTopic(testDB, topicTitle)
	if err != nil {
		t.Fatalf("RemoveTopic failed: %v", err)
	}

	PrintTableContents(testDB, "topics")

	var storedTitle string
	err = testDB.QueryRow("SELECT title FROM topics WHERE title = ?", topicTitle).Scan(&storedTitle)
	if err == nil {
		t.Errorf("expected topic %s to be removed, but it still exists", topicTitle)
	} else if err != sql.ErrNoRows {
		t.Fatalf("unexpected error while checking topic removal: %v", err)
	}

	err = RemoveTopic(testDB, "Nonexistent Topic")
	if err == nil {
		t.Error("expected RemoveTopic to fail for nonexistent topic, but it succeeded")
	}
}
