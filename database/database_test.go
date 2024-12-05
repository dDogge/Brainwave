package database

import (
	"database/sql"
	"fmt"
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
		t.Error("expected RemoveUser to fail for nonexistent user, but it succeeded")
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
			t.Fatalf("AddUser failed: %v", err)
		}
	}

	PrintTableContents(testDB, "users")

	allUsers, err := GetAllUsers(testDB)
	if err != nil {
		t.Fatalf("GetAllUsers failed: %v", err)
	}

	if len(allUsers) < len(users) {
		t.Errorf("expected at least %d users, got %d", len(users), len(allUsers))
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
	PrintTableContents(testDB, "users")

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

func TestUpVoteAndDownVoteTopic(t *testing.T) {
	username := "voteUser"
	email := "voteuser@test.com"
	password := "password"
	topicTitle := "Votable Topic"

	err := AddUser(testDB, username, email, password)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = AddTopic(testDB, topicTitle, username)
	if err != nil {
		t.Fatalf("AddTopic failed: %v", err)
	}

	PrintTableContents(testDB, "topics")

	var initialUpvotes int
	err = testDB.QueryRow("SELECT upvotes FROM topics WHERE title = ?", topicTitle).Scan(&initialUpvotes)
	if err != nil {
		t.Fatalf("failed to fetch initial upvotes for topic: %v", err)
	}

	err = UpVoteTopic(testDB, topicTitle, username)
	if err != nil {
		t.Fatalf("UpVoteTopic failed: %v", err)
	}

	PrintTableContents(testDB, "topics")

	var updatedUpvotes int
	err = testDB.QueryRow("SELECT upvotes FROM topics WHERE title = ?", topicTitle).Scan(&updatedUpvotes)
	if err != nil {
		t.Fatalf("expected upvotes to be %d, got %d", initialUpvotes+1, updatedUpvotes)
	}

	err = DownVoteTopic(testDB, topicTitle, username)
	if err != nil {
		t.Fatalf("DownVoteTopic failed: %v", err)
	}

	PrintTableContents(testDB, "topics")

	err = testDB.QueryRow("SELECT upvotes FROM topics WHERE title = ?", topicTitle).Scan(&updatedUpvotes)
	if err != nil {
		t.Fatalf("failed to fetch updated upvotes for topic: %v", err)
	}

	if updatedUpvotes != initialUpvotes {
		t.Errorf("expected upvotes to be %d, got %d", initialUpvotes, updatedUpvotes)
	}
}

func TestGetAllTopics(t *testing.T) {
	topics := []struct {
		title    string
		username string
	}{
		{"Topic 1", "userTopic1"},
		{"Topic 2", "userTopic2"},
		{"Topic 3", "userTopic3"},
	}

	for _, topic := range topics {
		err := AddUser(testDB, topic.username, fmt.Sprintf("%s@test.com", topic.username), "password")
		if err != nil {
			t.Fatalf("AddUser failed: %v", err)
		}

		err = AddTopic(testDB, topic.title, topic.username)
		if err != nil {
			t.Fatalf("AddTopic failed: %v", err)
		}
	}

	PrintTableContents(testDB, "topics")

	allTopics, err := GetAllTopics(testDB)
	if err != nil {
		t.Fatalf("GetAllTopics failed: %v", err)
	}

	if len(allTopics) < len(topics) {
		t.Errorf("expected at least %d topics, got %d", len(topics), len(allTopics))
	}

	for _, topic := range topics {
		found := false
		for _, t := range allTopics {
			if t["title"] == topic.title {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("topic %s not found in result", topic.title)
		}
	}
}

func TestGetTopicsByTitle(t *testing.T) {
	username := "userTopicByTitle"
	topicTitle := "Unique Topic"

	err := AddUser(testDB, username, fmt.Sprintf("%s@test.com", username), "password")
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = AddTopic(testDB, topicTitle, username)
	if err != nil {
		t.Fatalf("AddTopic failed: %v", err)
	}

	PrintTableContents(testDB, "topics")

	topic, err := GetTopicByTitle(testDB, topicTitle)
	if err != nil {
		t.Fatalf("GetTopicByTitle failed: %v", err)
	}

	if topic["title"] != topicTitle {
		t.Errorf("expected title %s, got %s", topicTitle, topic["title"])
	}

	if topic["creator_id"] == nil {
		t.Errorf("expected a creator_id for topic %s, but got nil", topicTitle)
	}

	_, err = GetTopicByTitle(testDB, "Nonexsitent Topic")
	if err == nil {
		t.Errorf("expected error for nonexistent topic, but got nil")
	}
}

func TestCountTopics(t *testing.T) {
	initialCount, err := CountTopics(testDB)
	if err != nil {
		t.Fatalf("CountTopics failed initially: %v", err)
	}

	username := "countUser"
	topicTitle := "Countable Topic"

	err = AddUser(testDB, username, fmt.Sprintf("%s@test.com", username), "password")
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = AddTopic(testDB, topicTitle, username)
	if err != nil {
		t.Fatalf("AddTopic failed: %v", err)
	}

	PrintTableContents(testDB, "topics")

	currentCount, err := CountTopics(testDB)
	if err != nil {
		t.Fatalf("CountTopics failed after adding a topic: %v", err)
	}

	if currentCount != initialCount+1 {
		t.Errorf("expected topic count to be %d, but got %d", initialCount+1, currentCount)
	}

	err = RemoveTopic(testDB, topicTitle)
	if err != nil {
		t.Fatalf("Removetopic failed: %v", err)
	}

	PrintTableContents(testDB, "topics")

	finalCount, err := CountTopics(testDB)
	if err != nil {
		t.Fatalf("CountTopics failed after removing a topic: %v", err)
	}

	if finalCount != initialCount {
		t.Errorf("expected topic count to return to %d, but got %d", initialCount, finalCount)
	}
}

func TestAddMessage(t *testing.T) {
	username := "messageUser"
	topic := "messageTopic"
	message := "This is a test message."

	err := AddUser(testDB, username, fmt.Sprintf("%s@test.com", username), "password")
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = AddTopic(testDB, topic, username)
	if err != nil {
		t.Fatalf("AddTopic failed: %v", err)
	}

	PrintTableContents(testDB, "users")
	PrintTableContents(testDB, "topics")

	err = AddMessage(testDB, topic, message, username)
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	PrintTableContents(testDB, "messages")

	var storedMessage string
	var storedUserID, storedTopicID int
	err = testDB.QueryRow("SELECT message, user_id, topic_id FROM messages WHERE message = ?", message).
		Scan(&storedMessage, &storedUserID, &storedTopicID)
	if err != nil {
		t.Fatalf("failed to fetch message: %v", err)
	}

	if storedMessage != message {
		t.Errorf("expected message %s, got %s", message, storedMessage)
	}

	var messagesSent, messagesInTopic int
	err = testDB.QueryRow("SELECT messages_sent FROM users WHERE username = ?", username).Scan(&messagesSent)
	if err != nil {
		t.Fatalf("failed to fetch messages_sent for user: %v", err)
	}

	if messagesSent != 1 {
		t.Errorf("expected messages_sent to be 1, got %d", messagesSent)
	}

	err = testDB.QueryRow("SELECT messages FROM topics WHERE title = ?", topic).Scan(&messagesInTopic)
	if err != nil {
		t.Fatalf("failed to fetch messages count for topic: %v", err)
	}

	if messagesInTopic != 1 {
		t.Errorf("expected messages in topic to be 1, got %d", messagesInTopic)
	}

	PrintTableContents(testDB, "users")
}

func TestSetParent(t *testing.T) {
	username := "parentUser"
	topic := "parentTopic"
	parentMessage := "This is the parent message."
	childMessage := "This is the child message."

	err := AddUser(testDB, username, fmt.Sprintf("%s@test.com", username), "password")
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = AddTopic(testDB, topic, username)
	if err != nil {
		t.Fatalf("AddTopic failed: %v", err)
	}

	err = AddMessage(testDB, topic, parentMessage, username)
	if err != nil {
		t.Fatalf("AddMessage failed for parentMessage: %v", err)
	}

	err = AddMessage(testDB, topic, childMessage, username)
	if err != nil {
		t.Fatalf("AddMessage failed for childMessage: %v", err)
	}

	PrintTableContents(testDB, "messages")

	var parentID, childID int
	err = testDB.QueryRow("SELECT id FROM messages WHERE message = ?", parentMessage).Scan(&parentID)
	if err != nil {
		t.Fatalf("failed to fetch parentID: %v", err)
	}

	err = testDB.QueryRow("SELECT id FROM messages WHERE message = ?", childMessage).Scan(&childID)
	if err != nil {
		t.Fatalf("failed to fetch childID: %v", err)
	}

	err = SetParent(testDB, parentID, childID)
	if err != nil {
		t.Fatalf("SetParent failed: %v", err)
	}

	PrintTableContents(testDB, "messages")

	var storedParentID int
	err = testDB.QueryRow("SELECT parent_id FROM messages WHERE id = ?", childID).Scan(&storedParentID)
	if err != nil {
		t.Fatalf("failed to fetch parent_id for child message: %v", err)
	}

	if storedParentID != parentID {
		t.Errorf("expected parent_id %d, got %d", parentID, storedParentID)
	}

	err = SetParent(testDB, parentID, childID+1)
	if err == nil {
		t.Errorf("expected error for mismatched topics, but got nil")
	}
}

func TestGetMessagesByTopic(t *testing.T) {
	topicTitle := "testTopicGet"
	username := "testUser"
	message1 := "This is the first test message"
	message2 := "This is the second test message"

	err := AddUser(testDB, username, "testuser@mail.com", "password")
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = AddTopic(testDB, topicTitle, username)
	if err != nil {
		t.Fatalf("AddTopic failed: %v", err)
	}

	var topicID int
	err = testDB.QueryRow("SELECT id FROM topics WHERE title = ?", topicTitle).Scan(&topicID)
	if err != nil {
		t.Fatalf("failed to fetch topic ID: %v", err)
	}

	err = AddMessage(testDB, topicTitle, message1, username)
	if err != nil {
		t.Fatalf("AddMessage failed %v", err)
	}

	err = AddMessage(testDB, topicTitle, message2, username)
	if err != nil {
		t.Fatalf("AddMessage failed %v", err)
	}

	PrintTableContents(testDB, "messages")

	messages, err := GetMessagesByTopic(testDB, topicID)
	if err != nil {
		t.Fatalf("GetMessagesByTopic failed: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}

	foundMessage1 := false
	foundMessage2 := false
	for _, msg := range messages {
		if msg["message"] == message1 {
			foundMessage1 = true
		} else if msg["message"] == message2 {
			foundMessage2 = true
		}
	}

	if !foundMessage1 {
		t.Errorf("message 1 not found in results")
	}

	if !foundMessage2 {
		t.Errorf("message 2 not found in results")
	}
}

func TestLikeMessage(t *testing.T) {
	topicTitle := "likeTopic"
	username := "likeUser"
	message := "This is a message to like"

	err := AddUser(testDB, username, "likeuser@mail.com", "password")
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = AddTopic(testDB, topicTitle, username)
	if err != nil {
		t.Fatalf("AddTopic failed: %v", err)
	}

	err = AddMessage(testDB, topicTitle, message, username)
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	PrintTableContents(testDB, "messages")

	var messageID int
	err = testDB.QueryRow("SELECT id FROM messages WHERE message = ?", message).Scan(&messageID)
	if err != nil {
		t.Fatalf("failed to fetch message ID: %v", err)
	}

	err = LikeMessage(testDB, messageID)
	if err != nil {
		t.Fatalf("LikeMessage failed: %v", err)
	}

	PrintTableContents(testDB, "messages")

	var likes int
	err = testDB.QueryRow("SELECT likes FROM messages WHERE id = ?", messageID).Scan(&likes)
	if err != nil {
		t.Fatalf("failed to fetch likes: %v", err)
	}

	if likes != 1 {
		t.Errorf("expected 1 like, got %d", likes)
	}
}

func TestDislikeMessage(t *testing.T) {
	topicTitle := "dislikeTopic"
	username := "dislikeUser"
	message := "This is a message to dislike"

	err := AddUser(testDB, username, "dislikeuser@mail.com", "password")
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = AddTopic(testDB, topicTitle, username)
	if err != nil {
		t.Fatalf("AddTopic failed: %v", err)
	}

	err = AddMessage(testDB, topicTitle, message, username)
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	PrintTableContents(testDB, "messages")

	var messageID int
	err = testDB.QueryRow("SELECT id FROM messages WHERE message = ?", message).Scan(&messageID)
	if err != nil {
		t.Fatalf("failed to fetch message ID: %v", err)
	}

	err = LikeMessage(testDB, messageID)
	if err != nil {
		t.Fatalf("LikeMessage failed: %v", err)
	}

	PrintTableContents(testDB, "messages")

	err = DislikeMessage(testDB, messageID)
	if err != nil {
		t.Fatalf("DislikeMessage failed: %v", err)
	}

	PrintTableContents(testDB, "messages")

	var likes int
	err = testDB.QueryRow("SELECT likes FROM messages WHERE id = ?", messageID).Scan(&likes)
	if err != nil {
		t.Fatalf("failed to fetch likes: %v", err)
	}

	if likes != 0 {
		t.Errorf("expected 1 like, got %d", likes)
	}
}
