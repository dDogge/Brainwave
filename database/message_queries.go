package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

func CreateMessageTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS messages (
    			id INTEGER PRIMARY KEY AUTOINCREMENT,
    			message TEXT,
    			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
				likes INTEGER DEFUALT 0,
    			user_id INTEGER,
    			parent_id INTEGER DEFAULT NULL,
    			topic_id INTEGER NOT NULL,
    			FOREIGN KEY (user_id) REFERENCES users(id),
    			FOREIGN KEY (parent_id) REFERENCES messages(id),
    			FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
			);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("error creating message table: ", err)
	}
}

func AddMessage(db *sql.DB, topic, message, username string) error {
	var creatorID int
	var topicID int

	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		log.Printf("error fetching creator_id: %v", err)
		return fmt.Errorf("could not fetch creator_id: %w", err)
	}

	err = db.QueryRow("SELECT id FROM topics WHERE title = ?", topic).Scan(&topicID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("topic not found")
		}
		log.Printf("error fetching creator_id: %v", err)
		return fmt.Errorf("could not fetch creator_id: %w", err)
	}

	stmt, err := db.Prepare("INSERT INTO topics (message, user_id, topic_id) VALUES (?, ?, ?)")
	if err != nil {
		log.Printf("error preparing statement: %v", err)
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(message, creatorID, topicID)
	if err != nil {
		log.Printf("error executing statement: %v", err)
		return fmt.Errorf("could not execute statement: %w", err)
	}

	_, err = db.Exec("UPDATE users SET messages_sent = messages_sent + 1 WHERE id = ?", creatorID)
	if err != nil {
		log.Printf("error incrementing messages_sent for user ID %d: %v", creatorID, err)
		return fmt.Errorf("could not increment messages_sent: %w", err)
	}

	_, err = db.Exec("UPDATE topics SET messages = messages + 1 WHERE id = ?", topicID)
	if err != nil {
		log.Printf("error incrementing messages for topic ID %d: %v", topicID, err)
		return fmt.Errorf("could not increment messages: %w", err)
	}

	log.Println("message added successfully:", message)
	return nil
}

func SetParent(db *sql.DB, parentID, childID int) error {
	var parentTopicID, childTopicID int

	err := db.QueryRow("SELECT topic_id FROM messages WHERE id = ?", parentID).Scan(&parentTopicID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parent message with ID %d not found", parentID)
		}
		log.Printf("error fetching topic_id for parentID %d: %v", parentID, err)
		return fmt.Errorf("could not fetch topic_id for parent message: %w", err)
	}

	err = db.QueryRow("SELECT topic_id FROM messages WHERE id = ?", childID).Scan(&childTopicID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("child message with ID %d not found", childID)
		}
		log.Printf("error fetching topic_id for childID %d: %v", childID, err)
		return fmt.Errorf("could not fetch topic_id for child message: %w", err)
	}

	if parentTopicID != childTopicID {
		return fmt.Errorf("messages are not in the same topic: parentID=%d, childID=%d", parentID, childID)
	}

	_, err = db.Exec("UPDATE messages SET parent_id = ? WHERE id = ?", parentID, childID)
	if err != nil {
		log.Printf("error setting parent_id for user ID %d: %v", childID, err)
		return fmt.Errorf("could not set parent_id: %w", err)
	}

	log.Printf("Parent for message set successfully: parentID=%d, childID=%d", parentID, childID)
	return nil
}
